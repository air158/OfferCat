package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/db"
	"time"
)

func VerifyEmail(c *gin.Context) {
	//email := c.Query("email")
	token := c.Query("token")

	var verification EmailVerification
	if err := db.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&verification).Error; err != nil {
		// 说明校验失败，删除Valid为false的用户
		db.DB.Where("valid = ?", false).Delete(&User{})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}
	var user User
	if err := db.DB.Where("id = ?", verification.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	user.Valid = true

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user verification status"})
		return
	}

	// 删除或失效验证令牌
	db.DB.Delete(&verification)

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
