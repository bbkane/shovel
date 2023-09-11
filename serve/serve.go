package serve

import (
	"context"
	"embed"
	"errors"
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

type buildTableParams struct {
	Qnames       []string
	RtypeStrs    []string
	Subnets      []net.IP
	Nameservers  []string
	ResMul       []dig.DigRepeatResult
	SubnetToName map[string]string
}

func buildTable(p buildTableParams) Table {
	// Add params to output table
	qLen := len(p.Qnames)
	rLen := len(p.RtypeStrs)
	sLen := len(p.Subnets)
	nLen := len(p.Nameservers)
	rows := qLen * rLen * sLen * nLen
	t := Table{
		Rows: make([]Row, rows),
	}

	qWidth := rows / qLen
	{
		i := 0
		for r := 0; r < rows; r += qWidth {
			td := TdData{Content: p.Qnames[i%qLen], Rowspan: qWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	rWidth := qWidth / rLen
	{
		i := 0
		for r := 0; r < rows; r += rWidth {
			td := TdData{Content: p.RtypeStrs[i%rLen], Rowspan: rWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	sWidth := rWidth / sLen
	{
		i := 0
		for r := 0; r < rows; r += sWidth {
			parsedSubnetStr := p.Subnets[i%sLen].String()
			content := parsedSubnetStr + " (" + p.SubnetToName[parsedSubnetStr] + ")"

			td := TdData{Content: content, Rowspan: sWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	nWidth := sWidth / nLen
	{
		i := 0
		for r := 0; r < rows; r += nWidth {
			td := TdData{Content: p.Nameservers[i%nLen], Rowspan: nWidth}
			t.Rows[r].Columns = append(t.Rows[r].Columns, td)
			i++
		}
	}

	// Add anserrs to table
	for i, r := range p.ResMul {
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

	return t
}

func submit(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	countForm := c.FormValue("count")
	qnames := strings.Split(c.FormValue("qnames"), " ")
	nameservers := strings.Split(c.FormValue("nameservers"), " ")
	proto := c.FormValue("protocol")
	rtypeStrs := strings.Split(c.FormValue("rtypes"), " ")
	subnetMapStrs := strings.Split(c.FormValue("subnetMap"), " ")
	subnets := strings.Split(c.FormValue("subnets"), " ")

	formErrors := []error{}

	count, err := strconv.Atoi(countForm)
	if err != nil {
		err := fmt.Errorf("error parsing count: %w", err)
		formErrors = append(formErrors, err)
	}

	rtypes, err := digcombine.ConvertRTypes(rtypeStrs)
	if err != nil {
		err := fmt.Errorf("error parsing rtypes: %w", err)
		formErrors = append(formErrors, err)
	}

	subnetMap := make(map[string]net.IP)
	for _, entry := range subnetMapStrs {
		name, subnetStr, found := strings.Cut(entry, "=")
		if !found {
			formErrors = append(formErrors, errors.New("unable to parse subnet entry: "+entry))
			continue
		}
		subnet := net.ParseIP(subnetStr)
		if subnet == nil {
			formErrors = append(formErrors, errors.New("unable to parse subnet in: "+entry))
			continue
		}
		subnetMap[name] = subnet
	}
	parsedSubnets, subnetToName, err := digcombine.ParseSubnets(subnets, subnetMap)
	if err != nil {
		err := fmt.Errorf("error parsing subnets: %w", err)
		formErrors = append(formErrors, err)
	}

	if len(formErrors) > 0 {
		return c.Render(http.StatusOK, "formerror.html", formErrors)
	}

	params := dig.CombineDigRepeatParams(
		nameservers,
		proto,
		qnames,
		rtypes,
		parsedSubnets,
		count,
	)

	resMul := dig.DigRepeatParallel(ctx, params, dig.DigOne)

	t := buildTable(buildTableParams{
		Qnames:       qnames,
		RtypeStrs:    rtypeStrs,
		Subnets:      parsedSubnets,
		Nameservers:  nameservers,
		ResMul:       resMul,
		SubnetToName: subnetToName,
	})

	return c.Render(http.StatusOK, "submit.html", t)
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
		"/static/3p/htmx.org@1.9.5/dist/htmx.min.js",
		func(c echo.Context) error {
			file, err := embeddedFiles.ReadFile("static/3p/htmx.org@1.9.5/dist/htmx.min.js")
			if err != nil {
				panic("oopsies bad fs path: " + err.Error())
			}
			return c.Blob(http.StatusOK, "application/javascript; charset=UTF-8", file)
		},
	)

	e.GET(
		"/submit",
		submit,
	)

	e.Logger.Fatal(e.Start(addrPort))
	return nil
}
