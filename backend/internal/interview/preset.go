package interview

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"offercat/v0/internal/resume"
	"offercat/v0/internal/store"
	"offercat/v0/internal/thirdparty/llm"
	pdf_analyser "offercat/v0/internal/thirdparty/pdf-analyser"
	"reflect"
	"time"
)

// Preset 面试前的预设，包括岗位信息、简历和偏好设置
type Preset struct {
	ID               int    `json:"id" gorm:"primaryKey"`
	UserID           int    `json:"-"`
	JobTitle         string `json:"job_title"`
	JobDescription   string `json:"job_description"`
	Company          string `json:"company"`
	Business         string `json:"business"`
	Location         string `json:"location"`
	Progress         string `json:"progress"`  // 第几面
	ResumeID         string `json:"resume_id"` // 简历ID，会另外上传
	Language         string `json:"language"`
	InterviewerStyle string `json:"interviewer_style"`
	AnswerLength     int    `json:"answer_length"`
}

// UpsertPreset 更新用户的预设信息
func UpsertPreset(c *gin.Context) {
	var inputPreset Preset
	if err := c.ShouldBindJSON(&inputPreset); err != nil {
		lib.Err(c, http.StatusBadRequest, "解析用户预设信息失败，可能是不合法的输入", err)
		return
	}
	uidInt := lib.Uid(c)

	var existingPreset Preset
	// 在数据库中查找该用户的预设信息
	if err := db.DB.Where("user_id = ?", uidInt).First(&existingPreset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户没有现有的预设信息，创建新的
			existingPreset = Preset{
				UserID: uidInt,
			}
		} else {
			// 其他错误，返回错误响应
			lib.Err(c, http.StatusInternalServerError, "查询预设信息失败", err)
			return
		}
	}

	// 使用反射自动更新非零值的字段
	inputValue := reflect.ValueOf(&inputPreset).Elem()
	existingValue := reflect.ValueOf(&existingPreset).Elem()

	for i := 0; i < inputValue.NumField(); i++ {
		inputField := inputValue.Field(i)
		if !inputField.IsZero() {
			existingValue.Field(i).Set(inputField)
		}
	}

	// 保存或更新预设信息
	if err := db.DB.Save(&existingPreset).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "保存预设信息失败", err)
		return
	}

	lib.Ok(c, "保存预设信息成功", existingPreset)
}

// GetPreset 获取用户的预设信息
func GetPreset(c *gin.Context) {
	uidInt := lib.Uid(c)

	var preset Preset
	if err := db.DB.Where("user_id = ?", uidInt).First(&preset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lib.Err(c, http.StatusNotFound, "预设信息未找到", err)
			return
		}
		lib.Err(c, http.StatusInternalServerError, "查询预设信息失败", err)
		return
	}

	lib.Ok(c, "获取预设信息成功", preset)
}

