package serve

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/shovel/digcombine"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	countForm := c.FormValue("count")
	qnames := strings.Split(c.FormValue("qnames"), " ")
	nameservers := strings.Split(c.FormValue("nameservers"), " ")
	proto := c.FormValue("protocol")
	rtypeStrs := strings.Split(c.FormValue("rtypes"), " ")
	// subnets := strings.Split(c.FormValue("subnets"), " ")

	// TODO: validate all of this or else I'mma be panicking!

	count, err := strconv.Atoi(countForm)
	if err != nil {
		panic(err)
	}

	rtypes, err := digcombine.ConvertRTypes(rtypeStrs)
	if err != nil {
		panic(err)
	}

	params := dig.CombineDigRepeatParams(
		nameservers,
		proto,
		qnames,
		rtypes,
		[]net.IP{nil}, // TOOD: subnets
		count,
	)

	resMul := dig.DigRepeatParallel(ctx, params, dig.DigOne)

	submitData := struct {
		Params  []dig.DigRepeatParams
		Results []dig.DigRepeatResult
	}{
		Params:  params,
		Results: resMul,
	}

	fmt.Println(submitData)

	// new fake data stuff yay!

	type TdData struct {
		Content string
		Rowspan int
	}
	type AnsErrCount struct {
		AnsErrs []string
		Count   int
	}
	type AnsErrCountSlice struct {
		AnsErrCounts []AnsErrCount
	}

	type Row struct {
		Columns []TdData
		AnsErrs []AnsErrCountSlice
	}
	type Table struct {
		Rows []Row
	}

	table := Table{
		Rows: []Row{
			Row{
				Columns: []TdData{
					TdData{Content: "www.linkedin.com", Rowspan: 2},
					TdData{Content: "A", Rowspan: 2},
					TdData{Content: "dns3.p09.nsone.net:53", Rowspan: 1},
				},
				AnsErrs: []AnsErrCountSlice{},
			},
			Row{
				Columns: []TdData{
					TdData{Content: "ns3-42.azure-dns.org:53", Rowspan: 1},
				},
			},
		},
	}

	return c.Render(http.StatusOK, "submit2.html", table)
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
		"/static/index.css",
		func(c echo.Context) error {
			file, err := embeddedFiles.ReadFile("static/index.css")
			if err != nil {
				panic("oopsies bad fs path: " + err.Error())
			}
			return c.Blob(http.StatusOK, "text/css", file)
		},
	)

	e.GET(
		"/submit",
		submit,
	)

	e.Logger.Fatal(e.Start(addrPort))
	return nil
}
