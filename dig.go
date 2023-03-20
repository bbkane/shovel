package main

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
	"go.bbkane.com/warg/command"
)

func dig(ctx command.Context) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("linkedin.com"), dns.TypeA)

	// Add subnet!
	// https://github.com/miekg/exdns/blob/d851fa434ad51cb84500b3e18b8aa7d3bead2c51/q/q.go#L209
	{
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
		ip := "101.251.8.0" // China
		e := &dns.EDNS0_SUBNET{
			Code:          dns.EDNS0SUBNET,
			Address:       net.ParseIP(ip),
			Family:        1, // IPv4
			SourceNetmask: net.IPv4len * 8,
			SourceScope:   0,
		}
		if e.Address == nil {
			return fmt.Errorf("failure to parse IP: %s", ip)
		}
		if e.Address.To4() == nil {
			e.Family = 2 // IP6
			e.SourceNetmask = net.IPv6len * 8
		}
		o.Option = append(o.Option, e)
		m.Extra = append(m.Extra, o)
	}

	in, err := dns.Exchange(m, "198.51.45.9:53") // dns2.p09.nsone.net.
	if err != nil {
		return fmt.Errorf("exchange err: %w", err)
	}
	fmt.Printf("rcode: %s\n", dns.RcodeToString[in.Rcode])
	if len(in.Answer) < 1 {
		return fmt.Errorf("no answers returned")
	}
	if t, ok := in.Answer[0].(*dns.A); ok {
		fmt.Printf("%s\n", t.A)
	} else {
		return fmt.Errorf("not an a record: %s", in.Answer[0])
	}

	return nil
}
