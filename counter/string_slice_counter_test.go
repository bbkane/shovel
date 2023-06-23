package counter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSliceCounter(t *testing.T) {
	c := NewStringSliceCounter()

	c.Add([]string{"a", "b"})
	c.Add([]string{"a"})
	c.Add([]string{"a"})
	c.Add([]string{"c"})

	actual := c.AsSortedSlice()

	expected := []StringSliceCount{
		{
			StringSlice: []string{"a"},
			Count:       2,
		},
		{
			StringSlice: []string{"a", "b"},
			Count:       1,
		},
		{
			StringSlice: []string{"c"},
			Count:       1,
		},
	}

	require.Equal(t, expected, actual)

}
