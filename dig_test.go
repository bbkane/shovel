package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.bbkane.com/shovel/dig"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
)

func Test_parseCmdCtx(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		cmdCtx         command.Context
		expectedParsed *parsedCmdCtx
		expectedErr    bool
	}{
		{
			name: "noSubnet",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "subnetPassedAsArg",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "1.2.3.0",
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "badSubnetPassedAsArg",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "badSubnet",
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr:    true,
			expectedParsed: nil,
		},
		{
			name: "subnetFromMap",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "mysubnet",
				"--subnet-map", "mysubnet=3.4.5.0",
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "subnetAll",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "1.2.3.4:53",
				// no --ns-map
				"--subnet", "all",
				"--subnet-map", "subnetName=1.1.1.0",
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "nsPassedAsArg",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "badNSPassedAsArg",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "badns",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: true,
		},
		{
			name: "nsFromMap",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "nsFromMap",
				"--ns-map", "nsFromMap=1.2.3.4:53",
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "nsAll",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "all",
				"--ns-map", "nsFromMap=1.2.3.4:53",
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "namedNameserver",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "dns1.p09.nsone.net.:53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []dig.DigRepeatParams{
					{
						DigOneParams: dig.DigOneParams{
							FQDN:             "linkedin.com",
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
			name: "namedNameserverErr",
			args: []string{
				"shovel", "dig", "combine",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "dns1.p09.nsone.net.53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			app := buildApp()
			pr, err := app.Parse(tt.args, warg.LookupMap(nil))
			require.Nil(t, err)

			actualParsed, actualErr := parseCmdCtx(pr.Context)
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

func TestRunCLI(t *testing.T) {
	updateGolden := os.Getenv("SHOVEL_TEST_UPDATE_GOLDEN") != ""
	tests := []struct {
		name   string
		app    warg.App
		args   []string
		lookup warg.LookupFunc
	}{
		{
			name: "simple",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--mock-dig-func", "simple", // don't really dig!
				"--ns", "0.0.0.0:53",
				"--rtype", "A",
			},
			lookup: warg.LookupMap(nil),
		},
		{
			name: "twocount",
			app:  buildApp(),
			args: []string{"shovel", "dig", "combine",
				"--config", "notthere", // Hack so shovel doesn't try to read a config
				"--count", "2",
				"--fqdn", "linkedin.com",
				"--mock-dig-func", "twocount", // don't really dig!
				"--ns", "0.0.0.0:53",
				"--rtype", "A",
			},
			lookup: warg.LookupMap(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warg.GoldenTest(t, tt.app, tt.args, tt.lookup, updateGolden)
		})
	}
}
