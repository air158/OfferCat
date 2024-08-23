package interview

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"net/url"
	"offercat/v0/internal/lib"
	"offercat/v0/internal/resume"
	pdf_analyser "offercat/v0/internal/thirdparty/pdf-analyser"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func saveToDatabase(db *gorm.DB, question Question) error {
	return db.Create(&question).Error
}

func ProxyLLM(targetHost string, db *gorm.DB) gin.HandlerFunc {
	log.Println("ProxyLLM started")
	return func(c *gin.Context) {
		// 从URL路径中获取任务标识
		task := c.Param("task")
		log.Println("Task:", task) // 添加日志
		var proxyPath string

		// 根据不同任务标识修改目标路径或处理逻辑
		if task == "question" {
			proxyPath = "/stream_questions" // 修改代理路径
		}
		if task == "llm-answer" {
			proxyPath = "/stream_answer"
		}
		if task == "result" {
			proxyPath = "/stream_result"
		}

		// 解析目标URL并设置代理
		targetURL, err := url.Parse(targetHost)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "解析目标URL失败", err)
			return
		}

		// 读取原始请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "读取原始请求体失败", err)
			return
		}

		// 解析请求体为JSON
		var requestData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
			lib.Err(c, http.StatusBadRequest, "解析请求体为JSON失败", err)
			return
		}

		// 添加 "chat_key" 字段
		requestData["chat_key"] = fmt.Sprintf("%s:%s", viper.GetString("spark.apiKey"), viper.GetString("spark.apiSecret"))
		log.Println("Modified request data:", requestData) // 添加日志

		if task == "question" {
			var preset Preset
			err = db.Where("user_id=?", lib.Uid(c)).First(&preset).Error

			if err != nil {
				lib.Err(c, http.StatusInternalServerError, "查询该用户的预设信息失败", err)
				return
			}
			requestData["job_title"] = preset.JobTitle
			requestData["job_description"] = preset.JobDescription
			var r resume.Resume
			// resume_id是最后上传的简历id
			db.Where("id=? and user_id=?", preset.ResumeID, lib.Uid(c)).First(&r)
			if r.Content == "" {
				path := r.FilePath
				// 从MinIO下载简历文件
				stringFromPDF := pdf_analyser.GetStringFromPDF(c, path)
				r.Content = stringFromPDF
				db.Save(&r)
			}
			requestData["resume_text"] = r.Content
		}
		if task == "llm-answer" {
			var preset Preset
			err := db.Where("user_id=?", lib.Uid(c)).First(&preset).Error
			if err != nil {
				lib.Err(c, http.StatusInternalServerError, "尚未设置预设job_title或未传job_title字段", err)
				return
			}
			requestData["job_title"] = preset.JobTitle
			requestData["job_description"] = preset.JobDescription

			var question Question
			err = db.Where("branch_id=? and interview_id=? and user_id=?", requestData["question_branch_id"], requestData["interview_id"], lib.Uid(c)).First(&question).Error
			if err != nil {
				lib.Err(c, http.StatusInternalServerError, "查询问题失败", err)
				return
			}
			requestData["question"] = question.Content
		}

		if task == "result" {
			requestData["prompt_text"], err = FormatInterviewResult(db, uint(requestData["interview_id"].(float64)))
			if err != nil {
				lib.Err(c, http.StatusInternalServerError, "格式化面试结果失败", err)
				return
			}

			var preset Preset
			err = db.Where("user_id=?", lib.Uid(c)).First(&preset).Error

			if err != nil {
				lib.Err(c, http.StatusInternalServerError, "查询该用户的预设信息失败", err)
				return
			}
			requestData["job_title"] = preset.JobTitle
		}
		promptText := requestData["prompt_text"]

		// 将加工后的JSON重新编码为字节数组
		modifiedBodyBytes, err := json.Marshal(requestData)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "编码JSON失败", err)
			return
		}

		// 创建新的请求
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL.ResolveReference(&url.URL{Path: proxyPath}).String(), bytes.NewBuffer(modifiedBodyBytes))
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "创建代理请求失败", err)
			return
		}
		// 复制请求头
		proxyReq.Header = c.Request.Header.Clone()
		proxyReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}

		resp, err := client.Do(proxyReq)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "执行代理请求失败", err)
			return
		}
		defer resp.Body.Close()

		// 流式传输，确保数据按照接收到的顺序逐步发送给客户端
		scanner := bufio.NewScanner(resp.Body)
		var completeDataB strings.Builder

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		for scanner.Scan() {
			line := scanner.Text()

			// 将原始数据直接写入客户端
			_, err := c.Writer.Write([]byte(line + "\n"))
			if err != nil {
				log.Printf("Failed to write data to client: %v", err)
				break
			}
			c.Writer.Flush()

			// 去除前缀和多余空格，仅保留在 completeDataB 中
			trimmed := strings.TrimPrefix(line, "data:")
			trimmed = strings.TrimSpace(trimmed)
			completeDataB.WriteString(trimmed)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading response body: %v", err)
		}

		completeData := completeDataB.String()
		log.Println("Complete data:", completeData) // 添加日志

		if task == "question" {
			// 将completeData按分隔符分割成多个问题
			questions := strings.Split(completeData, `¥¥`)

			// 去掉最后一个问题中的 "[DONE]"
			if len(questions) > 0 {
				questions[len(questions)-1] = strings.TrimSuffix(questions[len(questions)-1], "[DONE]")
			}

			// 尝试查找是否有现有的记录
			err := db.First(&Question{}, "interview_id = ? and user_id=?", requestData["interview_id"], lib.Uid(c)).Error

			if err != nil {
				// 没有找到记录，插入所有问题
				for i, content := range questions {
					db.Create(&Question{
						InterviewID: (uint)(requestData["interview_id"].(float64)),
						UserID:      uint(lib.Uid(c)),
						BranchID:    uint(i + 1),
						Content:     content,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					})
				}
			} else {
				// 找到记录，更新所有问题
				for i, content := range questions {
					db.Model(&Question{}).
						Where("interview_id = ? and user_id=? and branch_id=?", requestData["interview_id"], lib.Uid(c), i+1).
						Updates(map[string]interface{}{"content": content, "updated_at": time.Now()})
				}
			}

			// 更新预设信息中的问题长度
			var preset Preset
			if err = db.Where("user_id=?", lib.Uid(c)).First(&preset).Error; err != nil {
				log.Println("Failed to retrieve preset from database:", err)
			} else {
				preset.QuestionLength = int(requestData["ques_len"].(float64))
				if err = db.Save(&preset).Error; err != nil {
					log.Println("Failed to save question length to database:", err)
				} else {
					log.Println("Question length saved to database")
				}
			}
		}

		if task == "llm-answer" {
			questionBranchID := (uint)(requestData["question_branch_id"].(float64))
			interviewID := (uint)(requestData["interview_id"].(float64))
			userID := uint(lib.Uid(c))

			// 尝试在数据库中查找对应的记录
			var userAnswer Answer
			err := db.Where("interview_id = ? AND question_branch_id=? AND user_id = ?", interviewID, questionBranchID, userID).First(&userAnswer).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 如果记录不存在，则创建新的记录
					userAnswer = Answer{
						InterviewID:      interviewID,
						QuestionBranchID: questionBranchID,
						UserID:           userID,
						LLMAnswer:        completeData,
						CreatedAt:        time.Now(),
						UpdatedAt:        time.Now(),
					}
					if err := db.Create(&userAnswer).Error; err != nil {
						// 处理创建时的错误
						log.Printf("Failed to create new question record: %v", err)
					}
				} else {
					// 处理查询时的其他错误
					log.Printf("Failed to query question record: %v", err)
				}
			} else {
				// 如果记录存在，则更新记录
				userAnswer.LLMAnswer = completeData
				userAnswer.UpdatedAt = time.Now()
				if err := db.Save(&userAnswer).Error; err != nil {
					// 处理更新时的错误
					log.Printf("Failed to update question record: %v", err)
				}
			}
		}
		if task == "result" {
			id := uint(requestData["interview_id"].(float64))
			userID := uint(lib.Uid(c))
			finalSummary := completeData
			var interview Interview
			// 尝试在数据库中查找对应的记录,TODO：这里有注意带&符号
			err := db.Where("id = ? AND user_id=? ", id, userID).First(&interview).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 如果记录不存在，则创建新的记录
					interview = Interview{
						ID:           id,
						UserID:       userID,
						FinalSummary: finalSummary,
						Dialog:       promptText.(string),
					}
					if err := db.Create(&interview).Error; err != nil {
						// 处理创建时的错误
						log.Printf("Failed to create new question record: %v", err)
					}
				} else {
					// 处理查询时的其他错误
					log.Printf("Failed to query question record: %v", err)
				}
			} else {
				// 如果记录存在，则更新记录
				interview.FinalSummary = completeData
				interview.Dialog = promptText.(string)
				if err := db.Save(&interview).Error; err != nil {
					// 处理更新时的错误
					log.Printf("Failed to update question record: %v", err)
				}
			}

		}

	}
}

//func main() {
//	// 初始化数据库连接
//	db, err := gorm.Open(db.DB.Open("test.db"), &gorm.Config{})
//	if err != nil {
//		log.Fatal("Failed to connect database:", err)
//	}
//
//	// 自动迁移数据库结构
//	db.AutoMigrate(&Question{})
//
//	r := gin.Default()
//
//	// 转发到Flask服务，同时加工请求的JSON数据和保存响应到数据库
//	r.Any("/api/:task/*proxyPath", proxyLLM("http://localhost:8081", db))
//
//	r.Run(":8080")
//}
