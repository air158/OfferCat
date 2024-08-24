package middleware

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/lib"
)

// AdminMiddleware 这里代表只有admin才能访问的中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if lib.GetRole(c) != "admin" {
			c.Abort()
			lib.Err(c, 401, "需要管理员权限", nil)
			return
		}
		c.Next()
	}
}
