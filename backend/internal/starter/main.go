package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"offercat/v0/internal/auth"
	"offercat/v0/internal/db"
	"offercat/v0/internal/interview"
	"offercat/v0/internal/job"
	"offercat/v0/internal/resume"
	"offercat/v0/internal/utils"
)

func main() {
	var err error
	// This is the main function
	// This is the entry point of the application
	db.InitDB()
	// 设置环境变量
	//os.Setenv("SPARK_API_KEY", "a7c55761823e132301eacceb043913f2")
	//os.Setenv("SPARK_API_SECRET", "M2NjYzIzZmIyNmJmNGEyNzIyNGRhOGZi")

	log.Println("DB connected")
	// 自动迁移所有模型
	err = db.DB.AutoMigrate(
		&auth.User{},
		&interview.InterviewRecord{},
		&interview.Question{},
		&resume.Resume{},
		&auth.EmailVerification{},
		&interview.Preset{},
		&job.PresetJob{},
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

	// 开发者使用
	// 预设岗位添加
	r.POST("/preset/job/create", job.CreateJob)

	// 受保护的路由
	protected := r.Group("/api")
	protected.Use(auth.JWTAuthMiddleware())
	{
		// 面试预设接口
		protected.POST("/preset/upsert", interview.UpsertPreset)
		// 获取面试预设接口
		protected.GET("/preset", interview.GetPreset)

		// 上传简历接口
		protected.POST("/preset/resume/upload/pdf", interview.UploadResumePDF)

		// 简历评价接口
		protected.GET("/preset/resume/evaluate", interview.ResumeSuggestion)

		// 获取单个岗位信息
		protected.GET("/preset/job", job.GetJobByTitle)

		// 获取简历列表
		protected.GET("/resume", resume.GetResumeList)

		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("uid")
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
