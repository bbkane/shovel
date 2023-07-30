package digcombine

import (
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg/command"
)

func Test_parseCmdCtx(t *testing.T) {
	tests := []struct {
		name           string
		cmdCtx         command.Context
		expectedParsed *parsedCmdCtx
		expectedErr    bool
	}{
		{
			name:   "noSubnet",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "198.51.45.9:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
				Dig:             nil,
				Stdout:          nil,
			},
			expectedErr: false,
		},
		{
			name:   "subnetPassedAsArg",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--subnet": []string{"1.2.3.0"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "198.51.45.9:53",
							SubnetIP:         net.ParseIP("1.2.3.0"),
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{"1.2.3.0": "passed ip"},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "badSubnetPassedAsArg",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--subnet": []string{"badSubnet"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr:    true,
			expectedParsed: nil,
		},
		{
			name:   "subnetFromMap",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--subnet": []string{"mysubnet"}, "--subnet-map": map[string]netip.Addr{"mysubnet": netip.MustParseAddr("3.4.5.0")}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "198.51.45.9:53",
							SubnetIP:         net.ParseIP("3.4.5.0"),
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{"3.4.5.0": "mysubnet"},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "subnetAll",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"1.2.3.4:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--subnet": []string{"all"}, "--subnet-map": map[string]netip.Addr{"subnetName": netip.MustParseAddr("1.1.1.0")}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         net.ParseIP("1.1.1.0"),
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "passed ns:port"},
				SubnetNames:     map[string]string{"1.1.1.0": "subnetName"},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		// --ns tests!
		{
			name:   "nsPassedAsArg",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"198.51.45.9:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "198.51.45.9:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "badNSPassedAsArg",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"badns"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: true,
		},
		{
			name:   "nsFromMap",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"nsFromMap"}, "--ns-map": map[string]string{"nsFromMap": "1.2.3.4:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetNames:     map[string]string{},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "nsAll",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"all"}, "--ns-map": map[string]string{"nsFromMap": "1.2.3.4:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetNames:     map[string]string{},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "namedNameserver",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"dns1.p09.nsone.net.:53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							Qname:            "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "dns1.p09.nsone.net.:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
							Proto:            "udp",
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"dns1.p09.nsone.net.:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
				Dig:             nil,
				Stdout:          nil,
			},
		},
		{
			name:   "namedNameserverErr",
			cmdCtx: command.Context{Flags: command.PassedFlags{"--color": "auto", "--config": "", "--count": 1, "--qname": []string{"linkedin.com"}, "--help": "default", "--mock-dig-func": "none", "--ns": []string{"dns1.p09.nsone.net.53"}, "--protocol": "udp", "--rtype": []string{"A"}, "--timeout": 2 * time.Second}, Stderr: nil, Stdout: nil},

			expectedErr: true,
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
