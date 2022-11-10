package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(config TracingConfig) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.ExportUrl)))

	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()), //TODO: NEED TO CHANGE FOR PRODUCTION
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.Attributes.service),
			attribute.String("environment", config.Attributes.environment),
			attribute.String("ID", config.Attributes.id),
		)),
	)

	return tp, nil
}

// var tracer = otel.Tracer("gin-server")
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/jaeger/main.go
func initTracer(config TracingConfig) (*tracesdk.TracerProvider, error) {

	tp, err := tracerProvider(config)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func RegisterOpenTelemetry(ctx context.Context, cfg config.ServiceConfig, log logging.Logger) error {
	telemetryConfig, err := ConfigGetter(cfg.ExtraConfig)
	if err != nil {
		return ErrNoConfig
	}

	otelConfiguration := telemetryConfig.(TelemetryConfig).Tracing

	tp, err := initTracer(otelConfiguration)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error("Error shutting down tracer provider: %v", err)
		}
		cancel()
	}()

	return nil
}
