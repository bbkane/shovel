package serve

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/miekg/dns"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg/command"
)

// -- filesystem

//go:embed static
var embeddedFiles embed.FS

func getFileSystem(logger echo.Logger, useDir bool) http.FileSystem {
	var fsys fs.FS
	var mode string
	if useDir {
		fsys = os.DirFS("static")
		mode = "dir"
	} else {
		var err error
		fsys, err = fs.Sub(embeddedFiles, "static")
		if err != nil {
			panic(err)
		}
		mode = "embedded"
	}
	logger.Debugj(log.JSON{
		"message": "static serve mode",
		"mode":    mode,
	})
	return http.FS(fsys)
}

// -- template stuff

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Hello(c echo.Context) error {
	return c.Render(http.StatusOK, "hello", "World")
}

// -- http handlers

func submit(c echo.Context) error {
	countForm := c.FormValue("count")
	qnames := strings.Split(c.FormValue("qnames"), " ")
	nameservers := strings.Split(c.FormValue("nameservers"), " ")
	protocols := strings.Split(c.FormValue("protocols"), " ")
	rtypes := strings.Split(c.FormValue("rtypes"), " ")
	// subnets := strings.Split(c.FormValue("subnets"), " ")
	timeout := c.FormValue("timeout")

	// TODO: validate all of this or else I'mma be panicking!

	count, err := strconv.Atoi(countForm)
	if err != nil {
		panic(err)
	}

	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}

	res := dig.DigRepeat(
		context.Background(),
		dig.DigRepeatParams{
			DigOneParams: dig.DigOneParams{
				Qname:            qnames[0],
				Rtype:            dns.StringToType[rtypes[0]],
				NameserverIPPort: nameservers[0],
				SubnetIP:         nil, // TODO: get from form
				Timeout:          parsedTimeout,
				Proto:            protocols[0],
			},
			Count: count,
		},
		dig.DigOne,
	)

	return c.String(http.StatusOK, fmt.Sprint(res))
}

// -- Run

func Run(cmdCtx command.Context) error {

	serveStaticFrom := cmdCtx.Flags["--serve-static-from"].(string)
	addrPort := cmdCtx.Flags["--address"].(netip.AddrPort).String()

	useDir := serveStaticFrom == "dir" // only tww choices so this is ok...

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(LogReqMiddleware())

	t := &Template{
		// TODO: this still needs to be embedded...
		templates: template.Must(template.ParseGlob("serve/templates/*.html")),
	}
	e.Renderer = t

	assetHandler := http.FileServer(
		getFileSystem(e.Logger, useDir),
	)
	e.GET(
		"/static/*",
		echo.WrapHandler(
			http.StripPrefix("/static/", assetHandler),
		),
	)

	e.GET(
		"/submit",
		submit,
	)

	e.GET("/hello", Hello)

	e.Logger.Fatal(e.Start(addrPort))
	return nil
}
