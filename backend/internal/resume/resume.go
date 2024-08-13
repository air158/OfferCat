package resume

import "time"

type Resume struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	UserID     uint             `json:"user_id"`            // 关联的用户ID
	FilePath   string           `json:"file_path"`          // 简历文件的存储路径
	FileName   string           `json:"file_name"`          // 简历文件的原始名称
	UploadedAt time.Time        `json:"uploaded_at"`        // 简历上传时间
	Status     string           `json:"status"`             // 简历状态（如"待审核"，"审核完成"等）
	Feedback   []ResumeFeedback `json:"feedback,omitempty"` // 简历建议列表
}
type ResumeFeedback struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ResumeID     uint      `json:"resume_id"`         // 关联的简历ID
	UserID       uint      `json:"user_id,omitempty"` // 提供建议的用户ID，可能是管理员或系统
	FeedbackType string    `json:"feedback_type"`     // 建议类型（如"系统建议"，"人工建议"等）
	Content      string    `json:"content"`           // 建议的具体内容
	CreatedAt    time.Time `json:"created_at"`        // 建议创建时间
	UpdatedAt    time.Time `json:"updated_at"`        // 建议更新时间
}
