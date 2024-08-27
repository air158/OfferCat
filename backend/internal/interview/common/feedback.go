package common

import "time"

// Feedback 用户反馈
type Feedback struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	InterviewID  uint      `json:"interview_id"`     // 关联的面试记录ID
	UserID       uint      `json:"user_id"`          // 提供反馈的用户ID，可能是面试官或者系统
	FeedbackType string    `json:"feedback_type"`    // 反馈类型（如"面试官反馈"，"系统反馈"等）
	Content      string    `json:"content"`          // 反馈的具体内容
	Rating       int       `json:"rating,omitempty"` // 评分（如果有打分机制）
	CreatedAt    time.Time `json:"created_at"`       // 反馈创建时间
	UpdatedAt    time.Time `json:"updated_at"`       // 反馈更新时间
}
