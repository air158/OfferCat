package auth

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/lib"
)

func GetProfile(c *gin.Context) {
	uid := lib.Uid(c)
	role := lib.GetRole(c)
	username := lib.GetUsername(c)
	lib.Ok(c, "获取用户信息成功", gin.H{
		"uid":      uid,
		"username": username,
		"role":     role,
	})
}
