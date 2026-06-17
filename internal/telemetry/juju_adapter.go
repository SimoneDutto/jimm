// Copyright 2026 Canonical.

package telemetry

import (
	"context"

	jujuTrace "github.com/juju/juju/core/trace"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type otelTracer struct {
	tracer oteltrace.Tracer
}

// Start implements jujuTrace.Tracer.
func (t otelTracer) Start(ctx context.Context, name string, options ...jujuTrace.Option) (context.Context, jujuTrace.Span) {
	if t.tracer == nil {
		return ctx, jujuTrace.NoopSpan{}
	}

	ctx = remoteParentFromScope(ctx)

	spanOptions := []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}
	attributes := toOTelAttributes(traceAttributes(options...))
	if len(attributes) != 0 {
		spanOptions = append(spanOptions, oteltrace.WithAttributes(attributes...))
	}

	ctx, span := t.tracer.Start(ctx, name, spanOptions...)
	wrapped := otelSpan{span: span}
	spanContext := span.SpanContext()
	if spanContext.IsValid() {
		ctx = jujuTrace.WithTraceScope(ctx, spanContext.TraceID().String(), spanContext.SpanID().String(), int(spanContext.TraceFlags()))
	}
	ctx = jujuTrace.WithSpan(ctx, wrapped)

	return ctx, wrapped
}

// Enabled implements jujuTrace.Tracer.
func (t otelTracer) Enabled() bool {
	return t.tracer != nil
}

func remoteParentFromScope(ctx context.Context) context.Context {
	if oteltrace.SpanContextFromContext(ctx).IsValid() {
		return ctx
	}

	traceID, spanID, flags, ok := jujuTrace.ScopeFromContext(ctx)
	if !ok {
		return ctx
	}

	otelTraceID, err := oteltrace.TraceIDFromHex(traceID)
	if err != nil {
		return ctx
	}
	otelSpanID, err := oteltrace.SpanIDFromHex(spanID)
	if err != nil {
		return ctx
	}
	if flags < 0 || flags > 255 {
		return ctx
	}

	spanContext := oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
		TraceID:    otelTraceID,
		SpanID:     otelSpanID,
		TraceFlags: oteltrace.TraceFlags(flags),
		Remote:     true,
	})
	if !spanContext.IsValid() {
		return ctx
	}

	return oteltrace.ContextWithRemoteSpanContext(ctx, spanContext)
}

type otelSpan struct {
	span oteltrace.Span
}

func (s otelSpan) Scope() jujuTrace.Scope {
	spanContext := s.span.SpanContext()
	if !spanContext.IsValid() {
		return jujuTrace.NoopScope{}
	}
	return otelScope{spanContext: spanContext}
}

func (s otelSpan) AddEvent(name string, attributes ...jujuTrace.Attribute) {
	attrs := toOTelAttributes(attributes)
	if len(attrs) == 0 {
		s.span.AddEvent(name)
		return
	}
	s.span.AddEvent(name, oteltrace.WithAttributes(attrs...))
}

func (s otelSpan) RecordError(err error, attributes ...jujuTrace.Attribute) {
	if err == nil {
		return
	}
	attrs := toOTelAttributes(attributes)
	if len(attrs) == 0 {
		s.span.RecordError(err)
	} else {
		s.span.RecordError(err, oteltrace.WithAttributes(attrs...))
	}
	s.span.SetStatus(codes.Error, err.Error())
}

func (s otelSpan) End(attributes ...jujuTrace.Attribute) {
	attrs := toOTelAttributes(attributes)
	if len(attrs) != 0 {
		s.span.SetAttributes(attrs...)
	}
	s.span.End()
}

type otelScope struct {
	spanContext oteltrace.SpanContext
}

func (s otelScope) TraceID() string {
	return s.spanContext.TraceID().String()
}

func (s otelScope) SpanID() string {
	return s.spanContext.SpanID().String()
}

func (s otelScope) TraceFlags() int {
	return int(s.spanContext.TraceFlags())
}

func (s otelScope) IsSampled() bool {
	return s.spanContext.TraceFlags().IsSampled()
}
