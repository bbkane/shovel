package main

import (
	"net"
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
		dig         digOneFunc
		p           digOneParams
		expected    []string
		expectedErr bool
	}{
		{
			name: "linkedinNoSubnet",
			dig:  digOne,
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
			dig:  digOne,
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
		{
			// Google nameserver doesn't work from China
			name: "nsName",
			dig:  digOne,
			p: digOneParams{
				FQDN:  "linkedin.com",
				Rtype: dns.TypeA,
				// This can end in '.' or not, it's fine!
				NameserverIPPort: "dns1.p09.nsone.net:53",
				SubnetIP:         nil,
				Timeout:          2 * time.Second,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: false,
		},
		{
			name: "mock",
			dig: digOneFuncMock([]digOneReturns{
				{
					Answers: []string{"hi"},
					Err:     nil,
				},
			}),
			p:           emptyDigOneparams(),
			expected:    []string{"hi"},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := tt.dig(tt.p)

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

func Test_cmdCtxToDigRepeatParams(t *testing.T) {
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
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
			},
			expectedErr: false,
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
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{"1.2.3.0": "passed ip"},
			},
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
			expectedErr:    true,
			expectedParsed: nil,
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
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{"3.4.5.0": "mysubnet"},
			},
		},
		{
			name: "subnetAll",
			args: []string{
				"shovel", "dig",
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
				DigRepeatParams: []digRepeatParams{
					{
						DigOneParams: digOneParams{
							FQDN:             "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "1.2.3.4:53",
							SubnetIP:         net.ParseIP("1.1.1.0"),
							Timeout:          2 * time.Second,
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"1.2.3.4:53": "passed ns:port"},
				SubnetNames:     map[string]string{"1.1.1.0": "subnetName"},
			},
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
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
			},
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
			expectedErr: true,
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
			expectedErr: false,
			expectedParsed: &parsedCmdCtx{
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetNames:     map[string]string{},
			},
		},
		{
			name: "nsAll",
			args: []string{
				"shovel", "dig",
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
				DigRepeatParams: []digRepeatParams{
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
				NameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
				SubnetNames:     map[string]string{},
			},
		},
		{
			name: "namedNameserver",
			args: []string{
				"shovel", "dig",
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
				DigRepeatParams: []digRepeatParams{
					{
						DigOneParams: digOneParams{
							FQDN:             "linkedin.com",
							Rtype:            dns.TypeA,
							NameserverIPPort: "dns1.p09.nsone.net.:53",
							SubnetIP:         nil,
							Timeout:          2 * time.Second,
						},
						Count: 1,
					},
				},
				NameserverNames: map[string]string{"dns1.p09.nsone.net.:53": "passed ns:port"},
				SubnetNames:     map[string]string{},
			},
		},
		{
			name: "namedNameserverErr",
			args: []string{
				"shovel", "dig",
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

func Test_digVaried(t *testing.T) {

	tests := []struct {
		name     string
		params   []digRepeatParams
		dig      digOneFunc
		expected []digRepeatResult
	}{
		{
			name: "simple",
			params: []digRepeatParams{
				{
					DigOneParams: emptyDigOneparams(),
					Count:        1,
				},
			},
			dig: digOneFuncMock([]digOneReturns{
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
			}),
			expected: []digRepeatResult{
				{
					Answers: []stringSliceCount{
						{
							StringSlice: []string{"www.example.com"},
							Count:       1,
						},
					},
					Errors: nil,
				},
			},
		},
		{
			name: "two",
			params: []digRepeatParams{
				{
					DigOneParams: emptyDigOneparams(),
					Count:        2,
				},
			},
			dig: digOneFuncMock([]digOneReturns{
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
			}),
			expected: []digRepeatResult{
				{
					Answers: []stringSliceCount{
						{
							StringSlice: []string{"www.example.com"},
							Count:       2,
						},
					},
					Errors: nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			actual := digVaried(tt.params, tt.dig)
			require.Equal(t, tt.expected, actual)
		})
	}
}
