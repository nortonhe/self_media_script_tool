package observability

import (
	"context"
	"fmt"
	"self_medie_script_tool/config"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer
var Meter metric.Meter

func InitOTEL(ctx context.Context, cfg config.OTelConfig) (shutdown func(), err error) {
	Tracer = otel.Tracer(cfg.ServiceName)
	Meter  = otel.Meter(cfg.ServiceName)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Trace Exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(), // 本地开发关闭TLS，生产环境删掉这个
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRate)),
	)
	otel.SetTracerProvider(tp)

	// Metric Exporter
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, 
			sdkmetric.WithInterval(time.Duration(cfg.MetricsInterval) * time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	shutdown = func() {
		ctxTO, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctxTO)
		_ = mp.Shutdown(ctxTO)
	}
	return shutdown, nil
}

