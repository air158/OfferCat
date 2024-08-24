package saver

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"offercat/v0/internal/interview/common"
	"offercat/v0/internal/lib"
	"strings"
	"time"
)

func SaveQuestions(db *gorm.DB, requestData map[string]interface{}, completeData string, c *gin.Context) error {
	questions := strings.Split(completeData, `¥¥`)

	if len(questions) > 0 {
		questions[len(questions)-1] = strings.TrimSuffix(questions[len(questions)-1], "[DONE]")
	}

	err := db.First(&common.Question{}, "interview_id = ? and user_id=?", requestData["interview_id"], lib.Uid(c)).Error
	if err != nil {
		// 没有找到记录，插入所有问题
		for i, content := range questions {
			db.Create(&common.Question{
				InterviewID: (uint)(requestData["interview_id"].(float64)),
				UserID:      uint(lib.Uid(c)),
				BranchID:    uint(i + 1),
				Content:     content,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			})
		}
	} else {
		// 找到记录，更新所有问题
		for i, content := range questions {
			db.Model(&common.Question{}).
				Where("interview_id = ? and user_id=? and branch_id=?", requestData["interview_id"], lib.Uid(c), i+1).
				Updates(map[string]interface{}{"content": content, "updated_at": time.Now()})
		}
	}

	var preset common.Preset
	if err = db.Where("user_id=?", lib.Uid(c)).First(&preset).Error; err == nil {
		preset.QuestionLength = int(requestData["ques_len"].(float64))
		db.Save(&preset)
	}

	return nil
}
