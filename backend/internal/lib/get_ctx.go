package lib

import (
	"github.com/gin-gonic/gin"
)

func Uid(c *gin.Context) int {
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

func Code(c *gin.Context, code int) {
	c.Set("code", code)
}

func Msg(c *gin.Context, msg string) {
	c.Set("msg", msg)
}

func Data(c *gin.Context, data interface{}) {
	c.Set("data", data)
}

// 一个参数省略msg
func Ok(c *gin.Context, input ...interface{}) {
	Code(c, 200)
	if len(input) == 1 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, "success")
		Data(c, input[0])
	}
}

// 一个参数省略msg。
func Fail(c *gin.Context, input ...interface{}) {
	Code(c, 400)
	if len(input) == 1 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, "fail")
		Data(c, input[0])
	}
}
func Err(code int, msg string, err error) gin.HandlerFunc {
	return func(c *gin.Context) {
		Code(c, code)
		Msg(c, msg)
		Data(c, gin.H{
			"error": err.Error(),
		})
	}
}
