package ot

import (
	"context"
	"github.com/pkg/errors"
	crprometheus "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	resource2 "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	"google.golang.org/grpc/credentials"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"time"
)

func RegisterPrometheusMetrics() error {
	exporter, err := prometheus.New(
		prometheus.WithRegisterer(metrics.Registry.(*crprometheus.Registry)))
	if err != nil {
		return err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)

	return nil
}

func RegisterOtelMetrics(ctx context.Context, endpoint, serviceName string) (*metric.MeterProvider, error) {
	tlsConfig, err := getTls()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tls config")
	}

	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create opentelemetry metric exporter")
	}

	resource := resource2.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("machine_name", os.Getenv("HOSTNAME")),
	)
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(resource),
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(30*time.Second))),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}
