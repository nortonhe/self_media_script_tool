package logger

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
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

// ===== 新增：从ctx读取traceId打印日志 =====
func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	traceId := GetTraceId(ctx)
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Info(msg, fields...)
}

func CtxError(ctx context.Context, msg string, fields ...zap.Field) {
	traceId := GetTraceId(ctx)
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Error(msg, fields...)
}

func CtxWarn(ctx context.Context, msg string, fields ...zap.Field) {
	traceId := GetTraceId(ctx)
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Warn(msg, fields...)
}

func CtxDebug(ctx context.Context, msg string, fields ...zap.Field) {
	traceId := GetTraceId(ctx)
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Debug(msg, fields...)
}

// ===== 新增：直接传traceId打印日志 =====
// 为了避免每次都从context中取traceId，提供直接传traceId的日志函数，提高性能和简化代码
func TraceInfo(traceId string, msg string, fields ...zap.Field) {
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Info(msg, fields...)
}

func TraceError(traceId string, msg string, fields ...zap.Field) {
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Error(msg, fields...)
}

func TraceWarn(traceId string, msg string, fields ...zap.Field) {
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Warn(msg, fields...)
}

func TraceDebug(traceId string, msg string, fields ...zap.Field) {
	if traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	ZapLogger.Debug(msg, fields...)
}
