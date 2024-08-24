package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
)

func HandleResultTask(c *gin.Context, db *gorm.DB, requestData map[string]interface{}) error {
	promptText, err := common.FormatInterviewResult(db, uint(requestData["interview_id"].(float64)))
	if err != nil {
		return err
	}
	requestData["prompt_text"] = promptText

	var preset common.Preset
	err = db.Where("user_id=?", lib.Uid(c)).First(&preset).Error
	if err != nil {
		return err
	}

	requestData["job_title"] = preset.JobTitle
	return nil
}
