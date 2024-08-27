package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"net/url"
	"offercat/v0/internal/interview/handler"
	"offercat/v0/internal/interview/saver"
	"offercat/v0/internal/lib"
	"strings"
)

func ProxyLLM(targetHost string, db *gorm.DB) gin.HandlerFunc {
	log.Println("ProxyLLM started")
	return func(c *gin.Context) {
		task := c.Param("task")
		log.Println("Task:", task)
		proxyPath := getProxyPath(task)
		if proxyPath == "" {
			lib.Err(c, http.StatusBadRequest, "无效的任务标识", nil)
			return
		}

		// 设置代理URL
		targetURL, err := url.Parse(targetHost)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "解析目标URL失败", err)
			return
		}

		// 处理请求体
		requestData, err := processRequestBody(c)
		if err != nil {
			lib.Err(c, http.StatusBadRequest, "解析请求体为JSON失败", err)
			return
		}

		// 根据任务标识处理请求数据
		switch task {
		case "question":
			err = handler.HandleQuestionTask(c, db, requestData)
		case "llm-answer":
			err = handler.HandleLLMAnswerTask(c, db, requestData)
		case "result":
			err = handler.HandleResultTask(c, db, requestData)
		}

		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "处理任务失败", err)
			return
		}

		// 发起代理请求
		err = forwardRequest(c, targetURL, proxyPath, requestData)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "执行代理请求失败", err)
			return
		}

		// 保存代理响应数据到数据库
		err = saveResponseToDatabase(db, task, requestData, c)
		if err != nil {
			lib.Err(c, http.StatusInternalServerError, "保存数据到数据库失败", err)
			return
		}
		//lib.Ok(c, "数据收集完成", gin.H{
		//	"completeData": c.GetString("completeData"),
		//})
		// 确保中止处理链
		//c.Abort()
	}
}

func getProxyPath(task string) string {
	switch task {
	case "question":
		return "/stream_questions"
	case "llm-answer":
		return "/stream_answer"
	case "result":
		return "/stream_result"
	default:
		return ""
	}
}

func processRequestBody(c *gin.Context) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		return nil, err
	}

	// 添加 "chat_key" 字段
	requestData["chat_key"] = fmt.Sprintf("%s:%s", viper.GetString("spark.apiKey"), viper.GetString("spark.apiSecret"))
	return requestData, nil
}

func forwardRequest(c *gin.Context, targetURL *url.URL, proxyPath string, requestData map[string]interface{}) error {
	modifiedBodyBytes, err := json.Marshal(requestData)
	if err != nil {
		return err
	}

	proxyReq, err := http.NewRequest(c.Request.Method, targetURL.ResolveReference(&url.URL{Path: proxyPath}).String(), bytes.NewBuffer(modifiedBodyBytes))
	if err != nil {
		return err
	}

	proxyReq.Header = c.Request.Header.Clone()
	proxyReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	streamResponse(c, resp.Body)
	return nil
}

func streamResponse(c *gin.Context, body io.Reader) {
	scanner := bufio.NewScanner(body)
	var completeDataB strings.Builder

	//c.Writer.Header().Set("Content-Type", "text/event-stream")
	//c.Writer.WriteHeader(http.StatusOK)
	//c.Writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		//_, err := c.Writer.Write([]byte(line + "\n"))
		//if err != nil {
		//	log.Printf("Failed to write data to client: %v", err)
		//	break
		//}
		//c.Writer.Flush()

		trimmed := strings.TrimPrefix(line, "data:")
		trimmed = strings.TrimSpace(trimmed)
		completeDataB.WriteString(trimmed)

	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading response body: %v", err)
	}

	// 将完整数据保存到数据库
	completeData := completeDataB.String()
	// 将所有收集到的数据构造成 JSON 对象

	log.Println("Complete data:", completeData)
	c.Set("completeData", completeData) // 保存到上下文，供后续使用
}

func saveResponseToDatabase(db *gorm.DB, task string, requestData map[string]interface{}, c *gin.Context) error {
	completeData := c.GetString("completeData") // 从上下文中获取完整数据

	switch task {
	case "question":
		return saver.SaveQuestions(db, requestData, completeData, c)
	case "llm-answer":
		return saver.SaveLLMAnswer(db, requestData, completeData, c)
	case "result":
		return saver.SaveInterviewResult(db, requestData, completeData, c)
	default:
		return nil
	}
}
