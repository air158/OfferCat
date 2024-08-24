package redeem

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

const genLength = 32

// RedeemCode 表示激活码的数据模型
type RedeemCode struct {
	ID          uint      `gorm:"primaryKey"`
	Code        string    `gorm:"uniqueIndex;size:64"` // 激活码字符串，唯一
	ValidTo     time.Time // 激活码的有效期结束时间
	UsedCount   uint      `gorm:"default:0"` // 使用次数
	MaxUseCount uint      // 最大使用次数，不能同一个用户多次使用
	UserID      string    // 使用该激活码的用户ID（如果适用）
	CreatedAt   time.Time // 记录创建时间
	UpdatedAt   time.Time // 记录更新时间
	CreatorName string    // 创建者的用户名
	Tag         string    // 激活码的标签
}

// tag有2种类型：
// vip:1d（期限内可以使用并且不消耗面试点数，数字表示有效期天数），
// interviewPoint:1h（面试点数，数字表示有效期小时数）

type CreateCodeRequest struct {
	Tag string `json:"tag" binding:"required"`
}

// generateSecureCode 生成一个安全的随机激活码
func generateSecureCode(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GenerateUniqueCode 生成唯一的激活码
func GenerateUniqueCode() (string, error) {
	var code string
	var err error
	for {
		code, err = generateSecureCode(genLength)
		if err != nil {
			return "", err
		}
		var count int64
		err = db.DB.Model(&RedeemCode{}).Where("code = ?", code).Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			break
		}
	}
	return code, nil
}

func CreateCode(c *gin.Context) {
	var req CreateCodeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		lib.Err(c, 400, "参数错误", err)
		return
	}
	role := lib.GetRole(c)
	if role != "admin" {
		lib.Err(c, 403, "权限不足", nil)
		return
	}
	codeStr, err := GenerateUniqueCode()
	if err != nil {
		lib.Err(c, 500, "生成激活码失败", err)
		return
	}
	code := RedeemCode{
		Code:        codeStr,
		ValidTo:     time.Now().AddDate(0, 0, 2),
		MaxUseCount: 1,
		Tag:         req.Tag,
		CreatorName: lib.GetUsername(c),
	}
	err = db.DB.Create(&code).Error
	if err != nil {
		lib.Err(c, 500, "创建激活码失败", err)
		return
	}
	lib.Ok(c, "创建激活码成功", gin.H{
		"code": code,
	})
}

type CreateBatchCodeRequest struct {
	Count int    `json:"count" binding:"required"`
	Tag   string `json:"tag" binding:"required"`
}

func CreateBatchCode(c *gin.Context) {
	role := lib.GetRole(c)
	if role != "admin" {
		lib.Err(c, 403, "权限不足", nil)
		return
	}
	var req CreateBatchCodeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil || req.Count <= 0 {
		lib.Err(c, 400, "参数错误", err)
		return
	}

	var codes []RedeemCode
	for i := 0; i < req.Count; i++ {
		codeStr, err := GenerateUniqueCode()
		if err != nil {
			lib.Err(c, 500, "生成激活码失败", err)
			return
		}
		code := RedeemCode{
			Code:        codeStr,
			ValidTo:     time.Now().AddDate(0, 0, 2),
			MaxUseCount: 1,
			Tag:         req.Tag,
			CreatorName: lib.GetUsername(c),
		}
		codes = append(codes, code)
	}
	err = db.DB.Create(&codes).Error
	if err != nil {
		lib.Err(c, 500, "创建激活码失败", err)
		return
	}
	lib.Ok(c, "创建激活码成功", gin.H{
		"codes": codes,
	})
}
