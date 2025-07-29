package gorote

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	openlog "go.opentelemetry.io/otel/sdk/log"
	openmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	opentrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TelemetryConn(endpoint string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	return conn, nil
}

func TelemetryTrace(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*opentrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	provider := opentrace.NewTracerProvider(opentrace.WithBatcher(exporter), opentrace.WithResource(res))
	otel.SetTracerProvider(provider)
	return provider, nil
}

func TelemetryMetric(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*openmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	provider := openmetric.NewMeterProvider(openmetric.WithReader(openmetric.NewPeriodicReader(exporter)), openmetric.WithResource(res))
	otel.SetMeterProvider(provider)
	return provider, nil
}

func TelemetryLogger(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (*openlog.LoggerProvider, error) {
	loggerExp, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn), otlploggrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	provider := openlog.NewLoggerProvider(openlog.WithProcessor(openlog.NewBatchProcessor(loggerExp)), openlog.WithResource(res))
	global.SetLoggerProvider(provider)
	return provider, nil
}

func TelemetryResource(ctx context.Context, name string, version string) (*resource.Resource, error) {
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String(name),
		semconv.ServiceVersionKey.String(version),
	))
	if err != nil {
		return nil, err
	}
	return res, nil
}
