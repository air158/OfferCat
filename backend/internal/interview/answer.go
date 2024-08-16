package interview

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

type Answer struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
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

func CreateOrUpdateAnswer(c *gin.Context) {
	// 从请求中解析用户答案
	var answer Answer
	if err := c.ShouldBindJSON(&answer); err != nil {
		lib.Err(c, http.StatusBadRequest, "解析用户答案失败，可能是不合法的输入", err)
		return
	}

	// 设置用户ID和时间戳
	answer.UserID = uint(lib.Uid(c))
	answer.UpdatedAt = time.Now()

	// 在数据库中查找是否已有该 question 和 branch 的答案
	var existingAnswer Answer
	if err := db.DB.Where("user_id = ? AND question_id = ? AND question_branch_id = ?", answer.UserID, answer.QuestionID, answer.QuestionBranchID).First(&existingAnswer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有找到记录，则创建新记录
			answer.CreatedAt = time.Now()
			if err := db.DB.Create(&answer).Error; err != nil {
				lib.Err(c, http.StatusInternalServerError, "保存用户答案失败", err)
				return
			}
			lib.Ok(c, "创建用户答案成功", answer)
			return
		} else {
			// 处理其他查询错误
			lib.Err(c, http.StatusInternalServerError, "查询用户答案失败", err)
			return
		}
	}

	// 如果找到记录，则更新现有记录
	existingAnswer.Content = answer.Content
	existingAnswer.UpdatedAt = time.Now()

	if err := db.DB.Save(&existingAnswer).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "更新用户答案失败", err)
		return
	}

	lib.Ok(c, "更新用户答案成功", existingAnswer)
}
