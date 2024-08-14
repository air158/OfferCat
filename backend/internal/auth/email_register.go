package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/smtp"
	"offercat/v0/internal/db"
	"regexp"
	"time"
)

// EmailVerification 存储邮箱验证信息
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`                               // 关联的用户ID
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"` // 唯一的验证令牌
	CreatedAt time.Time `gorm:"autoCreateTime"`                         // 创建时间
	ExpiresAt time.Time `gorm:"not null"`                               // 令牌过期时间
}
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func EmailRegister(c *gin.Context) {
	var registerInput RegisterInput
	if err := c.ShouldBindJSON(&registerInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 校验邮箱格式
	if !isValidEmail(registerInput.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	// 检查邮箱是否已被注册
	var existingUser User
	var newUser User
	result := db.DB.Where("email = ?", registerInput.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// 密码哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	newUser.PasswordHash = string(hashedPassword)

	// 如果时间为空，则默认为当前时间
	if newUser.DateOfBirth.IsZero() {
		newUser.DateOfBirth = time.Now()
	}

	if newUser.LastLogin.IsZero() {
		newUser.LastLogin = time.Now()
	}
	newUser.Valid = false
	newUser.Username = registerInput.Username
	newUser.Email = registerInput.Email

	// 保存用户信息到数据库
	if err := db.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 生成验证令牌并保存到数据库
	token, err := generateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}
	verification := EmailVerification{
		UserID:    newUser.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 令牌24小时后过期
	}
	if err := db.DB.Create(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save verification token"})
		return
	}

	// 发送验证邮件
	err = sendVerificationEmail(newUser.ID, newUser.Email, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful, please check your email to verify your account."})
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

	verificationLink := fmt.Sprintf("http://127.0.0.1:12345/verify?token=%s", token)
	subject := "Verify your email address"
	body := fmt.Sprintf("Please click the following link to verify your email address: %s", verificationLink)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", senderEmail, email, subject, body)

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{email}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}