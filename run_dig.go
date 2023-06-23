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
	"go.bbkane.com/shovel/dig"
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

func getDigFunc(name string) dig.DigOneFunc {
	digs := map[string]dig.DigOneFunc{
		"none": dig.DigOne,
		"simple": dig.DigOneFuncMock(
			[]dig.DigOneResult{
				{Answers: []string{"1.2.3.4"}, Err: nil},
			},
		),
		"twocount": dig.DigOneFuncMock(
			[]dig.DigOneResult{
				{Answers: []string{"1.2.3.4"}, Err: nil},
				{Answers: []string{"1.2.3.4"}, Err: nil},
			},
		),
	}
	dig, exists := digs[name]
	if !exists {
		panic("could not find dig func: " + name)
	}
	return dig
}

type parsedCmdCtx struct {
	DigRepeatParams []dig.DigRepeatParams
	NameserverNames map[string]string
	SubnetNames     map[string]string
	Dig             dig.DigOneFunc
	Stdout          *os.File
}

func parseCmdCtx(cmdCtx command.Context) (*parsedCmdCtx, error) {

	// simple params
	count := cmdCtx.Flags["--count"].(int)
	fqdns := cmdCtx.Flags["--fqdn"].([]string)
	timeout := cmdCtx.Flags["--timeout"].(time.Duration)
	proto := cmdCtx.Flags["--protocol"].(string)

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

	mockDigFuncStr := cmdCtx.Flags["--mock-dig-func"].(string)
	digFunc := getDigFunc(mockDigFuncStr)

	// subnets
	var subnets []net.IP
	subnetNames := make(map[string]string)

	subnetMap, _ := cmdCtx.Flags["--subnet-map"].(map[string]netip.Addr)
	subnetStrs, _ := cmdCtx.Flags["--subnet"].([]string)

	// if '--subnet all' is the only thing passed, add all subnets from the map
	if len(subnetStrs) == 1 && subnetStrs[0] == "all" && len(subnetMap) > 0 {
		subnetStrs = []string{}
		for key := range subnetMap {
			subnetStrs = append(subnetStrs, key)
		}
	}

	for _, subnetStr := range subnetStrs {
		subnetIP, name, err := getSubnet(subnetMap, subnetStr)
		subnetNames[subnetIP.String()] = name
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
	nameserverNames := make(map[string]string)

	// NOTE: if the wrong types are asserted, the resulting map is nil...
	// It would be nice if Go was kind enough to panic...
	nameserverMap, _ := cmdCtx.Flags["--ns-map"].(map[string]string)

	// These might be names Or IP:Port, so let's not use this slice directly
	nameserverStrs := cmdCtx.Flags["--ns"].([]string)

	// if --ns all is the only thing passed, add all nameservers from the map
	if len(nameserverStrs) == 1 && nameserverStrs[0] == "all" && len(nameserverMap) > 0 {
		nameserverStrs = []string{}
		for key := range nameserverMap {
			nameserverStrs = append(nameserverStrs, key)
		}
	}

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
			return nil, fmt.Errorf("error in nameserver: %s: %w", nameserver, err)
		}
	}

	if len(nameservers) == 0 {
		return nil, errors.New("no nameservers passed")
	}

	digRepeatParamsSlice := []dig.DigRepeatParams{}

	for _, fqdn := range fqdns {
		for _, rtype := range rtypes {
			for _, subnet := range subnets {
				for _, nameserver := range nameservers {
					digRepeatParamsSlice = append(digRepeatParamsSlice, dig.DigRepeatParams{
						DigOneParams: dig.DigOneParams{
							FQDN:             fqdn,
							Rtype:            rtype,
							NameserverIPPort: nameserver,
							SubnetIP:         subnet,
							Timeout:          timeout,
							Proto:            proto,
						},
						Count: count,
					})
				}
			}
		}
	}

	return &parsedCmdCtx{
		DigRepeatParams: digRepeatParamsSlice,
		NameserverNames: nameserverNames,
		SubnetNames:     subnetNames,
		Dig:             digFunc,
		Stdout:          cmdCtx.Stdout,
	}, nil
}

func printDigRepeat(t table.Writer, parsed parsedCmdCtx, p dig.DigRepeatParams, r dig.DigRepeatResult) {

	fmtSubnet := func(subnet net.IP) string {
		if subnet == nil {
			return ""
		}
		subnetStr := subnet.String()
		name := parsed.SubnetNames[subnetStr]
		return "# " + name + "\n" + subnet.String()
	}

	fmtNS := func(ns string) string {
		name := parsed.NameserverNames[ns]
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

func runDigCombine(cmdCtx command.Context) error {

	parsed, err := parseCmdCtx(cmdCtx)
	if err != nil {
		return err
	}

	results := dig.DigList(parsed.DigRepeatParams, parsed.Dig)

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(parsed.Stdout)

	// due to the way parsing works, if the first subnet is nil,
	// we can assume the rest are too. If so, hide the subnet column
	hideSubnets := false
	if len(parsed.DigRepeatParams) > 0 && parsed.DigRepeatParams[0].DigOneParams.SubnetIP == nil {
		// columnConfigs[2].Hidden = true
		hideSubnets = true
	}

	hideCount := false
	// due to the way parsing works, if the first count is none,
	// we can assume the rest are too. If so, hide the count column
	if len(parsed.DigRepeatParams) > 0 && parsed.DigRepeatParams[0].Count == 1 {
		hideCount = true
	}

	columnConfigs := []table.ColumnConfig{
		{Name: "FQDN", AutoMerge: true},
		{Name: "Rtype", AutoMerge: true},
		{Name: "Subnet", AutoMerge: true, Hidden: hideSubnets},
		{Name: "Nameserver", AutoMerge: true},
		{Name: "Ans/Err"},
		{Name: "Count", Hidden: hideCount},
	}

	t.SetColumnConfigs(columnConfigs)

	t.AppendHeader(table.Row{"FQDN", "Rtype", "Subnet", "Nameserver", "Ans/Err", "Count"})

	for i := 0; i < len(parsed.DigRepeatParams); i++ {
		printDigRepeat(t, *parsed, parsed.DigRepeatParams[i], results[i])
	}

	t.Render()

	return nil
}
