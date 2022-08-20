package atomiccounter

import (
	"sync/atomic"
)

const (
	maxCpus = 64
)

// Int64 is an int64 atomic counter.
type Int64 struct {
	_     noCopy
	cells [maxCpus]cell
}

type cell struct {
	n [cacheLineSize / 8]int64
}

// NewInt64 creates a new Int64 object.
func NewInt64() *Int64 {
	return &Int64{}
}

// Add adds n to the counter.
func (c *Int64) Add(n int64) {
	idx := threadHash() % maxCpus
	atomic.AddInt64(&c.cells[idx].n[0], n)
}

// Inc adds 1 to the counter.
func (c *Int64) Inc() {
	c.Add(1)
}

// Set set the value of the counter to n.
func (c *Int64) Set(n int64) {
	c.Swap(n)
}

// Load return the current value.
func (c *Int64) Load() int64 {
	total := int64(0)
	for i := range c.cells {
		total += atomic.LoadInt64(&c.cells[i].n[0])
	}
	return total
}

// Swap returns the current value and swap it with n.
func (c *Int64) Swap(n int64) int64 {
	total := int64(0)
	for i := range c.cells {
		total += atomic.SwapInt64(&c.cells[i].n[0], 0)
	}
	return total
}
