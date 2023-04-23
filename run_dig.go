package main

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miekg/dns"
	"go.bbkane.com/warg/command"
)

// getSubnet, either from --subnet-map or directly from subnet.
func getSubnet(subnetMap map[string]netip.Addr, subnet string) (net.IP, string, error) {

	// check in map
	if subAddr, exists := subnetMap[subnet]; exists {
		return subAddr.AsSlice(), subnet, nil
	}

	// try to parse directly
	subIP := net.ParseIP(subnet)
	if subIP == nil {
		return nil, "", fmt.Errorf("Could not parse IP: %s", subnet)
	}
	return subIP, "passed ip", nil
}

// namemaps holds maps of <ip> -> name for nameservers and subnets
// this makes it nicer to print
type nameMaps struct {
	NameserverNames map[string]string
	SubnetNames     map[string]string
}

func validateNameserverStr(nameserverStr string) error {
	// ensure it ends in a port.
	_, after, found := strings.Cut(nameserverStr, ":")
	nsErr := errors.New("passed nameserver must end in :<portnumber>")
	if !found {
		return nsErr
	}
	_, err := strconv.Atoi(after)
	if err != nil {
		return nsErr
	}
	return nil
}

func cmdCtxToDigRepeatParams(cmdCtx command.Context) ([]digRepeatParams, *nameMaps, error) {

	// simple params
	count := cmdCtx.Flags["--count"].(int)
	fqdns := cmdCtx.Flags["--fqdn"].([]string)
	timeout := cmdCtx.Flags["--timeout"].(time.Duration)

	// rtypes
	rtypeStrs := cmdCtx.Flags["--rtype"].([]string)
	var rtypes []uint16
	for _, rtypeStr := range rtypeStrs {
		rtype, ok := dns.StringToType[rtypeStr]
		if !ok {
			return nil, nil, fmt.Errorf("Couldn't parse rtype: %v", rtype)
		}
		rtypes = append(rtypes, rtype)
	}

	// subnets
	var subnets []net.IP
	subnetNames := make(map[string]string)

	subnetMap, _ := cmdCtx.Flags["--subnet-map"].(map[string]netip.Addr)
	subnetStrs, _ := cmdCtx.Flags["--subnet"].([]string)

	for _, subnetStr := range subnetStrs {
		subnetIP, name, err := getSubnet(subnetMap, subnetStr)
		subnetNames[subnetIP.String()] = name
		if err != nil {
			return nil, nil, err
		}
		subnets = append(subnets, subnetIP)
	}

	// If we don't have any subnets, just use a list of one nil subnet :)
	if len(subnets) == 0 {
		subnets = append(subnets, nil)
	}

	// nameservers
	var nameservers []string
	nameserverNames := make(map[string]string)

	// NOTE: if the wrong types are asserted, the resulting map is nil...
	// It would be nice if Go was kind enough to panic...
	nameserverMap, _ := cmdCtx.Flags["--ns-map"].(map[string]string)

	// These might be names Or IP:Port, so let's not use this slice directly
	nameserverStrs := cmdCtx.Flags["--ns"].([]string)

	for _, nameserverStr := range nameserverStrs {
		// check in map
		if nsAddrPort, exists := nameserverMap[nameserverStr]; exists {
			nameservers = append(nameservers, nsAddrPort)
			nameserverNames[nsAddrPort] = nameserverStr
		} else {
			// use directly
			nameservers = append(nameservers, nameserverStr)
			nameserverNames[nameserverStr] = "passed ns:port"
		}
	}
	for _, nameserver := range nameservers {
		err := validateNameserverStr(nameserver)
		if err != nil {
			return nil, nil, fmt.Errorf("error in nameserver: %s: %w", nameserver, err)
		}
	}

	digRepeatParamsSlice := []digRepeatParams{}

	for _, fqdn := range fqdns {
		for _, rtype := range rtypes {
			for _, subnet := range subnets {
				for _, nameserver := range nameservers {
					digRepeatParamsSlice = append(digRepeatParamsSlice, digRepeatParams{
						DigOneParams: digOneParams{
							FQDN:             fqdn,
							Rtype:            rtype,
							NameserverIPPort: nameserver,
							SubnetIP:         subnet,
							Timeout:          timeout,
						},
						Count: count,
					})
				}
			}
		}
	}
	nameMaps := nameMaps{
		NameserverNames: nameserverNames,
		SubnetNames:     subnetNames,
	}
	return digRepeatParamsSlice, &nameMaps, nil
}

func printDigRepeat(t table.Writer, names nameMaps, p digRepeatParams, r digRepeatResult) {

	fmtSubnet := func(subnet net.IP) string {
		if subnet == nil {
			return ""
		}
		subnetStr := subnet.String()
		name := names.SubnetNames[subnetStr]
		return "# " + name + "\n" + subnet.String()
	}

	fmtNS := func(ns string) string {
		name := names.NameserverNames[ns]
		return "# " + name + "\n" + ns
	}

	// answers
	for _, ans := range r.Answers {
		t.AppendRow(table.Row{
			p.DigOneParams.FQDN,
			dns.TypeToString[p.DigOneParams.Rtype],
			fmtSubnet(p.DigOneParams.SubnetIP),
			fmtNS(p.DigOneParams.NameserverIPPort),
			strings.Join(ans.StringSlice, "\n"),
			ans.Count,
		})
	}
	// errors
	for _, err := range r.Errors {
		t.AppendRow(table.Row{
			p.DigOneParams.FQDN,
			dns.TypeToString[p.DigOneParams.Rtype],
			fmtSubnet(p.DigOneParams.SubnetIP),
			fmtNS(p.DigOneParams.NameserverIPPort),
			err.String,
			err.Count,
		})
	}

	t.AppendSeparator()

}

func runDig(cmdCtx command.Context) error {

	ps, names, err := cmdCtxToDigRepeatParams(cmdCtx)
	if err != nil {
		return err
	}

	results := digVaried(ps)

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(os.Stdout)

	columnConfigs := []table.ColumnConfig{
		{Number: 1, AutoMerge: true}, // FQDN
		{Number: 2, AutoMerge: true}, // Rtype
		{Number: 3, AutoMerge: true}, // Subnet
		{Number: 4, AutoMerge: true}, // Nameserver
	}
	// due to the way parsing works, if the first subnet is nil,
	// we can assume the rest are too.
	if len(ps) > 0 && ps[0].DigOneParams.SubnetIP == nil {
		columnConfigs[2].Hidden = true
	}
	t.SetColumnConfigs(columnConfigs)

	t.AppendHeader(table.Row{"FQDN", "Rtype", "Subnet", "Nameserver", "Ans/Err", "Count"})

	for i := 0; i < len(ps); i++ {
		printDigRepeat(t, *names, ps[i], results[i])
	}

	t.Render()

	return nil
}
