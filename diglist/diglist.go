package diglist

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg/command"
	"gopkg.in/yaml.v3"
)

type Rdata struct {
	Content []string `yaml:"content"`
	Count   int      `yaml:"count"`
}

type Error struct {
	Count int    `yaml:"count"`
	Msg   string `yaml:"msg"`
}

type Result struct {
	Rdata  []Rdata `yaml:"rdata"`
	Errors []Error `yaml:"errors"`
}

type Return struct {
	Results []Result `yaml:"results"`
}

func Run(cmdCtx command.Context) error {

	counts := cmdCtx.Flags["--count"].([]int)
	qnames := cmdCtx.Flags["--qname"].([]string)
	nameservers := cmdCtx.Flags["--nameserver"].([]string)
	protocols := cmdCtx.Flags["--protocol"].([]string)
	rtypes := cmdCtx.Flags["--rtype"].([]string)
	subnets := cmdCtx.Flags["--subnet"].([]string)
	timeouts := cmdCtx.Flags["--timeout"].([]time.Duration)

	lens := map[string]int{
		"--count":      len(counts),
		"--qname":      len(qnames),
		"--nameserver": len(nameservers),
		"--protocol":   len(protocols),
		"--rtype":      len(rtypes),
		"--subnet":     len(subnets),
		"--timeout":    len(timeouts),
	}
	target_len := lens["--qname"]
	for name, l := range lens {
		if l != target_len {
			return fmt.Errorf("unexpected count of flag %s: %d. Expected %d to match count of --qname", name, l, target_len)
		}
	}

	// convert subnet to net.IP:
	var subnetIPs []net.IP
	for _, subnet := range subnets {
		if subnet == "none" {
			subnetIPs = append(subnetIPs, nil)
			continue
		}
		subnetIP := net.ParseIP(subnet)
		if subnetIP == nil {
			return fmt.Errorf("Could not parse subnet IP: %s", subnet)
		}
		subnetIPs = append(subnetIPs, subnetIP)

	}

	// convert input params to API params
	digRepeatParamsSlice := []dig.DigRepeatParams{}

	for i := range qnames {
		digRepeatParamsSlice = append(digRepeatParamsSlice, dig.DigRepeatParams{
			DigOneParams: dig.DigOneParams{
				Qname:            qnames[i],
				NameserverIPPort: nameservers[i],
				Proto:            protocols[i],
				Rtype:            dns.StringToType[rtypes[i]],
				SubnetIP:         subnetIPs[i],
				Timeout:          timeouts[i],
			},
			Count: counts[i],
		},
		)
	}

	ctx := context.Background()
	dRes := dig.DigRepeatParallel(ctx, digRepeatParamsSlice, dig.DigOne)

	// convert API result to printable result

	ret := Return{
		Results: make([]Result, len(dRes)),
	}
	for i := range dRes {
		ret.Results[i].Rdata = make([]Rdata, len(dRes[i].Answers))
		for r := range dRes[i].Answers {
			ret.Results[i].Rdata[r].Content = dRes[i].Answers[r].StringSlice
			ret.Results[i].Rdata[r].Count = dRes[i].Answers[r].Count
		}

		ret.Results[i].Errors = make([]Error, len(dRes[i].Errors))
		for e := range dRes[i].Errors {
			ret.Results[i].Errors[e].Msg = dRes[i].Errors[e].String
			ret.Results[i].Errors[e].Count = dRes[i].Errors[e].Count

		}
	}

	encoder := yaml.NewEncoder(os.Stdout)
	encoder.SetIndent(2)
	err := encoder.Encode(&ret)
	if err != nil {
		return fmt.Errorf("could not serialize to yaml: %w", err)
	}

	return nil
}
