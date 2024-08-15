package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
)

func Login(c *gin.Context) {
	var loginVals struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginVals); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user User
	if err := db.DB.Where("email = ? and valid = ? ", loginVals.Email, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginVals.Password)); err != nil {
		log.Println(err)
		log.Println(user.PasswordHash)
		log.Println(loginVals.Password)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// 生成 JWT Token
	token, err := GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	lib.Code(c, 200)
	lib.Msg(c, "登录成功")
	lib.Data(c, gin.H{"token": token})
}
