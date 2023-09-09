package digcombine

import (
	"context"
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
			context.Background(),
			[]dig.DigOneResult{
				{Answers: []string{"1.2.3.4"}, Err: nil},
			},
		),
		"twocount": dig.DigOneFuncMock(
			context.Background(),
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
	Dig             dig.DigOneFunc
	DigRepeatParams []dig.DigRepeatParams
	GlobalTimeout   time.Duration
	NameserverNames map[string]string
	Stdout          *os.File
	SubnetToName    map[string]string
}

func ConvertRTypes(rtypeStrs []string) ([]uint16, error) {
	var rtypes []uint16
	for _, rtypeStr := range rtypeStrs {
		rtype, ok := dns.StringToType[rtypeStr]
		if !ok {
			return nil, fmt.Errorf("Couldn't parse rtype: %v", rtype)
		}
		rtypes = append(rtypes, rtype)
	}
	return rtypes, nil
}

// ParseSubnets turns a list of passed subnets into a list of net.IP for digging,
// a map of stringified subnet to name, and an error.
// It uses the following rules:
//
//   - If passedSubnets is empty, returns []net.IP{nil}.
//
//   - If passedSubnets == {"all"} and we have a non-empty subnetMap, return everything in subnetMap.
//
// Loop through passedSubnets
//
//   - first check if subnet == "none"
//
//   - then try to lookup up the passed subnet in subnetMap,
//
//   - then try to parse as an IP.
//
// Fail if we can't find it in the map or parse it as an IP.
func ParseSubnets(passedSubnets []string, subnetMap map[string]net.IP) ([]net.IP, map[string]string, error) {

	// no subnets -> {nil}
	if len(passedSubnets) == 0 {
		return []net.IP{nil}, nil, nil
	}
	// if "all" is the only thing passed, add everything from subnetMap
	if len(passedSubnets) == 1 && passedSubnets[0] == "all" && len(subnetMap) > 0 {
		parsed := []net.IP{}
		subnetToName := make(map[string]string)
		for name, ip := range subnetMap {
			parsed = append(parsed, ip)
			subnetToName[ip.String()] = name
		}
		return parsed, subnetToName, nil
	}

	// Loop through passed and try to parse
	parsed := []net.IP{}
	subnetToName := make(map[string]string)
	for _, passed := range passedSubnets {

		// check for "none"
		if passed == "none" {
			parsed = append(parsed, nil)
			// net.IP(nil).String() == "<nil>"
			subnetToName["<nil>"] = "none"
			continue
		}

		// try to retrieve from map
		if subIP, exists := subnetMap[passed]; exists {
			parsed = append(parsed, subIP)
			subnetToName[subIP.String()] = passed
			continue
		}

		// try to parse as IP
		subIP := net.ParseIP(passed)
		if subIP == nil {
			return nil, nil, fmt.Errorf("Could not parse IP: %s", passed)
		}
		parsed = append(parsed, subIP)
		subnetToName[passed] = "passed ip"
	}
	return parsed, subnetToName, nil
}

