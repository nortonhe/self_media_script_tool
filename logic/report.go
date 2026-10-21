package logic

import (
	"context"
	"self_medie_script_tool/observability"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type llmCallReporter struct {
	ctx     context.Context
	metrics *observability.LLMMetrics
	span    trace.Span
	attrs   []attribute.KeyValue
	start   time.Time
	usage   *openai.Usage
}

func newLLMCallReporter(
	ctx context.Context,
	span trace.Span,
	metrics *observability.LLMMetrics,
	model string,
	start time.Time,
) *llmCallReporter {
	attrs := []attribute.KeyValue{
		attribute.String("template_id", "short_video_sales"),
		attribute.String("template_version", "v1"),
		attribute.String("model", model),
	}
	span.SetAttributes(attrs...)
	return &llmCallReporter{
		ctx:     ctx,
		metrics: metrics,
		span:    span,
		attrs:   attrs,
		start:   start,
	}
}

func (r *llmCallReporter) SetUsage(usage *openai.Usage) {
	r.usage = usage
}

func (r *llmCallReporter) Record(err error) {
	status := "success"
	if err != nil {
		status = "error"
		r.span.SetStatus(codes.Error, err.Error())
		r.span.RecordError(err)
	}

	if r.metrics == nil {
		return
	}

	requestAttrs := append(append([]attribute.KeyValue(nil), r.attrs...), attribute.String("status", status))
	r.metrics.RequestTotal.Add(r.ctx, 1, metric.WithAttributes(requestAttrs...))
	r.metrics.Duration.Record(r.ctx, time.Since(r.start).Seconds(), metric.WithAttributes(requestAttrs...))

	if r.usage != nil {
		r.metrics.InputToken.Add(r.ctx, int64(r.usage.PromptTokens), metric.WithAttributes(r.attrs...))
		r.metrics.OutputToken.Add(r.ctx, int64(r.usage.CompletionTokens), metric.WithAttributes(r.attrs...))
	}
}