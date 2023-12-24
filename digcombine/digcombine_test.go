package digcombine

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNameservers(t *testing.T) {
	tests := []struct {
		name                    string
		passedNameservers       []string
		nameserverMap           map[string]string
		expectedNameservers     []string
		expectedNameserverNames map[string]string
		expectedErr             bool
	}{
		{
			name:                    "nsPassedAsArg",
			passedNameservers:       []string{"198.51.45.9:53"},
			nameserverMap:           nil,
			expectedNameservers:     []string{"198.51.45.9:53"},
			expectedNameserverNames: map[string]string{"198.51.45.9:53": "passed ns:port"},
			expectedErr:             false,
		},
		{
			name:                    "badNSPassedAsArg",
			passedNameservers:       []string{"badns"},
			nameserverMap:           nil,
			expectedNameservers:     nil,
			expectedNameserverNames: nil,
			expectedErr:             true,
		},
		{
			name:                    "nsFromMap",
			passedNameservers:       []string{"nsFromMap"},
			nameserverMap:           map[string]string{"nsFromMap": "1.2.3.4:53"},
			expectedNameservers:     []string{"1.2.3.4:53"},
			expectedNameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
			expectedErr:             false,
		},
		{
			name:                    "nsAll",
			passedNameservers:       []string{"all"},
			nameserverMap:           map[string]string{"nsFromMap": "1.2.3.4:53"},
			expectedNameservers:     []string{"1.2.3.4:53"},
			expectedNameserverNames: map[string]string{"1.2.3.4:53": "nsFromMap"},
			expectedErr:             false,
		},
		{
			name:                    "namedNameserverErr",
			passedNameservers:       []string{"dns1.p09.nsone.net.53"},
			nameserverMap:           nil,
			expectedNameservers:     nil,
			expectedNameserverNames: nil,
			expectedErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualNameservers, actualNameserverNames, actualErr := ParseNameservers(tt.passedNameservers, tt.nameserverMap)
			if tt.expectedErr {
				require.NotNil(t, actualErr)
			} else {
				require.Nil(t, actualErr)
			}
			require.Equal(t, tt.expectedNameservers, actualNameservers, "nameservers")
			require.Equal(t, tt.expectedNameserverNames, actualNameserverNames)
		})
	}
}

func TestParseSubnets(t *testing.T) {
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
