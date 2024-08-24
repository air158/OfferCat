package redeem

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/auth/model"
	"offercat/v0/internal/db"
	"offercat/v0/internal/lib"
	"time"
)

func RedeemMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if lib.GetRole(c) == "admin" {
			c.Next()
			return
		}

		uid := lib.Uid(c)
		if uid == 0 {
			lib.Err(c, 401, "未登录", nil)
			c.Abort()
			return
		}
		var user model.User
		if err := db.DB.Where("id = ?", uid).First(&user).Error; err != nil {
			lib.Err(c, 500, "数据库错误", err)
			c.Abort()
			return
		}
		if user.VipExpireAt.Before(time.Now()) {
			if user.InterviewPoint > 0 {
				//说明面试点数够,可以继续使用

				// 上下文中记录用户使用面试点数
				c.Set("cost_type", "interview_point")
				c.Next()
			} else {
				lib.Err(c, 401, "没有vip且面试点数不够，请尝试获取兑换码", nil)
				c.Abort()
				return
			}
		} else {
			// 说明是vip
			// 上下文中记录用户使用vip
			c.Set("cost_type", "vip")
			c.Next()
		}
	}
}