func parseCmdCtx(cmdCtx command.Context) (*parsedCmdCtx, error) {

	// simple params
	count := cmdCtx.Flags["--count"].(int)
	qnames := cmdCtx.Flags["--qname"].([]string)
	globalTimeout := cmdCtx.Flags["--global-timeout"].(time.Duration)
	proto := cmdCtx.Flags["--protocol"].(string)

	// rtypes
	rtypeStrs := cmdCtx.Flags["--rtype"].([]string)
	rtypes, err := ConvertRTypes(rtypeStrs)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse cmdCtx: %w", err)
	}

	mockDigFuncStr := cmdCtx.Flags["--mock-dig-func"].(string)
	digFunc := getDigFunc(mockDigFuncStr)

	// subnet
	nameToSubnetNetIPAddr, _ := cmdCtx.Flags["--subnet-map"].(map[string]netip.Addr)
	// convert from netip.Addr to net.IP
	nameToSubnet := make(map[string]net.IP)
	for name, addr := range nameToSubnetNetIPAddr {
		nameToSubnet[name] = net.IP(addr.AsSlice())
	}
	passedSubnetStrs, _ := cmdCtx.Flags["--subnet"].([]string)

	parsedSubnets, subnetToName, err := ParseSubnets(passedSubnetStrs, nameToSubnet)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse subnets: %w", err)
	}

	// nameservers
	var nameservers []string
	nameserverNames := make(map[string]string)

	// NOTE: if the wrong types are asserted, the resulting map is nil...
	// It would be nice if Go was kind enough to panic...
	nameserverMap, _ := cmdCtx.Flags["--nameserver-map"].(map[string]string)

	// These might be names Or IP:Port, so let's not use this slice directly
	nameserverStrs := cmdCtx.Flags["--nameserver"].([]string)

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

	digRepeatParamsSlice := dig.CombineDigRepeatParams(
		nameservers,
		proto,
		qnames,
		rtypes,
		parsedSubnets,
		count,
	)

	return &parsedCmdCtx{
		Dig:             digFunc,
		DigRepeatParams: digRepeatParamsSlice,
		GlobalTimeout:   globalTimeout,
		NameserverNames: nameserverNames,
		Stdout:          cmdCtx.Stdout,
		SubnetToName:    subnetToName,
	}, nil
}

func printDigRepeat(t table.Writer, parsed parsedCmdCtx, p dig.DigRepeatParams, r dig.DigRepeatResult) {

	fmtSubnet := func(subnet net.IP) string {
		if subnet == nil {
			return ""
		}
		subnetStr := subnet.String()
		name := parsed.SubnetToName[subnetStr]
		return "# " + name + "\n" + subnet.String()
	}

	fmtNS := func(ns string) string {
		name := parsed.NameserverNames[ns]
		return "# " + name + "\n" + ns
	}

	// answers
	for _, ans := range r.Answers {
		t.AppendRow(table.Row{
			p.DigOneParams.Qname,
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
			p.DigOneParams.Qname,
			dns.TypeToString[p.DigOneParams.Rtype],
			fmtSubnet(p.DigOneParams.SubnetIP),
			fmtNS(p.DigOneParams.NameserverIPPort),
			err.String,
			err.Count,
		})
	}

	t.AppendSeparator()

}

func Run(cmdCtx command.Context) error {

	parsed, err := parseCmdCtx(cmdCtx)
	if err != nil {
		return err
	}

	if len(parsed.DigRepeatParams) < 1 {
		return errors.New("no dig parameters passed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), parsed.GlobalTimeout)
	defer cancel()

	results := dig.DigRepeatParallel(ctx, parsed.DigRepeatParams, parsed.Dig)

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(parsed.Stdout)

	// due to the way parsing works, if the first subnet is nil,
	// we can assume the rest are too. If so, hide the subnet column
	hideSubnets := parsed.DigRepeatParams[0].DigOneParams.SubnetIP == nil

	// due to the way parsing works, if the first count is none,
	// we can assume the rest are too. If so, hide the count column
	hideCount := parsed.DigRepeatParams[0].Count == 1

	//nolint:exhaustruct
	columnConfigs := []table.ColumnConfig{
		{Name: "Qname", AutoMerge: true},
		{Name: "Rtype", AutoMerge: true},
		{Name: "Subnet", AutoMerge: true, Hidden: hideSubnets},
		{Name: "Nameserver", AutoMerge: true},
		{Name: "Ans/Err"},
		{Name: "Count", Hidden: hideCount},
	}

	t.SetColumnConfigs(columnConfigs)

	t.AppendHeader(table.Row{"Qname", "Rtype", "Subnet", "Nameserver", "Ans/Err", "Count"})

	for i := 0; i < len(parsed.DigRepeatParams); i++ {
		printDigRepeat(t, *parsed, parsed.DigRepeatParams[i], results[i])
	}

	t.Render()

	return nil
}
