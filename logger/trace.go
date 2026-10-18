package logger

import (
	"context"

	"github.com/google/uuid"
)

// 定义ctx key，不要用string直接当key，防止冲突
type ctxKey struct{}

// TraceIDKey 存traceId到context的key
var TraceIDKey = &ctxKey{}

// GetTraceId 从ctx获取traceId
func GetTraceId(ctx context.Context) string {
	val, ok := ctx.Value(TraceIDKey).(string)
	if !ok {
		return ""
	}
	return val
}

// SetTraceId 把traceId写入ctx，返回新的ctx
func SetTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceId)
}

// NewTraceId 生成新traceId（uuid v4）
func NewTraceId() string {
	return uuid.NewString()
}
