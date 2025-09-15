package gorote

import (
	"context"
	"log"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	openlog "go.opentelemetry.io/otel/sdk/log"
	openmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	opentrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TelemetryConn(endpoint string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
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

func TelemetryFiber(ctx context.Context, app *fiber.App, res *resource.Resource, collectorOpenTelemetry string) error {
	conn, err := TelemetryConn(collectorOpenTelemetry)
	if err != nil {
		return err
	}

	cleanupTrace, err := TelemetryTrace(ctx, res, conn)
	if err != nil {
		return err
	}

	cleanupMetric, err := TelemetryMetric(ctx, res, conn)
	if err != nil {
		return err
	}

	cleanupLogger, err := TelemetryLogger(ctx, res, conn)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		cleanupLogger.Shutdown(ctx)
		cleanupMetric.Shutdown(ctx)
		cleanupTrace.Shutdown(ctx)
		conn.Close()
		log.Println("telemetry shutdown")
	}()

	app.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(cleanupTrace),
		otelfiber.WithMeterProvider(cleanupMetric),
	))
	return nil
}
