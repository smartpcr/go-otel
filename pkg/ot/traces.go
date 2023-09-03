package ot

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	"k8s.io/component-base/version"
	"os"
	"time"
)

func RegisterTracing(ctx context.Context, endpoint, serviceName string, log *logrus.Entry) error {
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

func newTracerProvider(ctx context.Context, endpoint, serviceName string) (*trace.TracerProvider, error) {
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

	bsp := trace.NewBatchSpanProcessor(exporter)
	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{}) // W3C Trace Context format

	return traceProvider, nil
}
