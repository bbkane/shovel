package serve

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/netip"
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
	return c.Render(http.StatusOK, "submit.html", res)
}

// -- Run

func Run(cmdCtx command.Context) error {

	addrPort := cmdCtx.Flags["--address"].(netip.AddrPort).String()

	e := echo.New()
	e.HideBanner = true
	e.Logger.SetLevel(log.DEBUG)

	e.Use(middleware.Logger())
	e.Use(LogReqMiddleware())

	temp, err := template.ParseFS(embeddedFiles, "static/templates/*.html")
	if err != nil {
		return fmt.Errorf("could not parse embedded template files: %w", err)
	}
	t := &Template{
		templates: temp,
	}
	e.Renderer = t

	e.GET(
		"/",
		func(c echo.Context) error {
			file, err := embeddedFiles.ReadFile("static/form.html")
			if err != nil {
				panic("oopsies bad fs path: " + err.Error())
			}
			return c.HTMLBlob(http.StatusOK, file)
		},
	)

	e.GET(
		"/submit",
		submit,
	)

	e.Logger.Fatal(e.Start(addrPort))
	return nil
}
