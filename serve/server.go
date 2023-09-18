package serve

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/shovel/digcombine"
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
	// https://developer.mozilla.org/en-US/docs/Glossary/Origin
	HTTPOrigin string

	// Motd - message of the day
	Motd string
}

func (s *server) Submit(c echo.Context) error {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	resMul := dig.DigRepeatParallel(ctx, params, dig.DigOne)

	// This only works for GET
	// filledFormURL := s.HTTPOrigin + "/?" + c.Request().URL.RawQuery

	// This works for POST and I think it works for GET too?
	filledFormURL := s.HTTPOrigin + "/?" + c.Request().Form.Encode()

	t := buildTable(buildTableParams{
		Qnames:        qnames,
		RtypeStrs:     rtypeStrs,
		Subnets:       parsedSubnets,
		Nameservers:   nameservers,
		ResMul:        resMul,
		SubnetToName:  subnetToName,
		FilledFormURL: filledFormURL,
	})

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

		Motd string
	}

	f := indexData{
		Count:       c.FormValue("count"),
		Qnames:      c.FormValue("qnames"),
		Nameservers: c.FormValue("nameservers"),
		Proto:       c.FormValue("protocol"),
		Rtypes:      c.FormValue("rtypes"),
		SubnetMap:   c.FormValue("subnetMap"),
		Subnets:     c.FormValue("subnets"),
		Motd:        s.Motd,
	}
	return c.Render(http.StatusOK, "index.html", f)
}
