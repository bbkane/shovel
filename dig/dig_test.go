package dig

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"go.bbkane.com/shovel/counter"
)

func Test_digOne(t *testing.T) {

	integrationTest := os.Getenv("SHOVEL_INTEGRATION_TEST") != ""
	if !integrationTest {
		t.Skipf("To run integration tests, run: SHOVEL_INTEGRATION_TEST=1 go test ./... ")
	}

	tests := []struct {
		name        string
		dig         DigOneFunc
		p           DigOneParams
		expected    []string
		expectedErr bool
	}{
		{
			name: "linkedinNoSubnet",
			dig:  DigOne,
			p: DigOneParams{
				Qname:            "linkedin.com",
				Rtype:            dns.TypeA,
				NameserverIPPort: "8.8.8.8:53",
				SubnetIP:         nil,
				Proto:            "udp",
				Timeout:          0,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: false,
		},
		{
			// Google nameserver doesn't work from China
			name: "linkedinChinaSubnet",
			dig:  DigOne,
			p: DigOneParams{
				Qname:            "linkedin.com",
				Rtype:            dns.TypeA,
				NameserverIPPort: "8.8.8.8:53",
				SubnetIP:         net.ParseIP("101.251.8.0"),
				Proto:            "udp",
				Timeout:          0,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: true,
		},
		{
			// Google nameserver doesn't work from China
			name: "nsName",
			dig:  DigOne,
			p: DigOneParams{
				Qname: "linkedin.com",
				Rtype: dns.TypeA,
				// This can end in '.' or not, it's fine!
				NameserverIPPort: "dns1.p09.nsone.net:53",
				SubnetIP:         nil,
				Proto:            "udp",
				Timeout:          0,
			},
			expected:    []string{"13.107.42.14"},
			expectedErr: false,
		},
		{
			name: "mock",
			dig: DigOneFuncMock(context.Background(), []DigOneResult{
				{
					Answers: []string{"hi"},
					Err:     nil,
				},
			}),
			p:           EmptyDigOneparams(),
			expected:    []string{"hi"},
			expectedErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, actualErr := tt.dig(context.Background(), tt.p)

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

func Test_digVaried(t *testing.T) {

	tests := []struct {
		name     string
		params   []DigRepeatParams
		dig      DigOneFunc
		expected []DigRepeatResult
	}{
		{
			name: "simple",
			params: []DigRepeatParams{
				{
					DigOneParams: EmptyDigOneparams(),
					Count:        1,
				},
			},
			dig: DigOneFuncMock(context.Background(), []DigOneResult{
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
			}),
			expected: []DigRepeatResult{
				{
					Answers: []counter.StringSliceCount{
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
			params: []DigRepeatParams{
				{
					DigOneParams: EmptyDigOneparams(),
					Count:        2,
				},
			},
			dig: DigOneFuncMock(context.Background(), []DigOneResult{
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
				{
					Answers: []string{"www.example.com"},
					Err:     nil,
				},
			}),
			expected: []DigRepeatResult{
				{
					Answers: []counter.StringSliceCount{
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

			actual := DigList(context.Background(), tt.params, tt.dig)
			require.Equal(t, tt.expected, actual)
		})
	}
}
