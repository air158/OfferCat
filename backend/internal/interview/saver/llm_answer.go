package saver

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
	"time"
)

func SaveLLMAnswer(db *gorm.DB, requestData map[string]interface{}, completeData string, c *gin.Context) error {
	questionBranchID := (uint)(requestData["question_branch_id"].(float64))
	interviewID := (uint)(requestData["interview_id"].(float64))
	userID := uint(lib.Uid(c))

	var userAnswer common.Answer
	err := db.Where("interview_id = ? AND question_branch_id=? AND user_id = ?", interviewID, questionBranchID, userID).First(&userAnswer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userAnswer = common.Answer{
				InterviewID:      interviewID,
				QuestionBranchID: questionBranchID,
				UserID:           userID,
				LLMAnswer:        completeData,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			return db.Create(&userAnswer).Error
		}
		return err
	}

	userAnswer.LLMAnswer = completeData
	userAnswer.UpdatedAt = time.Now()
	return db.Save(&userAnswer).Error
}
