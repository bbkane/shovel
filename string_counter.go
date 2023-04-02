package main

import "sort"

type stringCounter struct {
	counter map[string]int
}

func newStringCounter() stringCounter {
	return stringCounter{
		counter: make(map[string]int),
	}
}

func (c *stringCounter) Add(str string) {
	c.counter[str]++
}

type stringCount struct {
	String string
	Count  int
}

// AsSortedSlice returns the counter as a slice sorted by name asc, count asc
func (c *stringCounter) AsSortedSlice() []stringCount {
	var ret []stringCount
	for key, val := range c.counter {
		ret = append(ret, stringCount{
			String: key,
			Count:  val,
		})
	}
	sort.Slice(
		ret,
		func(i, j int) bool {
			// NOTE: strings are guaranteed to be unique so not sure this is needed
			if ret[i].String == ret[j].String {
				return ret[i].Count < ret[j].Count
			}
			return ret[i].String < ret[j].String
		},
	)
	return ret
}
