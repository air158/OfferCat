package resume

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

type Resume struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id"`              // 关联的用户ID
	FilePath   string    `json:"file_path"`            // 简历文件的存储路径
	FileName   string    `json:"file_name"`            // 简历文件的原始名称
	UploadedAt time.Time `json:"uploaded_at"`          // 简历上传时间
	Status     string    `json:"status"`               // 简历状态（如"待审核"，"审核完成"等）
	Content    string    `json:"content,omitempty"`    // 简历内容，可能是解析后的文本内容
	Suggestion string    `json:"suggestion,omitempty"` // 简历改进建议
}

type ResumeFeedback struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ResumeID     uint      `json:"resume_id" gorm:"index"` // 关联的简历ID，创建索引以提高查询效率
	UserID       uint      `json:"user_id,omitempty"`      // 提供建议的用户ID，可能是管理员或系统
	FeedbackType string    `json:"feedback_type"`          // 建议类型（如"系统建议"，"人工建议"等）
	Content      string    `json:"content"`                // 建议的具体内容
	CreatedAt    time.Time `json:"created_at"`             // 建议创建时间
	UpdatedAt    time.Time `json:"updated_at"`             // 建议更新时间
}

func NewResume(resume *Resume) error {
	err := db.DB.Create(resume)
	return err.Error
}

func GetResumeByID(id uint) (*Resume, error) {
	var resume Resume
	err := db.DB.Where("id = ?", id).First(&resume).Error
	return &resume, err
}

func GetResumeListByUserID(userID uint) ([]Resume, error) {
	var resumes []Resume
	err := db.DB.Where("user_id = ?", userID).Find(&resumes).Error
	return resumes, err
}
func GetResumeList(c *gin.Context) {
	uid := lib.Uid(c)
	resumes, err := GetResumeListByUserID(uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resumes)
}

func DeleteResumeByID(id uint) error {
	return db.DB.Delete(&Resume{}, id).Error
}

func UpdateResumeByID(id uint, resume *Resume) error {
	return db.DB.Model(&Resume{}).Where("id = ?", id).Updates(resume).Error
}
