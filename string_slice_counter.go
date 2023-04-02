package main

import (
	"hash/fnv"
	"sort"

	"golang.org/x/exp/slices"
)

type stringSliceCounter struct {
	hashToSlice map[uint64][]string
	hashToCount map[uint64]int
}

func newStringSliceCounter() stringSliceCounter {
	return stringSliceCounter{
		hashToSlice: make(map[uint64][]string),
		hashToCount: make(map[uint64]int),
	}
}

func (c *stringSliceCounter) Add(slice []string) {
	// https://pkg.go.dev/hash@go1.20.2#example-package-BinaryMarshaler
	// https://stackoverflow.com/a/72562519/2958070

	hasher := fnv.New64()
	for _, s := range slice {
		hasher.Write([]byte(s))
	}
	hash := hasher.Sum64()

	c.hashToCount[hash]++
	c.hashToSlice[hash] = slice // TODO: how expensive is this?
}

type stringSliceCount struct {
	StringSlice []string
	Count       int
}

func (c *stringSliceCounter) AsSortedSlice() []stringSliceCount {
	var ret []stringSliceCount
	for key := range c.hashToCount {
		ret = append(ret, stringSliceCount{
			StringSlice: c.hashToSlice[key],
			Count:       c.hashToCount[key],
		})
	}

	sort.Slice(
		ret,
		func(i, j int) bool {

			slicesCompare := slices.Compare(ret[i].StringSlice, ret[j].StringSlice)
			if slicesCompare == 0 {
				return ret[i].Count < ret[j].Count
			}
			return slicesCompare < 0
		},
	)
	return ret
}
