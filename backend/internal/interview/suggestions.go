package interview

import (
	"fmt"
	"gorm.io/gorm"
	"log"
)

// 定义返回的字符串格式
func FormatInterviewResult(db *gorm.DB, interviewID uint) (string, error) {
	var questions []Question
	var answers []Answer

	// 查询问题和用户答案
	if err := db.Where("interview_id = ?", interviewID).Find(&questions).Error; err != nil {
		log.Println("Error fetching questions:", err)
		return "", err
	}

	questionId := GetQuestionIdByInterviewId1(interviewID)
	// 查询对应的答案
	if err := db.Where("question_id = ?", questionId).Find(&answers).Error; err != nil {
		log.Println("Error fetching answers:", err)
		return "", err
	}

	// 组装结果字符串
	var resultStr string
	for _, question := range questions {
		for _, answer := range answers {
			if question.ID == answer.QuestionID {
				record := fmt.Sprintf("面试官: “%s” 面试者: “%s”\n", question.Content, answer.Content)
				resultStr += record
			}
		}
	}

	return resultStr, nil
}
