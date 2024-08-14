package lib

import "github.com/gin-gonic/gin"

func GetUid(c *gin.Context) int {
	uid := c.MustGet("uid").(uint)
	uidInt := int(uid)
	return uidInt
}

func GetUsername(c *gin.Context) string {
	username := c.MustGet("username").(string)
	return username
}

func GetRole(c *gin.Context) string {
	role := c.MustGet("role").(string)
	return role
}
