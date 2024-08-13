package interview

import "time"

type Response struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	QuestionID  uint      `json:"question_id"`     // 关联的Question ID
	InterviewID uint      `json:"interview_id"`    // 关联的面试记录ID
	UserID      uint      `json:"user_id"`         // 回答者的用户ID
	Content     string    `json:"content"`         // 回答的内容
	Score       bool      `json:"score,omitempty"` // 得分
	CreatedAt   time.Time `json:"created_at"`      // 回答创建时间
	UpdatedAt   time.Time `json:"updated_at"`      // 回答更新时间
}
