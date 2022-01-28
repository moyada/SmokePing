package log

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type GinLogger struct {
	context *gin.Context
}

func New(context *gin.Context) BizLogger {
	return &GinLogger{context: context}
}

func (c *GinLogger) Info(msg string) {
	c.context.String(http.StatusOK, msg)
}

func (c *GinLogger) Error(err error) {
	c.context.Error(err)
}
