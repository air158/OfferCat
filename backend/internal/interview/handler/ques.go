package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
	"offercat/v0/internal/resume"
	pdf_analyser "offercat/v0/internal/thirdparty/pdf-analyser"
)

func HandleQuestionTask(c *gin.Context, db *gorm.DB, requestData map[string]interface{}) error {
	var preset common.Preset
	err := db.Where("user_id=?", lib.Uid(c)).First(&preset).Error
	if err != nil {
		return err
	}

	requestData["job_title"] = preset.JobTitle
	requestData["job_description"] = preset.JobDescription

	var r resume.Resume
	err = db.Where("id=? and user_id=?", preset.ResumeID, lib.Uid(c)).First(&r).Error
	if err != nil {
		return err
	}

	if r.Content == "" {
		r.Content = pdf_analyser.GetStringFromPDF(c, r.FilePath)
		if err := db.Save(&r).Error; err != nil {
			return err
		}
	}
	requestData["resume_text"] = r.Content
	return nil
}
