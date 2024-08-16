package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

func VerifyEmail(c *gin.Context) {
	//email := c.Query("email")
	token := c.Query("token")

	var verification EmailVerification

	if err := db.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&verification).Error; err != nil {
		// 说明校验失败，删除Valid为false的用户
		if verification.UserID != 0 {
			db.DB.Where("valid = ? and user_id=?", false, verification.UserID).Delete(&User{})
		}
		lib.Err(c, http.StatusBadRequest, "无效或过期的验证令牌", err)
		return
	}
	var user User
	if err := db.DB.Where("id = ?", verification.UserID).First(&user).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "用户不存在", err)
		return
	}
	user.Valid = true

	if err := db.DB.Save(&user).Error; err != nil {
		lib.Err(c, http.StatusInternalServerError, "无法更新用户验证状态", err)
		return
	}

	// 删除或失效验证令牌
	db.DB.Delete(&verification)

	lib.Ok(c, "邮箱验证成功，您的账号已可以使用")
}
