package counter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringCounter(t *testing.T) {
	t.Parallel()
	c := NewStringCounter()
	c.Add("4")
	c.Add("1")
	c.Add("2")
	c.Add("3")
	c.Add("2")
	c.Add("3")

	actual := c.AsSortedSlice()

	expected := []StringCount{
		{
			String: "1",
			Count:  1,
		},
		{
			String: "2",
			Count:  2,
		},
		{
			String: "3",
			Count:  2,
		},
		{
			String: "4",
			Count:  1,
		},
	}

	require.Equal(t, expected, actual)

}
