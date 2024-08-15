package utils

import (
	"github.com/gin-gonic/gin"
	"offercat/v0/internal/lib"
)

func Ping(c *gin.Context) {
	lib.Ok(c, "pong")
}
