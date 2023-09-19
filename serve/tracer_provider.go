package serve

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/netip"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type initHTTPTracerProviderParams struct {
	Endpoint           netip.AddrPort
	User               string
	Password           string
	ServiceName        string
	ServiceVersion     string
	ServiceEnvironment string
}

func initHTTPTracerProvider(p initHTTPTracerProviderParams) (*sdktrace.TracerProvider, error) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	otlptracehttp.NewClient()

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(p.User+":"+p.Password))

	otlpHTTPExporter, err := otlptracehttp.New(context.TODO(),
		otlptracehttp.WithInsecure(), // use http & not https
		otlptracehttp.WithEndpoint(p.Endpoint.String()),
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

	// stdExporter, _ := stdouttrace.New(
	// 	stdouttrace.WithWriter(io.Writer(os.Stdout)),
	// 	stdouttrace.WithPrettyPrint(),
	// )

	if err != nil {
		return nil, fmt.Errorf("Error creating HTTP OTLP exporter: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		// the service name used to display traces in backends
		semconv.ServiceNameKey.String(p.ServiceName),
		semconv.ServiceVersionKey.String(p.ServiceVersion),
		attribute.String("environment", p.ServiceEnvironment),
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
