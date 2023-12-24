package digcombine

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg/command"
)

func TestParseSubnet(t *testing.T) {
	tests := []struct {
		name                string
		passedSubnets       []string
		subnetMap           map[string]net.IP
		expectedSubnets     []net.IP
		expectedSubnetNames map[string]string
		expectedErr         bool
	}{
		{
			name:                "noSubnet",
			passedSubnets:       nil,
			subnetMap:           nil,
			expectedSubnets:     []net.IP{nil},
			expectedSubnetNames: nil,
			expectedErr:         false,
		},
		{
			name:                "subnetPassedAsArg",
			passedSubnets:       []string{"1.2.3.0"},
			subnetMap:           nil,
			expectedSubnets:     []net.IP{net.ParseIP("1.2.3.0")},
			expectedSubnetNames: map[string]string{"1.2.3.0": "passed ip"},
			expectedErr:         false,
		},
		{
			name:                "badSubnetPassedAsArg",
			passedSubnets:       []string{"badSubnet"},
			subnetMap:           nil,
			expectedSubnets:     nil,
			expectedSubnetNames: nil,
			expectedErr:         true,
		},
		{
			name:                "subnetFromMap",
			passedSubnets:       []string{"mysubnet"},
			subnetMap:           map[string]net.IP{"mysubnet": net.ParseIP("3.4.5.0")},
			expectedSubnets:     []net.IP{net.ParseIP("3.4.5.0")},
			expectedSubnetNames: map[string]string{"3.4.5.0": "mysubnet"},
			expectedErr:         false,
		},
		{
			name:                "subnetAll",
			passedSubnets:       []string{"all"},
			subnetMap:           map[string]net.IP{"subnetName": net.ParseIP("1.1.1.0")},
			expectedSubnets:     []net.IP{net.ParseIP("1.1.1.0")},
			expectedSubnetNames: map[string]string{"1.1.1.0": "subnetName"},
			expectedErr:         false,
		},
		{
			name:                "subnetNone",
			passedSubnets:       nil,
			subnetMap:           nil,
			expectedSubnets:     []net.IP{nil},
			expectedSubnetNames: nil,
			expectedErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualSubnets, actualSubnetNames, actualErr := ParseSubnets(tt.passedSubnets, tt.subnetMap)
			if tt.expectedErr {
				require.NotNil(t, actualErr)
			} else {
				require.Nil(t, actualErr)
			}
			require.Equal(t, tt.expectedSubnets, actualSubnets, "subnets")
			require.Equal(t, tt.expectedSubnetNames, actualSubnetNames, "subnetNames")
		})
	}
}

func Test_parseCmdCtx(t *testing.T) {
	tests := []struct {
		name           string
		cmdCtx         command.Context
		expectedParsed *parsedCmdCtx
		expectedErr    bool
	}{
		// --ns tests!
		{
			name:   "nsPassedAsArg",
			cmdCtx: command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "198.51.45.9:53",
							SubnetIP:         nil,
							Timeout:          0,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetToName:    nil,
				Dig:             nil,
				Stdout:          nil,
				GlobalTimeout:   2 * time.Second,
			},
		},
		{
			name:   "badNSPassedAsArg",
			cmdCtx: command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"badns"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr:    true,
			expectedParsed: nil,
		},
		{
			name:   "nsFromMap",
			cmdCtx: command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"nsFromMap"}, "--nameserver-map": map[string]string{"nsFromMap": "1.2.3.4:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         nil,
							Timeout:          0,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetToName:    nil,
				Dig:             nil,
				Stdout:          nil,
				GlobalTimeout:   2 * time.Second,
			},
		},
		{
			name:   "nsAll",
			cmdCtx: command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"all"}, "--nameserver-map": map[string]string{"nsFromMap": "1.2.3.4:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         nil,
							Timeout:          0,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetToName:    nil,
				Dig:             nil,
				Stdout:          nil,
				GlobalTimeout:   2 * time.Second,
			},
		},
		{
			name:   "namedNameserver",
			cmdCtx: command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"dns1.p09.nsone.net.:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "dns1.p09.nsone.net.:53",
							SubnetIP:         nil,
							Timeout:          0,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"dns1.p09.nsone.net.:53": "passed ns:port"},
				SubnetToName:    nil,
				Dig:             nil,
				Stdout:          nil,
				GlobalTimeout:   2 * time.Second,
			},
		},
		{
			name:           "namedNameserverErr",
			cmdCtx:         command.Context{Context: context.Background(), Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--nameserver": []string{"dns1.p09.nsone.net.53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--global-timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},
			expectedErr:    true,
			expectedParsed: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actualParsed, actualErr := parseCmdCtx(tt.cmdCtx)
			if tt.expectedErr {
				require.NotNil(t, actualErr)
				return
			} else {
				require.Nil(t, actualErr)
			}

			// Functions can't always compare false, so nil Dig out
			actualParsed.Dig = nil
			// I also don't want to mess with os Files
			actualParsed.Stdout = nil

			// Test subnets individually since they cannot be compared with '=='
			require.Equal(t, len(tt.expectedParsed.DigRepeatParams), len(actualParsed.DigRepeatParams))
			for i := 0; i < len(tt.expectedParsed.DigRepeatParams); i++ {
				if !tt.expectedParsed.DigRepeatParams[i].DigOneParams.SubnetIP.Equal(
					actualParsed.DigRepeatParams[i].DigOneParams.SubnetIP,
				) {
					t.Fatalf(
						"Expected equal subnets for index %d: %v != %v",
						i,
						tt.expectedParsed.DigRepeatParams[i].DigOneParams.SubnetIP,
						actualParsed.DigRepeatParams[i].DigOneParams.SubnetIP,
					)
				}
				// now set to nil!
				tt.expectedParsed.DigRepeatParams[i].DigOneParams.SubnetIP = nil
				actualParsed.DigRepeatParams[i].DigOneParams.SubnetIP = nil
			}
			require.Equal(t, tt.expectedParsed, actualParsed)

		})
	}
}
