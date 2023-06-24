package diglist

import (
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg/command"
)

func Run(cmdCtx command.Context) error {

	counts := cmdCtx.Flags["--count"].([]int)
	fqdns := cmdCtx.Flags["--fqdn"].([]string)
	nameservers := cmdCtx.Flags["--nameserver"].([]string)
	protocols := cmdCtx.Flags["--protocol"].([]string)
	rtypes := cmdCtx.Flags["--rtype"].([]string)
	subnets := cmdCtx.Flags["--subnet"].([]string)
	timeouts := cmdCtx.Flags["--timeout"].([]time.Duration)

	lens := map[string]int{
		"--count":      len(counts),
		"--fqdn":       len(fqdns),
		"--nameserver": len(nameservers),
		"--protocol":   len(protocols),
		"--rtype":      len(rtypes),
		"--subnet":     len(subnets),
		"--timeout":    len(timeouts),
	}
	target_len := lens["--fqdn"]
	for name, l := range lens {
		if l != target_len {
			return fmt.Errorf("unexpected count of flag %s: %d. Expected %d to match count of --fqdn", name, l, target_len)
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

	digRepeatParamsSlice := []dig.DigRepeatParams{}

	for _, count := range counts {
		for _, fqdn := range fqdns {
			for _, nameserver := range nameservers {
				for _, protocol := range protocols {
					for _, rtype := range rtypes {
						for _, subnetIP := range subnetIPs {
							for _, timeout := range timeouts {
								digRepeatParamsSlice = append(digRepeatParamsSlice, dig.DigRepeatParams{
									DigOneParams: dig.DigOneParams{
										FQDN:             fqdn,
										NameserverIPPort: nameserver, // TODO: ensure this ends in port...
										Proto:            protocol,
										Rtype:            dns.StringToType[rtype],
										SubnetIP:         subnetIP,
										Timeout:          timeout,
									},
									Count: count,
								},
								)
							}
						}
					}
				}
			}
		}
	}

	results := dig.DigList(digRepeatParamsSlice, dig.DigOne)
	fmt.Printf("%v\n", results)

	return nil
}
