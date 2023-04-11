package main

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/miekg/dns"
	"go.bbkane.com/warg/command"
)

// getSubnet, either from --subnet-map or directly from subnet.
func getSubnet(subnetMap map[string]netip.Addr, subnet string) (net.IP, error) {

	// check in map
	if subAddr, exists := subnetMap[subnet]; exists {
		return subAddr.AsSlice(), nil
	}

	// try to parse directly
	subIP := net.ParseIP(subnet)
	if subIP == nil {
		return nil, fmt.Errorf("Could not parse IP: %s", subnet)
	}
	return subIP, nil
}

func cmdCtxToDigRepeatParams(cmdCtx command.Context) (*digRepeatParams, error) {

	// simple params
	count := cmdCtx.Flags["--count"].(int)
	fqdn := cmdCtx.Flags["--fqdn"].([]string)
	timeout := cmdCtx.Flags["--timeout"].(time.Duration)

	// rtypes
	rtypeStrs := cmdCtx.Flags["--rtype"].([]string)
	var rtypes []uint16
	for _, rtypeStr := range rtypeStrs {
		rtype, ok := dns.StringToType[rtypeStr]
		if !ok {
			return nil, fmt.Errorf("Couldn't parse rtype: %v", rtype)
		}
		rtypes = append(rtypes, rtype)
	}

	// subnets
	var subnets []net.IP

	var subnetMap map[string]netip.Addr = nil
	if sm, exists := cmdCtx.Flags["--subnet-map"].(map[string]netip.Addr); exists {
		subnetMap = sm
	}

	subnetStrs, _ := cmdCtx.Flags["--subnet"].([]string)

	for _, subnetStr := range subnetStrs {
		subnetIP, err := getSubnet(subnetMap, subnetStr)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnetIP)
	}

	// If we don't have any subnets, just use a list of one nil subnet :)
	if len(subnets) == 0 {
		subnets = append(subnets, nil)
	}

	// nameservers
	var nameservers []string

	var nameserverMap map[string]netip.AddrPort = nil
	if nsm, exists := cmdCtx.Flags["--ns-map"].(map[string]netip.AddrPort); exists {
		nameserverMap = nsm
	}

	// These might be names Or IP:Port, so let's not commit to this slice
	nameserverStrs := cmdCtx.Flags["--ns"].([]string)

	for _, nameserverStr := range nameserverStrs {
		// check in map
		if nsAddrPort, exists := nameserverMap[nameserverStr]; exists {
			nameservers = append(nameservers, nsAddrPort.String())
		} else {
			// try to parse directly
			_, err := netip.ParseAddrPort(nameserverStr)
			if err != nil {
				return nil, fmt.Errorf("could not parse --ns: %s : %w", nameserverStr, err)
			}
			nameservers = append(nameservers, nameserverStr)
		}
	}

	fmt.Printf("ns: %#v\n", nameservers)

	return &digRepeatParams{
		DigOneParams: digOneParams{
			FQDN:             fqdn[0],
			Rtype:            rtypes[0],
			NameserverIPPort: nameservers[0],
			SubnetIP:         subnets[0],
			Timeout:          timeout,
		},
		Count: count,
	}, nil
}

func printDigRepeat(p digRepeatParams, r digRepeatResult) {
	// fmt.Printf("answers: %#v\n", r)
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"FQDN", "Rtype", "Nameserver", "Subnet", "Ans/Err", "Count"})
	// answers
	for _, ans := range r.Answers {
		t.AppendRow(table.Row{
			p.DigOneParams.FQDN,
			dns.TypeToString[p.DigOneParams.Rtype],
			p.DigOneParams.NameserverIPPort,
			p.DigOneParams.SubnetIP.String(),
			strings.Join(ans.StringSlice, "\n"),
			ans.Count,
		})
	}
	// errors
	for _, err := range r.Errors {
		t.AppendRow(table.Row{
			p.DigOneParams.FQDN,
			dns.TypeToString[p.DigOneParams.Rtype],
			p.DigOneParams.NameserverIPPort,
			p.DigOneParams.SubnetIP.String(),
			err.String,
			err.Count,
		})
	}

	t.AppendSeparator()

	t.Render()
}

func runDig(cmdCtx command.Context) error {

	p, err := cmdCtxToDigRepeatParams(cmdCtx)
	if err != nil {
		return err
	}

	result := digRepeat(
		*p,
	)
	printDigRepeat(*p, result)
	return nil
}
