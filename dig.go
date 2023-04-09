package main

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"sort"
	"time"

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

func cmdCtxToDigOneparams(cmdCtx command.Context) (*digOneParams, int, error) {
	// TODO:
	// - make another function to turn cmdCtx into digOneParams, count, so I can easily test it with app.Parse
	// - ns is now a string, so get that...
	// - get some tests; not sure I trust my subnet-getting code :)

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
		return nil, 0, err
	}

	rtypeStr := cmdCtx.Flags["--rtype"].(string)
	rtype, ok := dns.StringToType[rtypeStr]
	if !ok {
		return nil, 0, fmt.Errorf("Couldn't parse rtype: %v", rtype)
	}

	// nameserverIP := cmdCtx.Flags["--ns"].(netip.AddrPort)
	// nameserverIPPort := nameserverIP.String()

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
			return nil, 0, fmt.Errorf("could not parse --ns: %s : %w", nameserverIPPort, err)
		}
	}

	p := digOneParams{
		FQDN:             fqdn,
		Rtype:            rtype,
		NameserverIPPort: nameserverIPPort,
		SubnetIP:         subnetIP,
		Timeout:          timeout,
	}
	return &p, count, nil
}

func runDig(cmdCtx command.Context) error {

	p, count, err := cmdCtxToDigOneparams(cmdCtx)
	if err != nil {
		return err
	}

	answers, errors := digRepeat(
		*p,
		count,
	)
	printDigRepeat(*p, count, answers, errors)
	return nil
}

func printDigRepeat(p digOneParams, count int, answers []stringSliceCount, errors []stringCount) {
	fmt.Printf("answers: %#v\n", answers)
	fmt.Printf("errors: %#v\n", errors)
}

type digOneParams struct {
	FQDN             string
	Rtype            uint16
	NameserverIPPort string
	SubnetIP         net.IP
	Timeout          time.Duration
}

// digOne an fqdn! Returns an error for rcode != NOERROR or an empty list of answers.
// Returns answers sorted
func digOne(p digOneParams) ([]string, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(p.FQDN), p.Rtype)

	// Add subnet!
	// https://github.com/miekg/exdns/blob/d851fa434ad51cb84500b3e18b8aa7d3bead2c51/q/q.go#L209
	if p.SubnetIP != nil {
		o := &dns.OPT{
			Hdr: dns.RR_Header{
				Name:     ".",
				Rrtype:   dns.TypeOPT,
				Class:    0,
				Ttl:      0,
				Rdlength: 0,
			},
			Option: nil,
		}

		e := &dns.EDNS0_SUBNET{
			Code:          dns.EDNS0SUBNET,
			Address:       p.SubnetIP,
			Family:        1, // IPv4
			SourceNetmask: net.IPv4len * 8,
			SourceScope:   0,
		}
		if e.Address.To4() == nil {
			e.Family = 2 // IP6
			e.SourceNetmask = net.IPv6len * 8
		}
		o.Option = append(o.Option, e)
		m.Extra = append(m.Extra, o)
	}

	clientCtx, cancel := context.WithTimeout(context.Background(), p.Timeout)
	defer cancel()

	in, err := dns.ExchangeContext(clientCtx, m, p.NameserverIPPort)
	if err != nil {
		return nil, fmt.Errorf("exchange err: %w", err)
	}
	if in.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("non-success rcode: %s", dns.RcodeToString[in.Rcode])
	}

	if len(in.Answer) < 1 {
		// This can happen if we query for CNAME for example
		return nil, fmt.Errorf("no answers returned")
	}

	answers := []string{}
	for _, e := range in.Answer {

		switch t := e.(type) {
		case *dns.A:
			answers = append(answers, t.A.String())
		case *dns.AAAA:
			answers = append(answers, t.AAAA.String())
		case *dns.CNAME:
			answers = append(answers, t.Target)
		default:
			return nil, fmt.Errorf("unknown record type: %T", e)
		}

	}
	sort.Strings(answers)
	return answers, nil
}

func digRepeat(p digOneParams, count int) ([]stringSliceCount, []stringCount) {
	answerCounter := newStringSliceCounter()
	errorCounter := newStringCounter()

	for i := 0; i < count; i++ {
		answer, err := digOne(p)
		if err != nil {
			errorCounter.Add(err.Error())
		} else {
			answerCounter.Add(answer)
		}
	}
	return answerCounter.AsSortedSlice(), errorCounter.AsSortedSlice()
}
