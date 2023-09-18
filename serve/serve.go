package serve

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/netip"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.bbkane.com/warg/command"
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

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(LogReqMiddleware())

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

	s := &server{
		HTTPOrigin: httpOrigin,
		Motd:       motd,
	}

	addRoutes(e, s)

	e.Logger.Fatal(e.Start(addrPort))
	return nil
}
