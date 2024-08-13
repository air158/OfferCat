package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"offercat/v0/internal/auth"
	"offercat/v0/internal/db"
	"offercat/v0/internal/interview"
	"offercat/v0/internal/resume"
	"offercat/v0/internal/utils"
)

func main() {
	var err error
	// This is the main function
	// This is the entry point of the application
	db.InitDB()
	log.Println("DB connected")
	// 自动迁移所有模型
	err = db.DB.AutoMigrate(
		&auth.User{},
		&interview.InterviewRecord{},
		&interview.Question{},
		&resume.Resume{},
		&auth.EmailVerification{},
	)
	if err != nil {
		panic("failed to migrate database" + err.Error())
	}

	r := gin.Default()
	r.Use(cors.Default())
	// 注册接口
	r.POST("/register", auth.EmailRegister)
	// 邮箱验证接口
	r.GET("/verify", auth.VerifyEmail)
	// 登录接口
	r.POST("/login", auth.Login)

	r.GET("/ping", utils.Ping)

	// 受保护的路由
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			username, _ := c.Get("username")
			role, _ := c.Get("role")

			c.JSON(200, gin.H{
				"userID":   userID,
				"username": username,
				"role":     role,
			})
		})
	}
	r.Run(":12345")
}
