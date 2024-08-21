package auth

import (
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
	Language       string    `json:"language"`
}
