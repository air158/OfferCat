package lib

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Code(c *gin.Context, code int) {
	c.Set("code", code)
}

func Msg(c *gin.Context, msg string) {
	c.Set("massage", msg)
}

func Data(c *gin.Context, data interface{}) {
	c.Set("data", data)
}

// 一个参数省略msg
func Ok(c *gin.Context, input ...interface{}) {
	if len(input) >= 3 {
		log.Println("too many parameters")
		Err(c, 500, "参数过多，请后端开发人员排查", nil)
	}
	Code(c, 200)
	if len(input) == 2 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, "success")
		Data(c, input[0])
	}
}

// 一个参数省略msg。
func Fail(c *gin.Context, input ...interface{}) {
	if len(input) >= 3 {
		log.Println("too many parameters")
		Err(c, 500, "参数过多，请后端开发人员排查", nil)
	}
	Code(c, 400)
	if len(input) == 2 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, "fail")
		Data(c, input[0])
	}
}
func Err(c *gin.Context, code int, msg string, err error) {
	Code(c, code)
	Msg(c, msg)
	Data(c, gin.H{
		"error": err.Error(),
	})
}