func UploadResumePDF(c *gin.Context) {
	// 上传简历PDF文件
	file, err := c.FormFile("resume_file")
	if err != nil {
		log.Printf("Error in FormFile: %v", err)
		lib.Err(c, http.StatusBadRequest, "上传文件失败", err)
		return
	}
	log.Println("FormFile uploaded successfully")

	// 打开上传的文件
	srcFile, err := file.Open()
	if err != nil {
		log.Printf("Error opening uploaded file: %v", err)
		lib.Err(c, http.StatusInternalServerError, "打开上传文件失败", err)
		return
	}
	defer func(srcFile multipart.File) {
		err := srcFile.Close()
		if err != nil {
			log.Printf("Error closing uploaded file: %v", err)
		}
	}(srcFile)
	log.Println("Uploaded file opened successfully")

	uid := lib.Uid(c)
	log.Printf("User ID: %d", uid)

	var resumeEntity resume.Resume
	resumeEntity.UploadedAt = time.Now()
	resumeEntity.UserID = uint(uid)
	resumeEntity.FileName = file.Filename
	resumeEntity.Status = "待审核"

	endpoint, minioClient, done := store.MinioInit(c, err)
	if done {
		return
	}

	bucketName := "offercat-resume"                        // 替换为你要上传到的桶名称
	objectName := fmt.Sprintf("%d/%s", uid, file.Filename) // 文件在MinIO中的路径
	log.Printf("Object name in MinIO: %s", objectName)

	// 上传文件到MinIO
	_, err = minioClient.PutObject(c, bucketName, objectName, srcFile, file.Size, minio.PutObjectOptions{ContentType: "application/pdf"})
	if err != nil {
		log.Printf("Error uploading file to MinIO: %v", err)
		lib.Err(c, http.StatusInternalServerError, "上传文件到MinIO失败", err)
		return
	}
	log.Println("File uploaded to MinIO successfully")

	resumeEntity.FilePath = fmt.Sprintf("%s/%s/%s", endpoint, bucketName, objectName)
	log.Printf("Resume file path: %s", resumeEntity.FilePath)

	// 添加resume到数据库
	err = resume.NewResume(&resumeEntity)
	if err != nil {
		log.Printf("Error adding resume to database: %v", err)
		lib.Err(c, http.StatusInternalServerError, "添加简历到数据库失败", err)
		return
	}
	log.Println("Resume added to database successfully")

	// 准备更新的Preset数据
	presetUpdate := Preset{
		ResumeID: objectName,
	}
	log.Println("Preset update data prepared")

	// 将Preset数据转换为JSON
	presetJson, err := json.Marshal(presetUpdate)
	if err != nil {
		log.Printf("Error encoding preset data: %v", err)
		lib.Err(c, http.StatusInternalServerError, "编码预设数据失败", err)
		return
	}
	log.Println("Preset data encoded to JSON successfully")

	// 将JSON数据设置到请求体中，使其可被ShouldBindJSON解析
	c.Request.Body = io.NopCloser(bytes.NewBuffer(presetJson))
	log.Println("Request body set with preset JSON")

	// 调用UpsertPreset函数来更新或插入Preset记录
	UpsertPreset(c)

	if c.Writer.Status() == http.StatusOK {
		log.Println("Preset updated successfully, sending response")
		lib.Ok(c, "文件上传成功", gin.H{
			"url": fmt.Sprintf("%s/%s/%s", endpoint, bucketName, objectName),
		})
	} else {
		log.Println("Failed to update preset")
	}
}

type ResumeSuggestionRequest struct {
	ResumeID uint `json:"resume_id" form:"resume_id"`
}

// 简历建议和评价
func ResumeSuggestion(c *gin.Context) {
	uid := lib.Uid(c)
	var err error
	var req ResumeSuggestionRequest
	var resumeEntity *resume.Resume
	err = c.ShouldBindQuery(&req)
	if err != nil || req.ResumeID == 0 {
		lib.Err(c, http.StatusBadRequest, "解析请求失败，可能是不合法的输入", err)
		log.Println(req.ResumeID)
		return
	}

	// 数据库中查找简历
	var resumeList []resume.Resume

	resumeList, err = resume.GetResumeListByUserID(uint(uid))
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "查询建立信息失败", err)
		return
	}
	for _, r := range resumeList {
		if r.ID == req.ResumeID {
			resumeEntity = &r
			break
		}
	}
	if resumeEntity == nil {
		lib.Err(c, http.StatusNotFound, "简历未找到", nil)
		return
	}
	path := resumeEntity.FilePath
	// 从MinIO下载简历文件
	stringFromPDF := pdf_analyser.GetStringFromPDF(c, path)
	prompt := "上面是一份简历，请为这份简历提供建议"
	// 调用Spark API
	response, err := llm.CallSparkAPI(stringFromPDF + prompt)
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "调用Spark API失败", err)
		return
	}
	lib.Ok(c, "获取简历建议成功", gin.H{
		"suggestion": response,
	})
}
