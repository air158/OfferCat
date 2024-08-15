package auth

import (
	"offercat/v0/internal/interview"
	"offercat/v0/internal/resume"
	"time"
)

// 定义用户的结构
type User struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"unique;not null"`
	PasswordHash   string    `json:"-"` // 存储加密后的密码,不进行传输
	Email          string    `json:"email" gorm:"unique;not null"`
	Role           string    `json:"role"` // 用户角色，如"candidate"（候选人）、"interviewer"（面试官）、"admin"（管理员）
	DateOfBirth    time.Time `json:"date_of_birth,omitempty"`
	ProfilePicture string    `json:"profile_picture,omitempty"` // 用户头像
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLogin      time.Time `json:"last_login,omitempty"`
	Valid          bool      `json:"valid,omitempty"` // 用户是否有效

	JobProfile JobProfile `json:"job_profile"` // 用户的工作相关信息
	// Additional attributes
	Resume               []resume.Resume       `json:"resume,omitempty"`               // 候选人的简历
	SimulatedInterviews  []interview.Interview `json:"simulated_interviews,omitempty"` // 进行的模拟面试
	Feedback             []interview.Feedback  `json:"feedback,omitempty"`             // 用户收到的反馈
	Language             string                `json:"language"`
	NotificationSettings string                `json:"notification_settings,omitempty"` // 用户的通知设置          // 用户偏好设置
	// JWT Token Tracking (optional)
	JWTToken string `json:"-"`
}

type JobProfile struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	UserID             uint   `json:"user_id"`                    // 关联的用户ID
	DesiredJobTitle    string `json:"desired_job_title"`          // 用户期望的职位名称
	DesiredJobType     string `json:"desired_job_type"`           // 期望的职位类型（全职、兼职、实习等）
	DesiredLocation    string `json:"desired_location"`           // 期望的工作地点
	DesiredSalaryRange string `json:"desired_salary_range"`       // 期望的薪资范围
	DesiredIndustry    string `json:"desired_industry,omitempty"` // 期望的行业领域
	JobDescription     string `json:"job_description"`            // 用户期望的职位描述

	CurrentJobTitle   string `json:"current_job_title,omitempty"`    // 当前职位
	CurrentCompany    string `json:"current_company,omitempty"`      // 当前公司
	YearsAtCurrentJob int    `json:"years_at_current_job,omitempty"` // 在当前职位的工作年限
	CurrentSalary     string `json:"current_salary,omitempty"`       // 当前薪资
	CurrentIndustry   string `json:"current_industry,omitempty"`     // 当前行业领域
	AIResourcePath    string `json:"ai_resource_path,omitempty"`     // 职位描述文件路径，供AI后续读取的资源
}
