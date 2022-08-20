package atomiccounter

import (
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	// Number of cells in each chunk. the size is larger than usual CPU cores to reduce hash conflict.
	numChunkCells = 64
	// Number of int64s in each cell. there are 2 pad fields, it should not be too small to avoid waste memory.
	cellCapacity = 6 * unsafe.Sizeof(cpu.CacheLinePad{}) / 8
)

// Int64 is an int64 atomic counter.
type Int64 struct {
	cells *[numChunkCells]cell
	index uintptr // index to the n array in each cell
}

// cell is a value container for each cpu core.
type cell struct {
	// We have no way to ensure cache line aligned allocations. so the 2 pads are necessary.
	_ cpu.CacheLinePad
	// The sizeof(cell) shuld be integer multiple of sizeof(cpu.CacheLinePad) to avoid false sharing.
	n [cellCapacity]int64
	_ cpu.CacheLinePad
}

// chunk is used to saves memory by sharing cells between multiple Int64s
type chunk struct {
	cells     [numChunkCells]cell
	nextIndex uintptr
}

// allocate a new Int64 from the chunk.
func (st *chunk) allocate() Int64 {
	for i := atomic.LoadUintptr(&st.nextIndex); i+1 < cellCapacity && atomic.CompareAndSwapUintptr(&st.nextIndex, i, i+1); {
		return Int64{&st.cells, i}
	}
	return Int64{nil, 0}
}

// newChunk creates a new chunk.
func newChunk() *chunk {
	return &chunk{}
}

// the last create chunk. atomic.Pointer is better but it's unavailable until go1.19.
var lastChunk atomic.Value

// MakeInt64 creates a new Int64 object.
// Int64 objects must be created by this function, simply initialized doesn't work.
func MakeInt64() Int64 {
	ch, ok := lastChunk.Load().(*chunk)
	if ok {
		ret := ch.allocate()
		if ret.cells != nil {
			return ret
		}
	}
	ch = newChunk()
	ret := ch.allocate() // Must be success because there are no race
	lastChunk.Store(ch)
	return ret
}

// Add adds n to the counter.
func (c *Int64) Add(n int64) {
	idx := threadHash() % numChunkCells
	atomic.AddInt64(&c.cells[idx].n[c.index], n)
}

// Inc adds 1 to the counter.
func (c *Int64) Inc() {
	c.Add(1)
}

// Set set the value of the counter to n.
func (c *Int64) Set(n int64) {
	c.Swap(n)
}

// Read return the current value. it is a little slow so it should not be called frequently.
// Th result is not guaranteed to be accurate in race conditions.
//go:norace
func (c *Int64) Read() int64 {
	total := int64(0)
	for i := range c.cells {
		// total += atomic.LoadInt64(&c.cells[i].n[0])
		total += c.cells[i].n[c.index]
	}
	return total
}

// Swap returns the current value and swap it with n.
func (c *Int64) Swap(n int64) int64 {
	total := atomic.SwapInt64(&c.cells[0].n[c.index], n)
	for i := range c.cells[1:] {
		total += atomic.SwapInt64(&c.cells[i+1].n[c.index], 0)
	}
	return total
}
