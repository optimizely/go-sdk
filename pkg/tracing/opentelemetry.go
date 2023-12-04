package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Span interface {
	End()
	SetAttibutes(key string, value interface{})
}

type otelSpan struct {
	span trace.Span
}

func (s *otelSpan) SetAttibutes(key string, value interface{}) {
	s.span.SetAttributes(attribute.KeyValue{
		Key:   attribute.Key(key),
		Value: attribute.StringValue(value.(string)),
	})
}

func (s *otelSpan) End() {
	s.span.End()
}

type Tracer interface {
	StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, Span)
}

type otelTracer struct {
	enabled bool
}

func NewOtelTracer(t trace.Tracer) Tracer {
	return &otelTracer{
		enabled: true,
	}
}

func (t *otelTracer) StartSpan(pctx context.Context, tracerName, spanName string) (context.Context, Span) {
	ctx, span := otel.Tracer(tracerName).Start(pctx, spanName)
	return ctx, &otelSpan{
		span: span,
	}
}

type NoopTracer struct{}

func (t *NoopTracer) StartSpan(ctx context.Context, tracerName string, spanName string) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

type NoopSpan struct{}

func (s *NoopSpan) SetAttibutes(key string, value interface{}) {}
func (s *NoopSpan) End()                                       {}
