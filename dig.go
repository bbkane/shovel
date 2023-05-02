package main

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sourcegraph/conc/iter"
)

type digOneParams struct {
	FQDN             string
	Rtype            uint16
	NameserverIPPort string
	SubnetIP         net.IP
	Timeout          time.Duration
	Proto            string
}

func emptyDigOneparams() digOneParams {
	return digOneParams{
		FQDN:             "",
		Rtype:            0,
		NameserverIPPort: "",
		SubnetIP:         nil,
		Timeout:          0,
		Proto:            "",
	}
}

type digOneFunc func(p digOneParams) ([]string, error)

type digOneReturns struct {
	Answers []string
	Err     error
}

func digOneFuncMock(rets []digOneReturns) digOneFunc {
	var i int
	return func(p digOneParams) ([]string, error) {
		if i >= len(rets) {
			panic("Ran out of returns!")
		}
		ret := rets[i]
		i++
		return ret.Answers, ret.Err
	}
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

	// in, err := dns.ExchangeContext(clientCtx, m, p.NameserverIPPort)

	client := dns.Client{
		Net:            p.Proto,
		UDPSize:        0,
		TLSConfig:      nil,
		Dialer:         nil,
		Timeout:        0,
		DialTimeout:    0,
		ReadTimeout:    0,
		WriteTimeout:   0,
		TsigSecret:     nil,
		TsigProvider:   nil,
		SingleInflight: false,
	}
	in, _, err := client.ExchangeContext(clientCtx, m, p.NameserverIPPort)

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
		case *dns.TXT:
			// NOTE: the dns lib has a MUCH fancier private way to do this
			// Maybe I should copy that :)
			answers = append(answers, strings.Join(t.Txt, " "))
		default:
			return nil, fmt.Errorf("unknown record type: %T", e)
		}

	}
	sort.Strings(answers)
	return answers, nil
}

type digRepeatParams struct {
	DigOneParams digOneParams
	Count        int
}

type digRepeatResult struct {
	Answers []stringSliceCount
	Errors  []stringCount
}

func digRepeat(p digRepeatParams, dig digOneFunc) digRepeatResult {
	answerCounter := newStringSliceCounter()
	errorCounter := newStringCounter()

	for i := 0; i < p.Count; i++ {
		answer, err := dig(p.DigOneParams)
		if err != nil {
			errorCounter.Add(err.Error())
		} else {
			answerCounter.Add(answer)
		}
	}
	return digRepeatResult{
		Answers: answerCounter.AsSortedSlice(),
		Errors:  errorCounter.AsSortedSlice(),
	}
}

func digVaried(params []digRepeatParams, dig digOneFunc) []digRepeatResult {
	return iter.Map(params, func(p *digRepeatParams) digRepeatResult {
		return digRepeat(*p, dig)
	})
}
