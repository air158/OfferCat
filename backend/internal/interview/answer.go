package interview

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

type Answer struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	InterviewID      uint      `json:"interview_id"`
	QuestionID       uint      `json:"question_id"`
	QuestionBranchID uint      `json:"question_branch_id"`
	UserID           uint      `json:"user_id"`
	TimeSpent        uint      `json:"time_spent"`
	Content          string    `json:"content"`
	LLMAnswer        string    `json:"llm_answer"`      // LLM 模型生成的答案
	Score            bool      `json:"score,omitempty"` // 得分
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TODO:改成upsert
func CreateAnswer(c *gin.Context) {
	// 从请求中解析用户答案
	var answer Answer
	if err := c.ShouldBindJSON(&answer); err != nil {
		lib.Err(c, http.StatusBadRequest, "解析用户答案失败，可能是不合法的输入", err)
		return
	}
	answer.UserID = uint(lib.Uid(c))
	answer.CreatedAt = time.Now()
	answer.UpdatedAt = time.Now()

	// 将用户答案保存到数据库
	if err := db.DB.Create(&answer).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "保存用户答案失败", err)
		return
	}

	lib.Ok(c, "保存用户答案成功", answer)
}
