package main

import (
	"fmt"
	"net/netip"

	"go.bbkane.com/warg/value/contained"
)

func ipFromString(s string) (netip.Addr, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("could not parse ip from string: %s: %w", s, err)
	}
	return addr, nil
}

func Addr() contained.ContainedTypeInfo[netip.Addr] {
	return contained.ContainedTypeInfo[netip.Addr]{
		Description: "IP address",
		Empty: func() netip.Addr {
			return netip.Addr{}
		},
		FromIFace: func(iFace interface{}) (netip.Addr, error) {
			switch under := iFace.(type) {
			case netip.Addr:
				return under, nil
			case []byte:
				ip, ok := netip.AddrFromSlice(under)
				if !ok {
					return netip.Addr{}, fmt.Errorf("Could not convert %s to netip.Addr", string(under))
				}
				return ip, nil
			case string:
				return ipFromString(under)
			default:
				return netip.Addr{}, contained.ErrIncompatibleInterface
			}
		},
		FromInstance: func(a netip.Addr) (netip.Addr, error) {
			return a, nil
		},
		FromString: ipFromString,
	}
}

func AddrPort() contained.ContainedTypeInfo[netip.AddrPort] {
	return contained.ContainedTypeInfo[netip.AddrPort]{
		Description: "IP and Port number separated by a colon: ip:port ",
		Empty: func() netip.AddrPort {
			return netip.AddrPort{}
		},
		FromIFace: func(iFace interface{}) (netip.AddrPort, error) {
			switch under := iFace.(type) {
			case netip.AddrPort:
				return under, nil
			case string:
				return netip.ParseAddrPort(under)
			default:
				return netip.AddrPort{}, contained.ErrIncompatibleInterface
			}
		},
		FromString: netip.ParseAddrPort,
		FromInstance: func(ap netip.AddrPort) (netip.AddrPort, error) {
			return ap, nil
		},
	}
}
