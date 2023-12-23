package dig

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sourcegraph/conc/iter"
	"go.bbkane.com/shovel/counter"
)

type DigOneParams struct {
	NameserverIPPort string
	Proto            string
	Qname            string
	Rtype            uint16
	SubnetIP         net.IP
	Timeout          time.Duration
}

func EmptyDigOneparams() DigOneParams {
	return DigOneParams{
		NameserverIPPort: "",
		Proto:            "",
		Qname:            "",
		Rtype:            0,
		SubnetIP:         nil,
		Timeout:          0,
	}
}

type DigOneFuncCtxKey struct{}

type DigOneFunc func(ctx context.Context, p DigOneParams) ([]string, error)

type DigOneResult struct {
	Answers []string
	Err     error
}

func DigOneFuncMock(_ context.Context, rets []DigOneResult) DigOneFunc {
	var i int
	return func(_ context.Context, p DigOneParams) ([]string, error) {
		if i >= len(rets) {
			panic("Ran out of returns!")
		}
		ret := rets[i]
		i++
		return ret.Answers, ret.Err
	}
}

// DigOne a qname! Returns an error for rcode != NOERROR or an empty list of answers.
// Returns answers sorted alphabetically.
// If there is both a context deadline and a configured timeout on `DigOneParams`, the earliest of the two takes effect.
func DigOne(ctx context.Context, p DigOneParams) ([]string, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(p.Qname), p.Rtype)

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

	client := dns.Client{
		Net:            p.Proto,
		UDPSize:        0,
		TLSConfig:      nil,
		Dialer:         nil,
		Timeout:        p.Timeout,
		DialTimeout:    0,
		ReadTimeout:    0,
		WriteTimeout:   0,
		TsigSecret:     nil,
		TsigProvider:   nil,
		SingleInflight: false,
	}
	in, _, err := client.ExchangeContext(ctx, m, p.NameserverIPPort)

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
		case *dns.MX:
			answers = append(answers, t.Mx)
		case *dns.NS:
			answers = append(answers, t.Ns)
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

type DigRepeatParams struct {
	DigOneParams DigOneParams
	Count        int
}

type DigRepeatResult struct {
	Answers []counter.StringSliceCount
	Errors  []counter.StringCount
}

// DigRepeat runs DigOne multiple times and sums the answers and errors
func DigRepeat(ctx context.Context, p DigRepeatParams, dig DigOneFunc) DigRepeatResult {
	answerCounter := counter.NewStringSliceCounter()
	errorCounter := counter.NewStringCounter()

	for i := 0; i < p.Count; i++ {
		answer, err := dig(ctx, p.DigOneParams)
		if err != nil {
			errorCounter.Add(err.Error())
		} else {
			answerCounter.Add(answer)
		}
	}
	return DigRepeatResult{
		Answers: answerCounter.AsSortedSlice(),
		Errors:  errorCounter.AsSortedSlice(),
	}
}

// CombineDigRepeatParams combines all the slcies passed. Ensure all of them have a length > 0
func CombineDigRepeatParams(nameservers []string, proto string, qnames []string, rtypes []uint16, subnets []net.IP, count int) []DigRepeatParams {
	// TODO: range over protos
	digRepeatParamsSlice := []DigRepeatParams{}

	for _, qname := range qnames {
		for _, rtype := range rtypes {
			for _, subnet := range subnets {
				for _, nameserver := range nameservers {
					digRepeatParamsSlice = append(digRepeatParamsSlice, DigRepeatParams{
						DigOneParams: DigOneParams{
							NameserverIPPort: nameserver,
							Proto:            proto,
							Qname:            qname,
							Rtype:            rtype,
							SubnetIP:         subnet,
							Timeout:          0, // TODO: implement per-dig timeouts
						},
						Count: count,
					})
				}
			}
		}
	}
	return digRepeatParamsSlice
}

// DigRepeatParallel runs DigRepeats in parallel and returns a slice of their results
func DigRepeatParallel(ctx context.Context, params []DigRepeatParams, dig DigOneFunc) []DigRepeatResult {
	return iter.Map(params, func(p *DigRepeatParams) DigRepeatResult {
		return DigRepeat(ctx, *p, dig)
	})
}
