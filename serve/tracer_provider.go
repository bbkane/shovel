package serve

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/netip"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type initHTTPTracerProviderParams struct {
	Endpoint netip.AddrPort
	User     string
	Password string
}

type tracerResourceArgs struct {
	ServiceName        string
	ServiceVersion     string
	ServiceEnvironment string
}

func initHTTPTracerProvider(httpArgs initHTTPTracerProviderParams, tracerArgs tracerResourceArgs) (*sdktrace.TracerProvider, error) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	otlptracehttp.NewClient()

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(httpArgs.User+":"+httpArgs.Password))

	otlpHTTPExporter, err := otlptracehttp.New(
		context.TODO(),
		otlptracehttp.WithInsecure(), // use http & not https
		otlptracehttp.WithEndpoint(httpArgs.Endpoint.String()),
		otlptracehttp.WithURLPath("/api/default/traces"),
		// otlptracehttp.WithHeaders(map[string]string{
		// 	"Authorization": "Basic cm9vdEBleGFtcGxlLmNvbTpDb21wbGV4cGFzcyMxMjMK",
		// }),
		otlptracehttp.WithHeaders(
			map[string]string{
				"Authorization": authHeader,
			},
		),
	)

	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP OTLP exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String(tracerArgs.ServiceName),
		semconv.ServiceVersionKey.String(tracerArgs.ServiceVersion),
		attribute.String("environment", tracerArgs.ServiceEnvironment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(otlpHTTPExporter),
		// sdktrace.WithBatcher(stdExporter),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

func initStdoutTracerProvider(tracerArgs tracerResourceArgs) (*sdktrace.TracerProvider, error) {

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating stdout tracer: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String(tracerArgs.ServiceName),
		semconv.ServiceVersionKey.String(tracerArgs.ServiceVersion),
		attribute.String("environment", tracerArgs.ServiceEnvironment),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)

	return tp, nil

}
