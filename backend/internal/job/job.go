package job

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
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
		lib.Err(c, http.StatusInternalServerError, "无法查询到岗位信息", err)

		return
	}
	lib.Ok(c, "获取岗位信息成功", jobs)
}

// 根据job_title获取岗位信息
func GetJobByTitle(c *gin.Context) {
	var job PresetJob
	jobTitle := c.Query("job_title")
	if err := db.DB.Where("job_title = ?", jobTitle).First(&job).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "无法查询到岗位信息", err)
		return
	}
	lib.Ok(c, "获取岗位信息成功", job)
}

// CreateJob 创建新岗位信息
func CreateJob(c *gin.Context) {
	var job PresetJob
	if err := c.ShouldBindJSON(&job); err != nil {
		lib.Err(c, http.StatusBadRequest, "无法解析JSON数据", err)
		return
	}

	if err := db.DB.Create(&job).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "创建岗位信息失败", err)
		return
	}
	lib.Ok(c, "创建岗位信息成功", job)
}

// UpdateJob 更新岗位信息
func UpdateJob(c *gin.Context, db *gorm.DB) {
	var job PresetJob
	id := c.Param("id")
	if err := db.First(&job, id).Error; err != nil {
		lib.Err(c, http.StatusNotFound, "岗位信息不存在", err)
		return
	}

	if err := c.ShouldBindJSON(&job); err != nil {
		lib.Err(c, http.StatusBadRequest, "无法解析JSON数据", err)
		return
	}

	db.Save(&job)
	lib.Ok(c, "更新岗位信息成功", job)
}

// DeleteJob 删除岗位信息
func DeleteJob(c *gin.Context, db *gorm.DB) {
	var job PresetJob
	id := c.Param("id")
	if err := db.First(&job, id).Error; err != nil {
		lib.Err(c, http.StatusNotFound, "岗位信息不存在", err)
		return
	}

	if err := db.Delete(&job, id).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "删除岗位信息失败", err)
		return
	}
	lib.Ok(c, "删除岗位信息成功", nil)
}
