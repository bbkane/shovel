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

type stringCounterIterator struct {
	reverseCounter map[int]string
	sortedCounts   []int
	currentIndex   int
}

func (c *stringCounter) IterCountDescending() stringCounterIterator {
	iter := stringCounterIterator{
		reverseCounter: make(map[int]string, len(c.counter)),
		sortedCounts:   make([]int, 0, len(c.counter)),
		currentIndex:   0,
	}
	for k, v := range c.counter {
		iter.reverseCounter[v] = k
		iter.sortedCounts = append(iter.sortedCounts, v)
	}
	sort.Slice(iter.sortedCounts, func(i, j int) bool {
		return iter.sortedCounts[i] > iter.sortedCounts[j]
	})
	return iter
}

func (i *stringCounterIterator) HasNext() bool {
	return i.currentIndex < len(i.sortedCounts)
}

func (i *stringCounterIterator) Next() (string, int) {
	curCount := i.sortedCounts[i.currentIndex]
	curStr := i.reverseCounter[curCount]
	i.currentIndex++
	return curStr, curCount
}
