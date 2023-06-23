package counter

import "sort"

type StringCounter struct {
	counter map[string]int
}

func NewStringCounter() StringCounter {
	return StringCounter{
		counter: make(map[string]int),
	}
}

func (c *StringCounter) Add(str string) {
	c.counter[str]++
}

type StringCount struct {
	String string
	Count  int
}

// AsSortedSlice returns the counter as a slice sorted by name asc, count asc
func (c *StringCounter) AsSortedSlice() []StringCount {
	var ret []StringCount
	for key, val := range c.counter {
		ret = append(ret, StringCount{
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
