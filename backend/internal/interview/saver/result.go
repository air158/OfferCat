package saver

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	interview2 "offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
	"time"
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
		//lib.Err(c, 500, "数据库错误", err)
		//return err
	}

	interview.FinalSummary = finalSummary
	interview.Dialog = requestData["prompt_text"].(string)
	if interview.StartTime.IsZero() {
		interview.StartTime = time.Now()
	}
	lib.Ok(c, "面试结果保存成功", gin.H{
		"interview_id":  id,
		"final_summary": finalSummary,
	})
	return db.Save(&interview).Error
}
