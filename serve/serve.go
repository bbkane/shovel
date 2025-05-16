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
	"go.bbkane.com/warg/path"
	"go.bbkane.com/warg/wargcore"
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

func Run(cmdCtx wargcore.Context) error {

	addrPort := cmdCtx.Flags["--addr-port"].(netip.AddrPort).String()
	motd, _ := cmdCtx.Flags["--motd"].(string)
	footer, _ := cmdCtx.Flags["--footer"].(string)

	otelProvider := cmdCtx.Flags["--otel-provider"].(string)
	protocol := cmdCtx.Flags["--protocol"].(string)

	var certFile string
	var keyFile string
	switch protocol {
	case "HTTP":
		// do nothing, we're fine
	case "HTTPS":
		for _, name := range []string{"--https-certfile", "--https-keyfile"} {
			if _, exists := cmdCtx.Flags[name]; !exists {
				return errors.New("--protocol HTTPS requires this flag to be set: " + name)
			}
		}
		certFile = cmdCtx.Flags["--https-certfile"].(path.Path).MustExpand()
		keyFile = cmdCtx.Flags["--https-keyfile"].(path.Path).MustExpand()
	default:
		panic("Unknown serve protocol: " + protocol)
	}

	traceIDTemplate := cmdCtx.Flags["--trace-id-template"].(string)

	var tp *sdktrace.TracerProvider
	var tpErr error

	serviceName := []string{cmdCtx.App.Name}
	serviceName = append(serviceName, cmdCtx.ParseState.SectionPath...)
	serviceName = append(serviceName, cmdCtx.ParseState.CurrentCommandName)
	tracerArgs := tracerResourceArgs{
		ServiceName:        strings.Join(serviceName, " "),
		ServiceVersion:     cmdCtx.App.Version,
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

	// Add templates
	{
		fsTemplates, err := template.New("").
			Funcs(template.FuncMap{}).
			ParseFS(embeddedFiles, "static/templates/*.html")
		if err != nil {
			return fmt.Errorf("could not parse embedded template files: %w", err)
		}

		traceIDFlagTemplate, err := fsTemplates.New("trace-id-template").Parse(traceIDTemplate)
		if err != nil {
			return fmt.Errorf("could not parse --trace-id-template string: %w", err)

		}
		t := &Template{
			templates: traceIDFlagTemplate,
		}
		e.Renderer = t
	}

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
		Motd:    template.HTML(motd),
		Version: cmdCtx.App.Version,
		Tracer: tp.Tracer(
			"shovel serve", // TODO: get a better name
		),
		Footer: template.HTML(footer),
	}

	addRoutes(e, s)

	// Start server and shutdown gracefully. https://echo.labstack.com/cookbook/graceful-shutdown/
	go func() {

		var startFunc func() error
		switch protocol {
		case "HTTP":
			startFunc = func() error { return e.Start(addrPort) }
		case "HTTPS":
			startFunc = func() error {
				return e.StartTLS(addrPort, certFile, keyFile)
			}
		default:
			panic("Unknown serve protocol")
		}

		if err := startFunc(); err != nil && err != http.ErrServerClosed {
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
