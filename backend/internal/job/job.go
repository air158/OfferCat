package job

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"offercat/v0/internal/db"
)

type PresetJob struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	JobTitle       string `gorm:"type:varchar(100);not null" json:"job_title" form:"job_title"`
	JobDescription string `gorm:"type:text;not null " json:"job_description"`
}

// GetJobs 获取所有岗位信息
func GetJobs(c *gin.Context, db *gorm.DB) {
	var jobs []PresetJob
	if err := db.Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// 根据job_title获取岗位信息
func GetJobByTitle(c *gin.Context) {
	var job PresetJob
	jobTitle := c.Query("job_title")
	if err := db.DB.Where("job_title = ?", jobTitle).First(&job).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// CreateJob 创建新岗位信息
func CreateJob(c *gin.Context) {
	var job PresetJob
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}

// UpdateJob 更新岗位信息
func UpdateJob(c *gin.Context, db *gorm.DB) {
	var job PresetJob
	id := c.Param("id")
	if err := db.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&job)
	c.JSON(http.StatusOK, job)
}

// DeleteJob 删除岗位信息
func DeleteJob(c *gin.Context, db *gorm.DB) {
	var job PresetJob
	id := c.Param("id")
	if err := db.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	if err := db.Delete(&job, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}
