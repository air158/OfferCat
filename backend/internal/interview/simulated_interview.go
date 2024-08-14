package interview

import "time"

// SimulatedInterview 定义模拟面试的结构体
type SimulatedInterview struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id"`
	SimulationDate time.Time `json:"simulation_date"`
	Scenario       string    `json:"scenario"`              // 模拟面试场景
	LLMModel       string    `json:"llm_model"`             // 使用的 LLM 模型
	Performance    string    `json:"performance,omitempty"` // 用户在模拟面试中的表现
	Feedback       string    `json:"feedback,omitempty"`    // 系统生成的反馈
}

type Resume struct {
	ID      int    `json:"id" gorm:"primaryKey"`
	UserID  int    `json:"user_id"`
	Content string `json:"content"`
}
