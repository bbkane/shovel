package main

import (
	"net"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/command"
)

func Test_digOne(t *testing.T) {

	integrationTest := os.Getenv("SHOVEL_INTEGRATION_TEST") != ""
	if !integrationTest {
		t.Skipf("To run integration tests, run: SHOVEL_INTEGRATION_TEST=1 go test ./... ")
	}

	tests := []struct {
		name        string
		p           digOneParams
		expected    []string
		expectedErr bool
	}{
		{
			name: "linkedinNoSubnet",
			p: digOneParams{
				FQDN:             "linkedin.com",
				Rtype:            dns.TypeA,
				NameserverIPPort: "8.8.8.8:53",
				SubnetIP:         nil,
				Timeout:          2 * time.Second,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: false,
		},
		{
			// Google nameserver doesn't work from China
			name: "linkedinChinaSubnet",
			p: digOneParams{
				FQDN:             "linkedin.com",
				Rtype:            dns.TypeA,
				NameserverIPPort: "8.8.8.8:53",
				SubnetIP:         net.ParseIP("101.251.8.0"),
				Timeout:          2 * time.Second,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := digOne(tt.p)

			if tt.expectedErr {
				require.NotNil(t, actualErr)
				return
			} else {
				require.Nil(t, actualErr)
			}

			require.Equal(t, tt.expected, actual)

		})
	}
}

func Test_cmdCtxToDigOneparams(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		cmdCtx         command.Context
		expectedParams []digRepeatParams
		expectedNames  *nameMaps
		expectedErr    bool
	}{
		{
			name: "noSubnet",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: []digRepeatParams{
				{
					DigOneParams: digOneParams{
						FQDN:             "linkedin.com",
						Rtype:            dns.TypeA,
						NameserverIPPort: "198.51.45.9:53",
						SubnetIP:         nil,
						Timeout:          2 * time.Second,
					},
					Count: 1,
				},
			},
			expectedNames: &nameMaps{NameserverNames: map[string]string{"198.51.45.9:53": "passed ip:port"}, SubnetNames: map[string]string{}},
			expectedErr:   false,
		},
		{
			name: "subnetPassedAsArg",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "1.2.3.0",
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: []digRepeatParams{
				{
					DigOneParams: digOneParams{
						FQDN:             "linkedin.com",
						Rtype:            dns.TypeA,
						NameserverIPPort: "198.51.45.9:53",
						SubnetIP:         net.ParseIP("1.2.3.0"),
						Timeout:          2 * time.Second,
					},
					Count: 1,
				},
			},
			expectedErr:   false,
			expectedNames: &nameMaps{NameserverNames: map[string]string{"198.51.45.9:53": "passed ip:port"}, SubnetNames: map[string]string{"1.2.3.0": "passed ip"}},
		},
		{
			name: "badSubnetPassedAsArg",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "badSubnet",
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: nil,
			expectedErr:    true,
			expectedNames:  nil,
		},
		{
			name: "subnetFromMap",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				"--subnet", "mysubnet",
				"--subnet-map", "mysubnet=3.4.5.0",
				"--timeout", "2s",
			},
			expectedParams: []digRepeatParams{
				{
					DigOneParams: digOneParams{
						FQDN:             "linkedin.com",
						Rtype:            dns.TypeA,
						NameserverIPPort: "198.51.45.9:53",
						SubnetIP:         net.ParseIP("3.4.5.0"),
						Timeout:          2 * time.Second,
					},
					Count: 1,
				},
			},
			expectedErr:   false,
			expectedNames: &nameMaps{NameserverNames: map[string]string{"198.51.45.9:53": "passed ip:port"}, SubnetNames: map[string]string{"3.4.5.0": "mysubnet"}},
		},
		// --ns tests!
		{
			name: "nsPassedAsArg",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "198.51.45.9:53",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: []digRepeatParams{
				{
					DigOneParams: digOneParams{
						FQDN:             "linkedin.com",
						Rtype:            dns.TypeA,
						NameserverIPPort: "198.51.45.9:53",
						SubnetIP:         nil,
						Timeout:          2 * time.Second,
					},
					Count: 1,
				},
			},
			expectedErr:   false,
			expectedNames: &nameMaps{NameserverNames: map[string]string{"198.51.45.9:53": "passed ip:port"}, SubnetNames: map[string]string{}},
		},
		{
			name: "badNSPassedAsArg",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "badns",
				// no --ns-map
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: nil,
			expectedErr:    true,
			expectedNames:  nil,
		},
		{
			name: "nsFromMap",
			args: []string{
				"shovel", "dig",
				"--count", "1",
				"--fqdn", "linkedin.com",
				"--rtype", "A",
				"--ns", "nsFromMap",
				"--ns-map", "nsFromMap=1.2.3.4:53",
				// no --subnet
				// no --subnet-map
				"--timeout", "2s",
			},
			expectedParams: []digRepeatParams{
				{
					DigOneParams: digOneParams{
						FQDN:             "linkedin.com",
						Rtype:            dns.TypeA,
						NameserverIPPort: "1.2.3.4:53",
						SubnetIP:         nil,
						Timeout:          2 * time.Second,
					},
					Count: 1,
				},
			},
			expectedErr:   false,
			expectedNames: &nameMaps{NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"}, SubnetNames: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			app := buildApp()
			pr, err := app.Parse(tt.args, warg.LookupMap(nil))
			require.Nil(t, err)

			actualParams, actualNames, actualErr := cmdCtxToDigRepeatParams(pr.Context)
			if tt.expectedErr {
				require.NotNil(t, actualErr)
				return
			} else {
				require.Nil(t, actualErr)
			}

			require.Equal(t, tt.expectedNames, actualNames)

			// NOTE: net.IP is a slice of bytes and can have multiple []byte representations
			// so let's "normalize" them :)
			for i := 0; i < len(tt.expectedParams); i++ {
				if tt.expectedParams[i].DigOneParams.SubnetIP != nil {
					tt.expectedParams[i].DigOneParams.SubnetIP = netip.MustParseAddr(
						tt.expectedParams[i].DigOneParams.SubnetIP.String(),
					).AsSlice()
				}
			}
			for i := 0; i < len(actualParams); i++ {
				if actualParams[i].DigOneParams.SubnetIP != nil {
					actualParams[i].DigOneParams.SubnetIP = netip.MustParseAddr(
						actualParams[i].DigOneParams.SubnetIP.String(),
					).AsSlice()
				}
			}

			require.Equal(t, tt.expectedParams, actualParams)

		})
	}
}
