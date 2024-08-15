package interview

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

// Interview 定义模拟面试的结构体
type Interview struct {
	ID             uint      `json:"id" gorm:"primaryKey" form:"id"`
	UserID         uint      `json:"user_id"`
	SimulationDate time.Time `json:"simulation_date"`
	LLMModel       string    `json:"llm_model"`             // 使用的 LLM 模型
	Performance    string    `json:"performance,omitempty"` // 用户在模拟面试中的表现
	InterviewRole  string    `json:"interview_role"`        // 面试角色
	InterviewStyle string    `json:"interview_style"`       // 面试风格
	FinalSummary   string    `json:"final_summary"`         // 最终评价
	Type           string    `json:"type"`                  // 模拟面试类型
	FeedbackID     uint      `json:"feedback_id,omitempty"` // 反馈ID
	Dialog         string    `json:"dialog_id,omitempty"`   // 对话
}

func CreateSimulatedInterview(c *gin.Context) {
	// 从请求中解析模拟面试信息
	var entity Interview
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity.UserID = uint(lib.Uid(c))
	entity.SimulationDate = time.Now()

	// 将模拟面试信息保存到数据库
	if err := db.DB.Create(&entity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func GetSimulatedInterview(c *gin.Context) {
	//根据id获取模拟面试信息
	var entity Interview
	id := c.Param("id")
	if err := db.DB.Where("id = ?", id).First(&entity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}
