package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"massage,omitempty"`
}

func ResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 处理请求

		// 获取处理结果的状态码
		status := c.Writer.Status()

		// 如果响应已经写入，直接返回
		if c.Writer.Written() {
			return
		}

		// 获取原始响应数据
		var data interface{}
		if c.Keys != nil {
			data = c.Keys["data"]

		}
		msg := c.Keys["massage"]

		if status == 404 {
			msg = "Not Found"
		}

		// 构建统一响应结构体
		response := Response{
			Code: status,
			Data: data,
			Msg:  fmt.Sprintf("%v", msg),
		}

		// 以 JSON 形式返回响应
		c.JSON(status, response)
	}
}
