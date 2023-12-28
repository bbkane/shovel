package serve

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.bbkane.com/shovel/serve/custommiddleware"
	"go.bbkane.com/warg/command"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// -- filesystem

//go:embed static
var embeddedFiles embed.FS

// -- template stuff

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// I would like to pass errors up to the caller, but the echo v4 framework
	// silently swallows any error this function would return :(

	// an error for template not found is:
	// >> echo: http: panic serving 127.0.0.1:53382: html/template: "submit" is undefined
	err := t.templates.ExecuteTemplate(w, name, data)
	if err != nil {
		panic(err)
	}
	return nil
}

// -- Run

func mustRead(fSys embed.FS, path string) []byte {
	file, err := fSys.ReadFile(path)
	if err != nil {
		panic("oopsies bad fs path: " + err.Error())
	}
	return file
}

func addRoutes(e *echo.Echo, s *server) {
	e.GET(
		"/",
		s.Index,
	)
	e.GET(
		"/static/index.css",
		func(c echo.Context) error {
			blob := mustRead(embeddedFiles, "static/index.css")
			return c.Blob(http.StatusOK, "text/css", blob)
		},
	)
	e.GET(
		"/static/3p/htmx.org@1.9.5/dist/htmx.min.js",
		func(c echo.Context) error {
			blob := mustRead(embeddedFiles, "static/3p/htmx.org@1.9.5/dist/htmx.min.js")
			return c.Blob(http.StatusOK, "application/javascript; charset=UTF-8", blob)
		},
	)
	e.GET(
		"/static/loading-spinner.svg",
		func(c echo.Context) error {
			blob := mustRead(embeddedFiles, "static/loading-spinner.svg")
			return c.Blob(http.StatusOK, "image/svg+xml", blob)
		},
	)

	e.POST(
		"/submit",
		s.Submit,
	)
}

func Run(cmdCtx command.Context) error {

	addrPort := cmdCtx.Flags["--addr-port"].(netip.AddrPort).String()
	httpOrigin := cmdCtx.Flags["--http-origin"].(string)
	motd, _ := cmdCtx.Flags["--motd"].(string)

	otelProvider := cmdCtx.Flags["--otel-provider"].(string)

	var tp *sdktrace.TracerProvider
	var tpErr error

	serviceName := []string{cmdCtx.AppName}
	serviceName = append(serviceName, cmdCtx.Path...)
	tracerArgs := tracerResourceArgs{
		ServiceName:        strings.Join(serviceName, " "),
		ServiceVersion:     cmdCtx.Version,
		ServiceEnvironment: cmdCtx.Flags["--otel-service-env"].(string),
	}

	switch otelProvider {
	case "openobserve":
		for _, name := range []string{"--openobserve-endpoint", "--openobserve-user", "--openobserve-pass"} {
			if _, exists := cmdCtx.Flags[name]; !exists {
				return errors.New("--otel-provider requires this flag to be set: " + name)
			}
		}
		httpTracerParams := initHTTPTracerProviderParams{
			Endpoint: cmdCtx.Flags["--openobserve-endpoint"].(netip.AddrPort),
			User:     cmdCtx.Flags["--openobserve-user"].(string),
			Password: cmdCtx.Flags["--openobserve-pass"].(string),
		}
		tp, tpErr = initHTTPTracerProvider(httpTracerParams, tracerArgs)
	case "stdout":
		tp, tpErr = initStdoutTracerProvider(tracerArgs)
	default:
		return fmt.Errorf("unknown --otel-provider: %v", otelProvider)
	}

	if tpErr != nil {
		return fmt.Errorf("could not init tracerprovider: %w", tpErr)
	}

	e := echo.New()

	// echo customization
	e.HideBanner = true

	// templates
	temp, err := template.New("").
		Funcs(template.FuncMap{}).
		ParseFS(embeddedFiles, "static/templates/*.html")
	if err != nil {
		return fmt.Errorf("could not parse embedded template files: %w", err)
	}
	t := &Template{
		templates: temp,
	}
	e.Renderer = t

	// TODO: add ability to log to file
	// logger
	// e.Logger.SetLevel(log.DEBUG)
	// e.Use(middleware.Logger())
	// e.Use(custommiddleware.LogRequest())

	e.Use(otelecho.Middleware("shovel"))
	e.Use(custommiddleware.TraceID())

	// recover from panics
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		ctx := c.Request().Context()
		trace.SpanFromContext(ctx).RecordError(err)

		e.DefaultHTTPErrorHandler(err, c)
	}

	// routes
	s := &server{
		HTTPOrigin: httpOrigin,
		Motd:       motd,
		Version:    cmdCtx.Version,
		Tracer: tp.Tracer(
			"shovel serve", // TODO: get a better name
		),
	}

	addRoutes(e, s)

	// Start server and shutdown gracefully. https://echo.labstack.com/cookbook/graceful-shutdown/
	go func() {
		if err := e.Start(addrPort); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalj(log.JSON{
				"message": "start error, shutting down",
				"err":     err.Error(),
			})
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
		return err
	}
	e.Logger.Info("echo shutdown complete")

	if err := tp.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Info("traceprovider shutdown complete")

	return nil
}
