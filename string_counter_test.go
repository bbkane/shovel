package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringCounter(t *testing.T) {
	c := newStringCounter()
	c.Add("1")
	c.Add("2")
	c.Add("2")

	actualStrs := []string{}
	actualCounts := []int{}
	for i := c.IterCountDescending(); i.HasNext(); {
		str, count := i.Next()
		actualStrs = append(actualStrs, str)
		actualCounts = append(actualCounts, count)
	}
	expectedCounts := []int{2, 1}
	expectedStrs := []string{"2", "1"}
	require.Equal(t, expectedCounts, actualCounts)
	require.Equal(t, expectedStrs, actualStrs)
}
