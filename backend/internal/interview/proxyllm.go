package interview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"offercat/v0/internal/lib"
	"strings"
	"sync"
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

		// 解析目标URL并设置代理
		target, err := url.Parse(targetHost)
		proxy := httputil.NewSingleHostReverseProxy(target)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse target URL"})
			return
		}

		// 读取原始请求体
		bodyBytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
			return
		}

		// 解析请求体为JSON
		var requestData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// 添加 "chat_key" 字段
		requestData["chat_key"] = fmt.Sprintf("%s:%s", viper.GetString("spark.apiKey"), viper.GetString("spark.apiSecret"))
		log.Println("Modified request data:", requestData) // 添加日志

		// 将加工后的JSON重新编码为字节数组
		modifiedBodyBytes, err := json.Marshal(requestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode JSON"})
			return
		}

		// 创建一个新的请求，并将修改后的请求体设置为新的请求体
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
		c.Request.ContentLength = int64(len(modifiedBodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		// 确保路径被正确设置
		c.Request.URL.Path = proxyPath

		// 创建一个io.Pipe用于拦截和转发数据
		reader, writer := io.Pipe()
		var buffer bytes.Buffer
		var wg sync.WaitGroup

		proxy.ModifyResponse = func(response *http.Response) error {
			log.Println("ModifyResponse started")
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func(writer *io.PipeWriter) {
					err := writer.Close()
					if err != nil {
						log.Println("Error closing pipe writer:", err)
					}
				}(writer)

				buf := make([]byte, 4096) // 定义一个足够大的缓冲区
				for {
					n, err := response.Body.Read(buf)
					if n > 0 {
						buffer.Write(buf[:n])
						if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
							log.Println("Error writing to pipe:", writeErr)
							return
						}
					}
					if err != nil {
						if err != io.EOF {
							log.Println("Error reading from response body:", err)
						}
						break
					}
				}
			}()
			return nil
		}

		// 使用Goroutine将代理服务器的数据流转发给客户端
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Copying to client started")
			_, err := io.Copy(c.Writer, reader)
			if err != nil {
				log.Println("Error copying to client:", err)
				return
			}
		}()

		// 实际发送请求到目标服务器
		proxy.ServeHTTP(c.Writer, c.Request)

		// 等待所有的Goroutine完成
		wg.Wait()
		log.Println("All Goroutines finished")

		// 在这里将完整的数据保存到数据库
		completeData := buffer.String()
		log.Println("Complete data:", completeData)
		// 使用 strings.Builder 来高效地构建最终的字符串
		var finalBuilder strings.Builder

		// 分割字符串为行
		lines := strings.Split(completeData, "\n")

		for _, line := range lines {

			// 去掉行首的 "data: " 和空白
			cleanLine := strings.TrimSpace(strings.TrimPrefix(line, "data: "))

			// 忽略空行或只包含 ". " 的行
			if cleanLine != "" && cleanLine != "." {
				finalBuilder.WriteString(cleanLine)
				finalBuilder.WriteString("¥¥") // 添加 ¥¥ 作为行之间的分隔符
			}
		}

		// 获取最终结果
		completeData = finalBuilder.String()

		if task == "question" {
			question := Question{
				// 这个id注意前端json里面要用数字形式传进来
				InterviewID: (uint)(requestData["interview_id"].(float64)),
				UserID:      uint(lib.GetUid(c)),
				Content:     completeData,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			err = CreateQuestion(question)
			if err != nil {
				log.Println("Failed to save question to database:", err)
			} else {
				log.Println("Question saved to database")
			}
		}
		if task == "llm-answer" {
			questionBranchID := (uint)(requestData["question_branch_id"].(float64))
			questionID := (uint)(requestData["question_id"].(float64))
			userID := uint(lib.GetUid(c))

			// 尝试在数据库中查找对应的记录
			var userAnswer Answer
			err := db.Where("question_id = ? AND question_branch_id=? AND user_id = ?", questionID, questionBranchID, userID).First(&userAnswer).Error

			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// 如果记录不存在，则创建新的记录
					userAnswer = Answer{
						QuestionID:       questionID,
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
