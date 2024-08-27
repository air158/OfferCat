package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"offercat/v0/internal/auth/jwt"
	"offercat/v0/internal/lib"
)

// JWTAuthMiddleware 是一个Gin中间件，用于验证JWT token
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			lib.Err(c, http.StatusUnauthorized, "需要提供鉴权token", nil)
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(token)
		if err != nil {
			lib.Err(c, http.StatusUnauthorized, "无效的或过期的token", err)
			c.Abort()
			return
		}

		// 将用户信息存储在上下文中，以便后续处理使用
		c.Set("uid", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
