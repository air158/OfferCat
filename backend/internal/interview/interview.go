package interview

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"reflect"
	"time"
)

// Interview 定义模拟面试的结构体
type Interview struct {
	ID                   uint      `json:"id" gorm:"primaryKey" form:"id"`
	UserID               uint      `json:"user_id"`
	SimulationDate       time.Time `json:"simulation_date"`
	LLMModel             string    `json:"llm_model"`               // 使用的 LLM 模型
	Performance          string    `json:"performance,omitempty"`   // 用户在模拟面试中的表现
	InterviewRole        string    `json:"interview_role"`          // 面试角色
	InterviewStyle       string    `json:"interview_style"`         // 面试风格
	FinalSummary         string    `json:"final_summary"`           // 最终评价
	Type                 string    `json:"type"`                    // 模拟面试类型
	FeedbackID           uint      `json:"feedback_id,omitempty"`   // 反馈ID
	Dialog               string    `json:"dialog_id,omitempty"`     // 对话
	TimeLimitPerQuestion int       `json:"time_limit_per_question"` // 每个问题的时间限制
}

type RegisterRequest struct {
	JobTitle       string `json:"job_title"`
	JobDescription string `json:"job_description"`
	Company        string `json:"company"`
	Business       string `json:"business"`
	Location       string `json:"location"`
	Progress       string `json:"progress"`  // 第几面
	ResumeID       uint   `json:"resume_id"` // 简历ID，会另外上传
	Language       string `json:"language"`
	InterviewStyle string `json:"interview_style"`

	InterviewRole        string `json:"interview_role"`          // 面试角色
	Type                 string `json:"type"`                    // 模拟面试类型
	TimeLimitPerQuestion int    `json:"time_limit_per_question"` // 每个问题的时间限制
}

func CreateSimulatedInterview(c *gin.Context) {
	// 从请求中解析模拟面试信息
	var entity Interview
	if err := c.ShouldBindJSON(&entity); err != nil {
		lib.Err(c, http.StatusBadRequest, "解析模拟面试信息失败，可能是不合法的输入", err)
		return
	}
	entity.UserID = uint(lib.Uid(c))
	entity.SimulationDate = time.Now()

	// 将模拟面试信息保存到数据库
	if err := db.DB.Create(&entity).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "保存模拟面试信息失败", err)
		return
	}

	lib.Ok(c, "保存模拟面试信息成功", entity)
}

func UpsertPresetAndCreateInterview(c *gin.Context) {
	var inputPreset Preset
	var registerRequest RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		lib.Err(c, http.StatusBadRequest, "解析用户预设信息失败，可能是不合法的输入", err)
		return
	}
	uidInt := lib.Uid(c)

	inputPreset = Preset{
		JobTitle:       registerRequest.JobTitle,
		JobDescription: registerRequest.JobDescription,
		Company:        registerRequest.Company,
		Business:       registerRequest.Business,
		Location:       registerRequest.Location,
		Progress:       registerRequest.Progress,
		ResumeID:       registerRequest.ResumeID,
		Language:       registerRequest.Language,
		InterviewStyle: registerRequest.InterviewStyle,
	}

	var existingPreset Preset
	// 在数据库中查找该用户的预设信息
	if err := db.DB.Where("user_id = ?", uidInt).First(&existingPreset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户没有现有的预设信息，创建新的
			existingPreset = Preset{
				UserID: uidInt,
			}
		} else {
			// 其他错误，返回错误响应
			lib.Err(c, http.StatusInternalServerError, "查询预设信息失败", err)
			return
		}
	}

	// 使用反射自动更新非零值的字段
	inputValue := reflect.ValueOf(&inputPreset).Elem()
	existingValue := reflect.ValueOf(&existingPreset).Elem()

	for i := 0; i < inputValue.NumField(); i++ {
		inputField := inputValue.Field(i)
		if !inputField.IsZero() {
			existingValue.Field(i).Set(inputField)
		}
	}

	// 保存或更新预设信息
	if err := db.DB.Save(&existingPreset).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "保存预设信息失败", err)
		return
	}
	interview := Interview{
		UserID:               uint(lib.Uid(c)),
		SimulationDate:       time.Now(),
		InterviewRole:        registerRequest.InterviewRole,
		InterviewStyle:       registerRequest.InterviewStyle,
		Type:                 registerRequest.Type,
		TimeLimitPerQuestion: registerRequest.TimeLimitPerQuestion,
	}
	// 将模拟面试信息保存到数据库
	if err := db.DB.Create(&interview).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "保存模拟面试信息失败", err)
		return
	}

	lib.Ok(c, "保存预设信息成功", gin.H{
		"preset":    existingPreset,
		"interview": interview,
	})

}

func GetSimulatedInterview(c *gin.Context) {
	//根据id获取模拟面试信息
	var entity Interview
	id := c.Param("id")
	if err := db.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "获取模拟面试信息失败", err)
		return
	}

	lib.Ok(c, "获取模拟面试信息成功", entity)
}
