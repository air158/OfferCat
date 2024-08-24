package saver

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	interview2 "offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
)

import (
	"errors"
)

func SaveInterviewResult(db *gorm.DB, requestData map[string]interface{}, completeData string, c *gin.Context) error {
	id := uint(requestData["interview_id"].(float64))
	userID := uint(lib.Uid(c))
	finalSummary := completeData
	var interview interview2.Interview

	err := db.Where("id = ? AND user_id=? ", id, userID).First(&interview).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			interview = interview2.Interview{
				ID:           id,
				UserID:       userID,
				FinalSummary: finalSummary,
				Dialog:       requestData["prompt_text"].(string),
			}
			return db.Create(&interview).Error
		}
		return err
	}

	interview.FinalSummary = finalSummary
	interview.Dialog = requestData["prompt_text"].(string)
	return db.Save(&interview).Error
}
