package model

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
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
	VipExpireAt    time.Time `json:"vip_expire_at,omitempty"`   // vip到期时间
	InterviewPoint int       `json:"interview_point,omitempty"` // 面试点数
}

type SetBirthRequest struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func SetBirth(c *gin.Context) {
	uid := lib.Uid(c)
	var req SetBirthRequest
	c.ShouldBindJSON(&req)
	var user User
	db.DB.Where("id = ?", uid).First(&user)

	user.DateOfBirth = time.Date(req.Year, time.Month(req.Month), req.Day, 0, 0, 0, 0, time.UTC)
	db.DB.Save(&user)
	lib.Ok(c, "设置生日成功", gin.H{
		"date_of_birth": user.DateOfBirth,
	})

}
