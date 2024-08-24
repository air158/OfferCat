package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"offercat/v0/internal/auth"
	middleware2 "offercat/v0/internal/auth/middleware"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/auth/redeem"
	"offercat/v0/internal/common/middleware"
	"offercat/v0/internal/db"
	"offercat/v0/internal/interview"
	common2 "offercat/v0/internal/interview/common"
	"offercat/v0/internal/interview/proxy"
	"offercat/v0/internal/job"
	"offercat/v0/internal/resume"
	"offercat/v0/internal/utils"
)

// This is the main function
// This is the entry point of the application
func main() {
	var err error
	db.InitDB()

	log.Println("DB connected")
	// 自动迁移所有模型
	err = db.DB.AutoMigrate(
		&model.User{},
		&common2.Question{},
		&resume.Resume{},
		&auth.EmailVerification{},
		&common2.Preset{},
		&job.PresetJob{},
		&common2.Interview{},
		&common2.Answer{},
		&redeem.RedeemCode{},
	)
	if err != nil {
		panic("failed to migrate database" + err.Error())
	}

	r := gin.Default()

	r.Use(cors.Default())
	r.Use(middleware.ResponseMiddleware())

	// 注册接口
	r.POST("/api/register", auth.EmailRegister)
	// 邮箱验证接口
	r.GET("/api/verify", auth.VerifyEmail)
	// 登录接口
	r.POST("/api/login", auth.Login)
	r.GET("/ping", utils.Ping)
	r.GET("/api/preset/job/list", job.GetPresetJobList)

	// 受保护的路由
	protected := r.Group("/api")
	// 使用 ResponseMiddleware 中间件
	protected.Use(middleware.ResponseMiddleware())
	protected.Use(middleware2.JWTAuthMiddleware())
	{
		adminGroup := protected.Group("")
		adminGroup.Use(middleware2.AdminMiddleware())
		{
			adminGroup.GET("/admin/ping", utils.Ping)
			// 开发者使用
			// 预设岗位添加
			adminGroup.POST("/preset/job/create", job.CreateJob)
			adminGroup.POST("/redeem-code/create", redeem.CreateCode)
			adminGroup.POST("/redeem-code/create/batch", redeem.CreateBatchCode)

		}
		// 需要激活码验证的受保护路由组
		activated := protected.Group("")
		activated.Use(redeem.RedeemMiddleware())
		{

			activated.GET("/activated/ping", utils.Ping)

			activated.POST("/interview/register", common2.UpsertPresetAndCreateInterview)
			activated.POST("/simulation/answer", common2.CreateOrUpdateAnswer)
			// 简历评价接口
			activated.GET("/preset/resume/evaluate", common2.ResumeSuggestion)
			// 流代理
			activated.Any("/:mod/:task/proxy", proxy.ProxyLLM("http://localhost:5000", db.DB))
			// 关闭面试
			activated.POST("/interview/close", interview.CloseInterview)
		}
		// 激活码验证通过后的受保护路由
		// 获取面试预设接口
		protected.GET("/preset", common2.GetPreset)
		// 更新预设
		protected.POST("/preset/upsert", common2.UpsertPreset)
		// 上传简历
		protected.POST("/preset/resume/upload/pdf", common2.UploadResumePDF)
		// 获取单个岗位信息
		protected.GET("/preset/job", job.GetJobByTitle)
		// 获取简历列表
		protected.GET("/resume", resume.GetResumeList)
		//获取问答记录
		protected.POST("/interview/record", common2.QueryInterviewResult(db.DB))
		//获取单个面试信息
		protected.GET("/simulation/:id", common2.GetSimulatedInterview)
		// 一个暂时废弃的接口
		//activated.GET("/interview/question", common2.GetQuestionIdByInterviewId)
		// 获取面试（历史）列表
		protected.GET("/interview/list", common2.GetInterviewListByUid)
		// 删除简历
		protected.POST("/resume/delete", resume.DeleteResumeByID)
		// 设置生日
		protected.POST("/user/birth/set", model.SetBirth)
		protected.POST("/redeem-code/verify", redeem.VerifyCode)

		protected.GET("/profile", auth.GetProfile)
	}
	err = r.Run(":12345")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
