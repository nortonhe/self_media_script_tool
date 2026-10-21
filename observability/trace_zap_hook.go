package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TraceField(ctx context.Context) zap.Field {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return zap.Skip()
	}
	sc := span.SpanContext()
	return zap.String("trace_id", sc.TraceID().String())
}
