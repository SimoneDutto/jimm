// Copyright 2026 Canonical.

package telemetry

import (
	"context"
	"fmt"
	"strings"

	jujuTrace "github.com/juju/juju/core/trace"
	"github.com/juju/zaputil/zapctx"
	otlpgrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otlphttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.uber.org/zap"
)

const (
	defaultTraceProtocol = "http/protobuf"
	tracerName           = "github.com/canonical/jimm/v3"
)

// Params holds the OpenTelemetry exporter configuration used to send spans.
type Params struct {
	ServiceName string
	Endpoint    string
	Protocol    string
	SampleRatio *float64
}

// NewTracer returns a Juju-compatible tracer backed by OpenTelemetry when OTLP
// export is configured.
func NewTracer(ctx context.Context, params Params) (jujuTrace.Tracer, func(context.Context) error, error) {
	if !tracingConfigured(params) {
		return jujuTrace.NoopTracer{}, func(context.Context) error { return nil }, nil
	}

	exporter, err := newExporter(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	providerOptions := []sdktrace.TracerProviderOption{sdktrace.WithBatcher(exporter)}
	if params.SampleRatio != nil {
		providerOptions = append(providerOptions, sdktrace.WithSampler(
			sdktrace.ParentBased(sdktrace.TraceIDRatioBased(*params.SampleRatio)),
		))
	}
	if serviceName := strings.TrimSpace(params.ServiceName); serviceName != "" {
		providerOptions = append(providerOptions, sdktrace.WithResource(
			resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)),
		))
	}
	provider := sdktrace.NewTracerProvider(providerOptions...)
	zapctx.Info(ctx, "otlp tracing enabled", zap.String("protocol", traceProtocol(params)))

	return otelTracer{tracer: provider.Tracer(tracerName)}, provider.Shutdown, nil
}

func tracingConfigured(params Params) bool {
	return strings.TrimSpace(params.Endpoint) != ""
}

func traceProtocol(params Params) string {
	protocol := params.Protocol
	if protocol == "" {
		return defaultTraceProtocol
	}
	return strings.ToLower(strings.TrimSpace(protocol))
}

func newExporter(ctx context.Context, params Params) (sdktrace.SpanExporter, error) {
	switch traceProtocol(params) {
	case "grpc":
		return otlpgrpc.New(ctx, grpcEndpointOption(params.Endpoint))
	case "http/protobuf":
		return otlphttp.New(ctx, httpEndpointOption(params.Endpoint))
	default:
		return nil, fmt.Errorf("unsupported OTLP trace protocol %q", traceProtocol(params))
	}
}

func grpcEndpointOption(endpoint string) otlpgrpc.Option {
	if strings.Contains(endpoint, "://") {
		return otlpgrpc.WithEndpointURL(endpoint)
	}
	return otlpgrpc.WithEndpoint(endpoint)
}

func httpEndpointOption(endpoint string) otlphttp.Option {
	if strings.Contains(endpoint, "://") {
		return otlphttp.WithEndpointURL(endpoint)
	}
	return otlphttp.WithEndpoint(endpoint)
}
