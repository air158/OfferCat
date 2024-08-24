package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/smtp"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"regexp"
	"time"
)

// EmailVerification 存储邮箱验证信息
// 为什么是varchar(255)？就是怕token太长，导致数据库存储不下。说实话其实不设置默认也是255，但是加了比没加好
// uniqueIndex的话是primary的阉割版，可以保证token的唯一性。
// autoCreateTime是自动创建时间，按照go应用的时区。
// id一般不是负的，所以用uint，可以节省一半的空间
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`                               // 关联的用户ID
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"` // 唯一的验证令牌
	CreatedAt time.Time `gorm:"autoCreateTime"`                         // 创建时间
	ExpiresAt time.Time `gorm:"not null"`                               // 令牌过期时间
}

// RegisterInput 存储注册信息
// required是必填项，如果没有填写，就会在下面ShouldBindJSON的时候返回400错误（bad request）
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterOutput struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Valid    bool   `json:"valid"`
}

// EmailRegister 处理邮箱注册请求
// 每次请求共享一个上下文Context，指针传递，所以不用担心并发问题
func EmailRegister(c *gin.Context) {
	var registerInput RegisterInput
	// 从请求中解析JSON数据到registerInput结构体
	// ShouldBindJSON必须绑定的是结构体指针
	if err := c.ShouldBindJSON(&registerInput); err != nil {
		lib.Err(c, http.StatusBadRequest, "不合法的输入", nil)
		return
	}

	// 校验邮箱格式
	if !isValidEmail(registerInput.Email) {
		lib.Err(c, http.StatusBadRequest, "不合法的邮箱格式", nil)
		return
	}

	// 检查邮箱是否已被注册
	var existingUser model.User
	result := db.DB.Where("email = ?", registerInput.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		lib.Registered(c, "邮箱已被注册")
		return
	}
	var newUser model.User
	// 密码哈希处理，cost是哈希计算的成本，越高越安全，但是也越耗时，默认为10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerInput.Password), bcrypt.DefaultCost)
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "密码哈希失败", err)
		return
	}
	newUser.PasswordHash = string(hashedPassword)
	newUser.Role = "user" // 默认为普通用户，如果要管理员直接在数据库修改
	newUser.Username = registerInput.Username
	newUser.Email = registerInput.Email
	newUser.DateOfBirth = time.Now()
	newUser.LastLogin = time.Now()
	newUser.VipExpireAt = time.Now() // 不赠送VIP，所以注册时间就是过期时间
	newUser.InterviewPoint = 2 * 60  // 注册送2*60个面试点（分钟），也就是2小时

	// 验证之后才可以true
	newUser.Valid = false

	// 保存用户信息到数据库
	if err := db.DB.Create(&newUser).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "建立用户信息失败", err)
		return
	}

	// 生成验证令牌并保存到数据库
	token, err := generateVerificationToken()
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "生成验证令牌失败", err)
		return
	}
	verification := EmailVerification{
		UserID:    newUser.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 令牌24小时后过期
	}
	if err := db.DB.Create(&verification).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "创建验证信息失败", err)
		return
	}

	// 发送验证邮件
	err = sendVerificationEmail(newUser.ID, newUser.Email, token)
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "发送验证邮件失败", err)
		return
	}

	lib.Ok(c, "请查看邮箱，点击校验链接以继续完成注册", RegisterOutput{
		Email:    newUser.Email,
		Username: newUser.Username,
		Valid:    newUser.Valid,
	})
}

// 验证邮箱格式是否正确
func isValidEmail(email string) bool {
	regex := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
}

// 生成随机的验证令牌
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// 发送验证邮件
func sendVerificationEmail(userID uint, email, token string) error {
	smtpHost := "smtp.qq.com"
	smtpPort := "587"
	senderEmail := "1195396626@qq.com"
	senderPassword := "izsyvpvqegeyjaia"

	var newUser model.User
	err2 := db.DB.Where("id = ?", userID).First(&newUser).Error
	if err2 != nil {
		return fmt.Errorf("failed to get user: %w", err2)
	}
	username := newUser.Username

	verificationLink := fmt.Sprintf("http://116.198.207.159:12345/api/verify?token=%s", token)
	//verificationLink := fmt.Sprintf("http://127.0.0.1:12345/api/verify?token=%s", token)
	subject := "【OfferCat】尊敬的" + username + "，请验证您的邮箱"
	body := fmt.Sprintf("尊敬的%s:\n您的邮箱被用于在OfferCat上注册了一个新账号，请验证您的邮箱。\n请点击以下链接以继续完成注册:\n %s", username, verificationLink)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", senderEmail, email, subject, body)

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{email}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
