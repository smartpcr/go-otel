package ot

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/component-base/version"
	"os"
	"time"
)

const CorrelationIdKey = "x-ms-correlation-request-id"

var ServiceName string

func RegisterTracing(ctx context.Context, endpoint, serviceName string, log *logrus.Entry) error {
	ServiceName = serviceName
	traceProvider, err := newTracerProvider(ctx, endpoint, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to create opentelemetry trace provider")
	}

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := traceProvider.Shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown opentelemetry trace provider: %v", err)
		}
	}()

	return nil
}

func StartSpanLogger(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *logrus.Entry) {
	tracer := &tracer{
		Tracer: otel.Tracer(ServiceName),
	}
	ctx, span := tracer.startSpan(ctx, name, opts...)
	return ctx, span, logrus.WithContext(ctx)
}

func newTracerProvider(ctx context.Context, endpoint, serviceName string) (*sdkTrace.TracerProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("exporter", "otlp"),
			attribute.String("version", version.Get().String()),
			attribute.String("machine_name", os.Getenv("HOSTNAME")),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create opentelemetry resource")
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create opentelemetry trace exporter")
	}

	bsp := sdkTrace.NewBatchSpanProcessor(exporter)
	traceProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithResource(res),
		sdkTrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{}) // W3C Trace Context format

	return traceProvider, nil
}

type tracer struct {
	trace.Tracer
}

// StartSpan starts a span with the given name and returns a context containing the
// created span.
func (t tracer) startSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, corrId := ensureCorrelationId(ctx)
	opts = append(
		opts,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String(CorrelationIdKey, corrId)),
	)
	return t.Tracer.Start(ctx, name, opts...)
}

func ensureCorrelationId(ctx context.Context) (context.Context, string) {
	if corrId := ctx.Value(CorrelationIdKey); corrId != nil {
		if corrIdString, ok := corrId.(string); ok {
			return ctx, corrIdString
		}
	}
	newCorrId := uuid.New().String()
	return context.WithValue(ctx, CorrelationIdKey, newCorrId), newCorrId
}
