package auth

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"offercat/v0/internal/auth/jwt"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

func Login(c *gin.Context) {
	var loginVals struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginVals); err != nil {
		lib.Err(c, http.StatusBadRequest, "不合法的输入", err)
		return
	}

	var user model.User
	if err := db.DB.Where("email = ? and valid = ? ", loginVals.Email, true).First(&user).Error; err != nil {
		lib.Err(c, http.StatusUnauthorized, "用户名或密码错误", err)

		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginVals.Password)); err != nil {
		log.Println(err)
		log.Println(user.PasswordHash)
		log.Println(loginVals.Password)
		lib.Err(c, http.StatusUnauthorized, "用户名或密码错误", err)
		return
	}

	// 生成 JWT Token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		lib.Err(c, http.StatusInternalServerError, "生成token失败", err)
		return
	}
	user.LastLogin = time.Now()
	db.DB.Save(&user)

	lib.Ok(c, "登录成功", gin.H{
		"username": user.Username,
		"token":    token,
	})
}
