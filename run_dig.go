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
func getSubnet(subnetMap map[string]netip.Addr, subnet *string) (net.IP, error) {
	if subnet == nil {
		return nil, nil
	}

	// check in map
	if subAddr, exists := subnetMap[*subnet]; exists {
		return subAddr.AsSlice(), nil
	}

	// try to parse directly
	subIP := net.ParseIP(*subnet)
	if subIP == nil {
		return nil, fmt.Errorf("Could not parse IP: %s", *subnet)
	}
	return subIP, nil
}

func cmdCtxToDigRepeatParams(cmdCtx command.Context) (*digRepeatParams, error) {

	count := cmdCtx.Flags["--count"].(int)
	fqdn := cmdCtx.Flags["--fqdn"].(string)
	timeout := cmdCtx.Flags["--timeout"].(time.Duration)

	// get ns IP:Port

	// Get a subnet
	var subnetMap map[string]netip.Addr = nil
	if sm, exists := cmdCtx.Flags["--subnet-map"].(map[string]netip.Addr); exists {
		subnetMap = sm
	}
	var subnetStr *string = nil
	if sub, exists := cmdCtx.Flags["--subnet"].(string); exists {
		subnetStr = &sub
	}
	subnetIP, err := getSubnet(subnetMap, subnetStr)
	if err != nil {
		return nil, err
	}

	rtypeStr := cmdCtx.Flags["--rtype"].(string)
	rtype, ok := dns.StringToType[rtypeStr]
	if !ok {
		return nil, fmt.Errorf("Couldn't parse rtype: %v", rtype)
	}

	nameserverIPPort := cmdCtx.Flags["--ns"].(string)
	var nameserverMap map[string]netip.AddrPort = nil
	if nsm, exists := cmdCtx.Flags["--ns-map"].(map[string]netip.AddrPort); exists {
		nameserverMap = nsm
	}
	// check in map
	if nsAddrPort, exists := nameserverMap[nameserverIPPort]; exists {
		nameserverIPPort = nsAddrPort.String()
	} else {
		_, err := netip.ParseAddrPort(nameserverIPPort)
		if err != nil {
			return nil, fmt.Errorf("could not parse --ns: %s : %w", nameserverIPPort, err)
		}
	}

	return &digRepeatParams{
		DigOneParams: digOneParams{
			FQDN:             fqdn,
			Rtype:            rtype,
			NameserverIPPort: nameserverIPPort,
			SubnetIP:         subnetIP,
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
