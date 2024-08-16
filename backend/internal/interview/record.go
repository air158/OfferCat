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
			lib.Err(c, http.StatusBadRequest, "解析请求体失败", err)
			return
		}
		// 鉴权
		uid := lib.Uid(c)
		if err := db.Where("id = ? AND user_id = ?", req.InterviewID, uid).First(&Interview{}).Error; err != nil {
			log.Println("Error fetching interview:", err)
			lib.Err(c, http.StatusUnauthorized, "未授权或面试不存在", err)
			return
		}

		// 查询问题和用户答案
		var questions []Question
		if err := db.Where("interview_id = ?", req.InterviewID).Find(&questions).Error; err != nil {
			log.Println("Error fetching questions:", err)
			lib.Err(c, http.StatusInternalServerError, "获取问题失败", err)
			return
		}

		// 查询对应的答案
		var answers []Answer
		if err := db.Where("interview_id = ?", req.InterviewID).Find(&answers).Error; err != nil {
			log.Println("Error fetching answers:", err)
			lib.Err(c, http.StatusInternalServerError, "获取答案失败", err)
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
		lib.Ok(c, "获取面试结果成功", gin.H{
			"results": results,
		})
	}
}
