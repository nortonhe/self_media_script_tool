package logger

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GinZap 自定义请求日志中间件，使用项目zap日志
func GinZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// 执行后续handler
		c.Next()
		// 处理完请求后记录日志
		latency := time.Since(start)
		ZapLogger.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", strings.TrimSpace(c.PostForm("query"))),
			zap.String("clientIP", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("final_response", c.GetString("final_response")),
		)
	}
}

func RecoveryWithZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic到zap日志
				ZapLogger.Error("http panic recovered",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

