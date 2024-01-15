/****************************************************************************
 * Copyright 2023 Optimizely, Inc. and contributors                    		*
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package tracing //
package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides the necessary method to collect telemetry trace data.
// Tracer should not depend on any specific tool. To make it possible it returns a Span interface.
type Tracer interface {
	// StartSpan starts a trace span. Span can be a parent or child span based on the passed context.
	StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, Span)
}

// Span interface implements the trace span returned by Tracer.
type Span interface {
	End()
	SetAttibutes(key string, value interface{})
}

// otelTracer is an OpenTelemetry implementation of Tracer
type otelTracer struct {
	enabled bool
}

// NewOtelTracer returns a new instance of Tracer
func NewOtelTracer(t trace.Tracer) Tracer {
	return &otelTracer{
		enabled: true,
	}
}

// StartSpan starts a trace span. Span can be a parent or child span based on the passed context.
func (t *otelTracer) StartSpan(pctx context.Context, tracerName, spanName string) (context.Context, Span) {
	ctx, span := otel.Tracer(tracerName).Start(pctx, spanName)
	return ctx, &otelSpan{
		span: span,
	}
}

// otelSpan is an OpenTelemetry Span implementation of Span
type otelSpan struct {
	span trace.Span
}

// SetAttibutes sets the attributes for the span
func (s *otelSpan) SetAttibutes(key string, value interface{}) {
	s.span.SetAttributes(attribute.KeyValue{
		Key:   attribute.Key(key),
		Value: attribute.StringValue(value.(string)),
	})
}

// End ends the span
func (s *otelSpan) End() {
	s.span.End()
}

// NoopTracer is a no-op implementation of Tracer
type NoopTracer struct{}

// StartSpan returns a new instance of NoopTracer
func (t *NoopTracer) StartSpan(ctx context.Context, tracerName, spanName string) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

// NoopSpan is a no-op implementation of Span
type NoopSpan struct{}

// SetAttibutes sets the attributes for the noop-span
func (s *NoopSpan) SetAttibutes(key string, value interface{}) {}

// End ends the noop-span
func (s *NoopSpan) End() {}
