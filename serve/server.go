package serve

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/miekg/dns"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/shovel/digcombine"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// splitFormValue splits by " " and removes "" from output slice
func splitFormValue(formValue string) []string {
	vals := strings.Split(formValue, " ")

	ret := []string{}
	for _, v := range vals {
		if v != "" {
			ret = append(ret, v)
		}
	}
	return ret
}

type server struct {

	// Motd - message of the day
	Motd   template.HTML
	Footer template.HTML

	// Version of our software
	Version string

	Tracer trace.Tracer
}

func (s *server) Submit(c echo.Context) error {

	ctx, cancel := context.WithTimeout(c.Request().Context(), 30*time.Second)
	defer cancel()

	countForm := c.FormValue("count")
	qnames := splitFormValue(c.FormValue("qnames"))
	nameservers := splitFormValue(c.FormValue("nameservers"))
	proto := c.FormValue("protocol")
	rtypeStrs := splitFormValue(c.FormValue("rtypes"))
	subnetMapStrs := splitFormValue(c.FormValue("subnetMap"))
	subnets := splitFormValue(c.FormValue("subnets"))

	formErrors := []error{}

	if proto != "udp" && proto != "tcp" && proto != "tcp-tls" {
		formErrors = append(formErrors, errors.New("unsupported proto (should be one of udp, tcp, tcp-tls): "+proto))
	}

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
		return c.Render(http.StatusOK, "submiterror.html", formErrors)
	}

	params := dig.CombineDigRepeatParams(
		nameservers,
		proto,
		qnames,
		rtypes,
		parsedSubnets,
		count,
	)

	// resMul := dig.DigRepeatParallel(ctx, params, dig.DigOne)
	resMul := dig.DigRepeatParallel(ctx, params, s.DigOne)

	// This only works for GET
	// filledFormURL := s.HTTPOrigin + "/?" + c.Request().URL.RawQuery

	// TODO: don't hardcode the http protocol...
	httpOrigin := "http://" + c.Request().Host

	// This works for POST and I think it works for GET too?
	filledFormURL := httpOrigin + "/?" + c.Request().Form.Encode()

	traceID := trace.SpanContextFromContext(
		c.Request().Context(),
	).TraceID().String()

	tableYAMLStr, err := buildTableYAML(params, resMul)
	if err != nil {
		// TODO: non-fatal error, send to traces and continue...
		panic(err)
	}

	t := ResultTable{
		FilledFormURL: filledFormURL,
		Rows: buildRows(buildRowParams{
			Qnames:       qnames,
			RtypeStrs:    rtypeStrs,
			Subnets:      parsedSubnets,
			Nameservers:  nameservers,
			ResMul:       resMul,
			SubnetToName: subnetToName,
		}),
		TraceIDTemplateArgs: TraceIDTemplateArgs{TraceID: traceID},
		TableYAML:           tableYAMLStr,
	}

	return c.Render(http.StatusOK, "submit.html", t)
}

func (s *server) Index(c echo.Context) error {

	type indexData struct {
		Count       string
		Qnames      string
		Nameservers string
		Proto       string
		Rtypes      string
		SubnetMap   string
		Subnets     string

		Footer     template.HTML
		Motd       template.HTML
		Version    string
		VersionURL string
	}

	f := indexData{
		Count:       c.FormValue("count"),
		Qnames:      c.FormValue("qnames"),
		Nameservers: c.FormValue("nameservers"),
		Proto:       c.FormValue("protocol"),
		Rtypes:      c.FormValue("rtypes"),
		SubnetMap:   c.FormValue("subnetMap"),
		Subnets:     c.FormValue("subnets"),
		Footer:      s.Footer,
		Motd:        s.Motd,
		Version:     s.Version,
		// Might be better not to hardcode this, but I don't see it changing ever...
		VersionURL: "https://github.com/bbkane/shovel",
	}
	return c.Render(http.StatusOK, "index.html", f)
}

func (s *server) DigOne(ctx context.Context, p dig.DigOneParams) ([]string, error) {
	// https://opentelemetry.io/docs/concepts/signals/traces/#spans
	ctx, span := s.Tracer.Start(
		ctx,
		"DigOne",
		trace.WithAttributes(
			attribute.String("NameserverIPPort", p.NameserverIPPort),
			attribute.String("Proto", p.Proto),
			attribute.String("Qname", p.Qname),
			attribute.String("Rtype", dns.TypeToString[p.Rtype]),
			attribute.String("SubnetIP", p.SubnetIP.String()),
			attribute.Int64("Timeout", int64(p.Timeout)),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()
	results, err := dig.DigOne(ctx, p)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	return results, err
}
