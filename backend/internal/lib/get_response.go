package lib

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Code(c *gin.Context, code int) {
	// 设置 HTTP 状态码
	c.Status(code)
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
		Msg(c, input[0].(string))
		Data(c, nil)
	}
}

// 一个参数省略msg
func Registered(c *gin.Context, input ...interface{}) {
	if len(input) >= 3 {
		log.Println("too many parameters")
		Err(c, 500, "参数过多，请后端开发人员排查", nil)
	}
	Code(c, 201)
	if len(input) == 2 {
		Msg(c, input[0].(string))
		Data(c, input[1])
	} else {
		Msg(c, input[0].(string))
		Data(c, nil)
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

	var errorMsg string
	if err != nil {
		errorMsg = err.Error()
	} else {
		errorMsg = "Unknown error"
	}

	Data(c, gin.H{
		"error": errorMsg,
	})
}
