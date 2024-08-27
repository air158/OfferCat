package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
)

func HandleLLMAnswerTask(c *gin.Context, db *gorm.DB, requestData map[string]interface{}) error {
	var preset common.Preset
	err := db.Where("user_id=?", lib.Uid(c)).First(&preset).Error
	if err != nil {
		return err
	}

	requestData["job_title"] = preset.JobTitle
	requestData["job_description"] = preset.JobDescription

	var question common.Question
	err = db.Where("branch_id=? and interview_id=? and user_id=?", requestData["question_branch_id"], requestData["interview_id"], lib.Uid(c)).First(&question).Error
	if err != nil {
		return err
	}

	requestData["question"] = question.Content
	return nil
}
