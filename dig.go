package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"go.bbkane.com/warg/command"
)

func runDig(cmdCtx command.Context) error {

	fqdn := cmdCtx.Flags["--fqdn"].(string)
	rtypeStr := cmdCtx.Flags["--rtype"].(string)
	nameserverIP := cmdCtx.Flags["--nameserver-ip"].(string)
	nameserverPort := cmdCtx.Flags["--nameserver-port"].(int)

	// TODO: instead of parsing froma string, add IP type to warg
	var subnetIP net.IP = nil
	if sub, exists := cmdCtx.Flags["--subnet-ip"].(string); exists {
		subnetIP = net.ParseIP(sub)
		if subnetIP == nil {
			return fmt.Errorf("failure to parse IP: %s", subnetIP)

		}
	}

	timeout := cmdCtx.Flags["--timeout"].(time.Duration)

	rtype, ok := dns.StringToType[rtypeStr]
	if !ok {
		return fmt.Errorf("Couldn't parse rtype: %v", rtype)
	}

	nameserverIPPort := nameserverIP + ":" + strconv.Itoa(nameserverPort)

	rcode, answers, err := dig(
		fqdn,
		rtype,
		nameserverIPPort,
		subnetIP,
		timeout,
	)
	fmt.Printf("rcode: %s\n", dns.RcodeToString[rcode])
	fmt.Printf("answers: %s\n", answers)
	return err

}

func dig(fqdn string, rtype uint16, nameserverIPPort string, subnetIP net.IP, timeout time.Duration) (int, []string, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(fqdn), rtype)

	// Add subnet!
	// https://github.com/miekg/exdns/blob/d851fa434ad51cb84500b3e18b8aa7d3bead2c51/q/q.go#L209
	if subnetIP != nil {
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
			Address:       subnetIP,
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

	clientCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	in, err := dns.ExchangeContext(clientCtx, m, nameserverIPPort)
	if err != nil {
		return 0, nil, fmt.Errorf("exchange err: %w", err)
	}
	if len(in.Answer) < 1 {
		return 0, nil, fmt.Errorf("no answers returned")
	}

	answers := []string{}
	for _, e := range in.Answer {
		// TODO: don't rely on this being an A record! try to convert to proper type
		if t, ok := e.(*dns.A); ok {
			answers = append(answers, t.A.String())
		} else {
			return 0, nil, fmt.Errorf("not an a record: %s", e)
		}
	}

	return in.Rcode, answers, nil
}
