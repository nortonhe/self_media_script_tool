package middleware

import (
	"self_medie_script_tool/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
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
		logger.CtxInfo(c.Request.Context(), "http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("clientIP", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
		)
	}
}

// RecoveryWithZap 自定义panic恢复中间件，使用项目zap日志
func RecoveryWithZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic到zap日志
				logger.CtxError(c.Request.Context(), "http panic recovered",
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

// HeaderTraceId 定义traceId在http header中的key
const HeaderTraceId = "X-Trace-Id"

// TraceMiddleware 链路追踪中间件
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var traceID string
		// 优先从span中获取，对齐trace链路
		span := trace.SpanFromContext(c)
		if span.IsRecording() {
			sc := span.SpanContext()
			traceID = sc.TraceID().String()
		}
	
		// 从header获取上游传递的traceId，没有则新建
		if traceID == "" {
			traceID = c.GetHeader(HeaderTraceId)
		}
		
		// 通过uuid自建
		if traceID == "" {
			traceID = logger.NewTraceId()
		}

		// 把traceId存入gin的context（gin.Context内嵌context.Context）
		newCtx := logger.SetTraceId(c.Request.Context(), traceID)
		// 替换请求底层的context
		c.Request = c.Request.WithContext(newCtx)

		// 可选：把traceId回写到响应header，方便前端排查
		c.Header(HeaderTraceId, traceID)

		c.Next()
	}
}
