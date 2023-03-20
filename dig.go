package main

import (
	"fmt"

	"github.com/miekg/dns"
	"go.bbkane.com/warg/command"
)

func dig(ctx command.Context) error {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("linkedin.com"), dns.TypeA)

	in, err := dns.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return fmt.Errorf("exchange err: %w", err)
	}
	if t, ok := in.Answer[0].(*dns.A); ok {
		fmt.Printf("%s\n", t)
	}

	return nil
}
