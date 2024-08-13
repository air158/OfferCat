package interview

import "time"

// InterviewRecord 定义面试记录的结构体
type InterviewRecord struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        uint       `json:"user_id"`
	InterviewDate time.Time  `json:"interview_date"`
	Position      string     `json:"position"`
	Company       string     `json:"company"`
	Questions     []Question `json:"questions,omitempty" gorm:"foreignKey:InterviewID"` // 面试中问到的问题列表
	Responses     []Response `json:"responses,omitempty" gorm:"foreignKey:InterviewID"` // 用户回答的列表
	Feedback      []Feedback `json:"feedback,omitempty" gorm:"foreignKey:InterviewID"`  // 面试官或系统的反馈
}

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
