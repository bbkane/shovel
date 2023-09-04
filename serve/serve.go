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

	type TdData struct {
		Content string
		Rowspan int
	}
	type AnsErrCount struct {
		AnsErrs []string
		Count   int
	}

	type Row struct {
		Columns      []TdData
		AnsErrCounts []AnsErrCount
	}
	type Table struct {
		Rows []Row
	}

	// Add params to output table
	qLen := len(qnames)
	rLen := len(rtypes)
	nLen := len(nameservers)
	rows := qLen * rLen * nLen
	t := Table{
		Rows: make([]Row, rows),
	}

	qWidth := rows / qLen
	{
		i := 0
		for r := 0; r < rows; r += qWidth {
			td := TdData{Content: qnames[i%qLen], Rowspan: qWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	rWidth := qWidth / rLen
	{
		i := 0
		for r := 0; r < rows; r += rWidth {
			td := TdData{Content: rtypeStrs[i%rLen], Rowspan: rWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	nWidth := rWidth / nLen
	{
		i := 0
		for r := 0; r < rows; r += nWidth {
			td := TdData{Content: nameservers[i%nLen], Rowspan: nWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	// Add anserrs to table
	for i, r := range resMul {
		aecs := []AnsErrCount{}
		for _, a := range r.Answers {
			aecs = append(
				aecs,
				AnsErrCount{AnsErrs: a.StringSlice, Count: a.Count},
			)
		}
		for _, e := range r.Errors {
			aecs = append(
				aecs,
				AnsErrCount{AnsErrs: []string{e.String}, Count: e.Count},
			)
		}
		t.Rows[i].AnsErrCounts = aecs

	}

	return c.Render(http.StatusOK, "submit2.html", t)
}

// -- Run

func Run(cmdCtx command.Context) error {

	addrPort := cmdCtx.Flags["--address"].(netip.AddrPort).String()

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
