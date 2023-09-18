package serve

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.bbkane.com/shovel/serve/custommiddleware"
	"go.bbkane.com/warg/command"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
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

	// TODO: make these configurable
	// export UPTRACE_DSN=...
	uptrace.ConfigureOpentelemetry(
		uptrace.WithServiceName("shovel"),
		uptrace.WithServiceVersion("v0.0.1"),
		uptrace.WithDeploymentEnvironment("dev"),
	)

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

	// logger
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())
	e.Use(custommiddleware.LogRequest())

	// tracing
	// TODO: connect with logger? Ditch logger and use this?
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
	}

	addRoutes(e, s)

	// // start
	// e.Logger.Fatal(e.Start(addrPort))
	// return nil

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
	e.Logger.Print("echo shutdown complete")

	if err := uptrace.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Print("uptrace shutdown complete")

	return nil
}
