package interview

import "time"

type Question struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	InterviewID  uint      `json:"interview_id"`  // 关联的面试记录ID
	Content      string    `json:"content"`       // 问题的内容
	QuestionType string    `json:"question_type"` // 问题类型（如"技术问题"，"行为问题"等）
	CreatedAt    time.Time `json:"created_at"`    // 问题创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 问题更新时间
}
