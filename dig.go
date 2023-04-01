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

func runDigOne(cmdCtx command.Context) error {

	fqdn := cmdCtx.Flags["--fqdn"].(string)
	rtypeStr := cmdCtx.Flags["--rtype"].(string)
	nameserverIP := cmdCtx.Flags["--nameserver-addr-port"].(netip.AddrPort)

	// warg needs a netip.Addr because it's comparable, but dig wants a net.IP
	var subnetIP net.IP = nil
	if sub, exists := cmdCtx.Flags["--subnet-addr"].(netip.Addr); exists {
		subnetIP = sub.AsSlice()
	}

	timeout := cmdCtx.Flags["--timeout"].(time.Duration)

	rtype, ok := dns.StringToType[rtypeStr]
	if !ok {
		return fmt.Errorf("Couldn't parse rtype: %v", rtype)
	}

	nameserverIPPort := nameserverIP.String()

	answers, err := digOne(
		digOneParams{
			FQDN:             fqdn,
			Rtype:            rtype,
			NameserverIPPort: nameserverIPPort,
			SubnetIP:         subnetIP,
			Timeout:          timeout,
		},
	)
	fmt.Printf("answers: %s\n", answers)
	return err

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
