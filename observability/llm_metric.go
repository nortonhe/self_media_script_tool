package observability

import (
	"go.opentelemetry.io/otel/metric"
)

type LLMMetrics struct {
	RequestTotal metric.Int64Counter   // llm总调用次数
	Duration     metric.Float64Histogram // 调用耗时直方图 P50/P95/P99
	InputToken   metric.Int64Counter   // 输入token
	OutputToken  metric.Int64Counter   // 输出token
}

func NewLLMMetrics(meter metric.Meter) (*LLMMetrics, error) {
	reqTotal, err := meter.Int64Counter("llm.request.total", metric.WithDescription("LLM调用总数"))
	if err != nil {
		return nil, err
	}
	durHist, err := meter.Float64Histogram("llm.request.duration_seconds", metric.WithDescription("LLM推理耗时(秒)"))
	if err != nil {
		return nil, err
	}
	inToken, err := meter.Int64Counter("llm.token.input", metric.WithDescription("LLM输入token总量"))
	if err != nil {
		return nil, err
	}
	outToken, err := meter.Int64Counter("llm.token.output", metric.WithDescription("LLM输出token总量"))
	if err != nil {
		return nil, err
	}
	return &LLMMetrics{
		RequestTotal: reqTotal,
		Duration:     durHist,
		InputToken:   inToken,
		OutputToken:  outToken,
	}, nil
}
