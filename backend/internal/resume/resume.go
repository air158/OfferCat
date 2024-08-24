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
	UserID     uint      `json:"user_id"`                     // 关联的用户ID
	FilePath   string    `json:"file_path"`                   // 简历文件的存储路径
	FileName   string    `json:"file_name"`                   // 简历文件的原始名称
	UploadedAt time.Time `json:"uploaded_at"`                 // 简历上传时间
	Status     string    `json:"status"`                      // 简历状态（如"待审核"，"审核完成"等）
	Content    string    `json:"content,omitempty"`           // 简历内容，可能是解析后的文本内容
	Suggestion string    `json:"suggestion,omitempty"`        // 简历改进建议
	IsDeleted  bool      `json:"is_deleted" gorm:"default:0"` // 是否已删除
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
	err := db.DB.
		Where("user_id = ? ", userID).
		Where("is_deleted = 0").
		Find(&resumes).Error
	return resumes, err
}
func GetResumeList(c *gin.Context) {
	uid := lib.Uid(c)
	resumes, err := GetResumeListByUserID(uint(uid))
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "获取简历列表失败", err)
		return
	}
	lib.Ok(c, "获取简历列表成功", gin.H{
		"resumes": resumes,
	})
}

type DeleteRequest struct {
	ID uint `json:"id"`
}

func DeleteResumeByID(c *gin.Context) {
	uid := lib.Uid(c)
	var req DeleteRequest
	var resume Resume
	err := c.ShouldBindJSON(&req)
	if err != nil {
		lib.Err(c, http.StatusBadRequest, "参数错误", err)
	}
	resume.ID = req.ID
	err = db.DB.Model(&resume).Where("user_id=?", uid).Where("id = ? ", resume.ID).Update("is_deleted", 1).Error
	if err != nil {
		if err.Error() == "record not found" {
			lib.Err(c, http.StatusNotFound, "简历不存在", err)
			return
		} else {
			lib.Err(c, http.StatusInternalServerError, "删除简历失败", err)
			return
		}
	}

	lib.Ok(c, "删除简历成功", gin.H{
		"resume_id": resume.ID,
	})
}

func UpdateResumeByID(id uint, resume *Resume) error {
	return db.DB.Model(&Resume{}).Where("id = ?", id).Updates(resume).Error
}
