package common

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

type Question struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	BranchID    uint   `json:"branch_id" `                       // 关联的Question ID//废除
	InterviewID uint   `json:"interview_id" form:"interview_id"` // 关联的面试记录ID
	UserID      uint   `json:"user_id"`                          // 回答者的用户ID
	Content     string `json:"content"`                          // 回答的内容

	CreatedAt time.Time `json:"created_at"` // 回答创建时间
	UpdatedAt time.Time `json:"updated_at"` // 回答更新时间
}

func CreateQuestion(question Question) error {
	return db.DB.Create(&question).Error
}

func GetQuestionIdByInterviewId(c *gin.Context) {
	var question Question
	err := c.BindQuery(&question)
	if err != nil {
		lib.Err(c, 400, "解析请求体失败", err)
		return
	}
	err = db.DB.Where("interview_id = ? and user_id=?", question.InterviewID, lib.Uid(c)).First(&question).Error
	if err != nil {
		lib.Err(c, 400, "查询问题失败", err)
		return
	}
	lib.Ok(c, "查询问题成功", gin.H{
		"question_id": question.ID,
	})

}
func GetQuestionIdByInterviewId1(interviewId uint) (questionId int) {
	var question Question
	err := db.DB.Where("interview_id = ?", interviewId).First(&question).Error
	if err != nil {
		return 0
	}
	return int(question.ID)
}
