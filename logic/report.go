package logic

import (
	"context"
	"encoding/json"
	"self_medie_script_tool/observability"
	"self_medie_script_tool/svrcontext"
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
) *llmCallReporter {
	return &llmCallReporter{
		ctx:     ctx,
		metrics: metrics,
		span:    span,
		start:   time.Now(),
	}
}

func (r *llmCallReporter) SetUsage(usage *openai.Usage) {
	r.usage = usage
}

func (r *llmCallReporter) Record(
	err error,
	request openai.ChatCompletionRequest,
	c *svrcontext.ServerContext,
) {
	metricsAttrs := []attribute.KeyValue{
		attribute.String("template_id", "short_video_sales"),
		attribute.String("template_version", "v1"),
		attribute.String("model", request.Model),
	}

	traceAttrs := append([]attribute.KeyValue(nil), metricsAttrs...)
	traceAttrs = append(traceAttrs,
		attribute.Bool("stream", request.Stream),
		attribute.String("trace_id", c.GetTraceID()),
	)
	
	if requestBody, err := json.Marshal(request); err == nil {
		traceAttrs = append(traceAttrs, attribute.String("llm.request.body", string(requestBody)))
	}

	status := "success"
	if err != nil {
		status = "error"
		r.span.SetStatus(codes.Error, err.Error())
		r.span.RecordError(err)
	}

	traceAttrs = append(traceAttrs, attribute.String("status", status))
	traceAttrs = append(traceAttrs, attribute.String("llm.response.body", c.GetResult()))
	r.span.SetAttributes(traceAttrs...)
	if r.metrics == nil {
		return
	}

	metricsAttrs = append(metricsAttrs, attribute.String("status", status))
	r.metrics.RequestTotal.Add(r.ctx, 1, metric.WithAttributes(metricsAttrs...))
	r.metrics.Duration.Record(r.ctx, time.Since(r.start).Seconds(), metric.WithAttributes(metricsAttrs...))
	if r.usage != nil {
		r.metrics.InputToken.Add(r.ctx, int64(r.usage.PromptTokens), metric.WithAttributes(metricsAttrs...))
		r.metrics.OutputToken.Add(r.ctx, int64(r.usage.CompletionTokens), metric.WithAttributes(metricsAttrs...))
	}
}
