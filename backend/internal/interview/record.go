package interview

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"offercat/v0/internal/lib"
)

type QueryInterviewResultRequest struct {
	InterviewID uint `json:"interview_id" binding:"required"`
}

type InterviewResultResponse struct {
	QuestionContent string `json:"question_content"`
	UserAnswer      string `json:"user_answer"`
	LLMAnswer       string `json:"llm_answer"`
}

func QueryInterviewResult(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QueryInterviewResultRequest

		// 解析请求体中的 JSON 数据
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// 鉴权
		uid := lib.Uid(c)
		if err := db.Where("id = ? AND user_id = ?", req.InterviewID, uid).First(&Interview{}).Error; err != nil {
			log.Println("Error fetching interview:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized Or Interview Not Found"})
		}

		// 查询问题和用户答案
		var questions []Question
		if err := db.Where("interview_id = ?", req.InterviewID).Find(&questions).Error; err != nil {
			log.Println("Error fetching questions:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
			return
		}

		// 查询对应的答案
		var answers []Answer
		if err := db.Where("interview_id = ?", req.InterviewID).Find(&answers).Error; err != nil {
			log.Println("Error fetching answers:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch answers"})
			return
		}

		// 组装结果
		var results []InterviewResultResponse
		for _, question := range questions {
			for _, answer := range answers {
				if question.ID == answer.QuestionID {
					result := InterviewResultResponse{
						QuestionContent: question.Content,
						UserAnswer:      answer.Content,
						LLMAnswer:       answer.LLMAnswer,
					}
					results = append(results, result)
				}
			}
		}

		// 返回结果
		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}
