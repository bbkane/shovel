package main

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func Test_digOne(t *testing.T) {

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
