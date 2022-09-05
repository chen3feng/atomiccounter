# atomiccounter

English | [简体中文](README_zh.md)

[![License Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-red.svg)](COPYING)
[![Golang](https://img.shields.io/badge/Language-go1.18+-blue.svg)](https://go.dev/)
![Build Status](https://github.com/chen3feng/atomiccounter/actions/workflows/go.yml/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/chen3feng/atomiccounter/badge.svg?branch=master)](https://coveralls.io/github/chen3feng/atomiccounter?branch=master)
[![GoReport](https://goreportcard.com/badge/github.com/securego/gosec)](https://goreportcard.com/report/github.com/chen3feng/atomiccounter)
[![Go Reference](https://pkg.go.dev/badge/github.com/chen3feng/atomiccounter.svg)](https://pkg.go.dev/github.com/chen3feng/atomiccounter)

A High Performance Atomic Counter for Concurrent Write-More-Read-Less Scenario in Go.

Similar to [LongAdder](https://docs.oracle.com/javase/8/docs/api/java/util/concurrent/atomic/LongAdder.html) in Java, or
[ThreadCachedInt](https://github.com/facebook/folly/blob/main/folly/docs/ThreadCachedInt.md) in [folly](https://github.com/facebook/folly),
In scenarios of high concurrent writes but few reads, it can provide dozens of times the write performance than `sync/atomic`.

## Benchmark

per 100 calls.

Under MacOS with M1 Pro:

```console
goos: darwin
goarch: arm64
pkg: github.com/chen3feng/atomiccounter
BenchmarkNonAtomicAdd-10        47337121                22.14 ns/op
BenchmarkAtomicAdd-10             180942                 6861 ns/op
BenchmarkCounter-10             14871549                81.02 ns/op
```

Under Linux:

```console
goos: linux
goarch: amd64
pkg: github.com/chen3feng/atomiccounter
cpu: Intel(R) Xeon(R) Gold 6133 CPU @ 2.50GHz
BenchmarkNonAtomicAdd-16    	 9508723	       135.3 ns/op
BenchmarkAtomicAdd-16       	  582798	        2070 ns/op
BenchmarkCounter-16         	 4748263	       263.1 ns/op
```

From top to bottom are writing time-consuming of non-atomic (and thus unsafe), atomic, and `atomiccounter`.
It can be seen that in the case of high concurrent writes, `atomiccounter` is only a few times more slower
than non-atomic writes, but much faster than atomic writes.

But it is much slower reads:

```console
goos: darwin
goarch: arm64
pkg: github.com/chen3feng/atomiccounter
BenchmarkNonAtomicRead-10       1000000000               0.3112 ns/op
BenchmarkAtomicRead-10          1000000000               0.5336 ns/op
BenchmarkCounterRead-10         54609476                  21.20 ns/op
```

In addition, each `atomiccounter.Int64` object needs to consume 8K memory, so please only use it in a small number of
scenarios with a large number of concurrent writes but few reads, such as counting the number of requests.

## Compare with Similar Libraries

I found 3 similar Libraries in GitHub (the later 2 seems same):

- https://github.com/puzpuzpuz/xsync
- https://github.com/linxGnu/go-adder
- https://github.com/line/garr

And got the following benchmark result under the Apple M1 Pro chip.

```console
BenchmarkAdd_NonAtomic-10               49337793                22.02 ns/op
BenchmarkAdd_Atomic-10                    206678                 6854 ns/op
BenchmarkAdd_AtomicCounter-10           14658782                82.22 ns/op
BenchmarkAdd_XsyncCounter-10             9599529                144.6 ns/op
BenchmarkAdd_GoAdder-10                   825858                 1339 ns/op
BenchmarkAdd_GarrAdder-10                 915090                 1305 ns/op

BenchmarkRead_NonAtomic-10             263460258                4.087 ns/op
BenchmarkRead_Atomic-10                172530186                6.945 ns/op
BenchmarkRead_AtomicCounter-10           2793618                425.2 ns/op
BenchmarkRead_XSyncCounter-10            2396407                489.6 ns/op
BenchmarkRead_GoAdder-10                32101244                36.02 ns/op
BenchmarkRead_GarrAdder-10              29420326                35.40 ns/op
```

Obviously, `atomiccounter` is the fastest for concurrent writing.

See [atomiccounter_bench](https://github.com/chen3feng/atomiccounter_bench) for source code.

## Implementation

Data race is one of the biggest performance killers in multi-core programs. For counters with a large number of writes,
if ordinary atomic is used, the performance will be severely affected.

In scenarios with few reads, a common solution is to spread the writes across different variables and accumulate them when they are read.
Such as Java's LongAdder and folly ThreadCachedInt, and per-cpu in the Linux kernel are all used this this method.
Although the implementation details are different, the idea is similar.

At present, there is no well-known implementation for this kind of purpose in go, so I implemented this library.

To reduce memory footprint, multiple `Int64` objects may share same memory chunk.

### Memory Layout

An int64 array of multiple sizes of CPU [cache line size](https://en.wikipedia.org/wiki/CPU_cache#Cache_entries) becomes a cell.
A group of cells is called a chunk.

The size of the cell is an integer multiple of the cache line size of the CPU, and the first and last fields are paded
with blanks of the size of the cache line size, thus avoiding [false sharing](https://www.google.com/search?q=false+sharing).

The `chunk.lastIndex` member is used to record the last unused index for allocating the `Int64` object.

Each `Int64` object contains 2 fields: the chunk pointer and the index in the cell, so multiple `Int64` objects can share the same chunk,
but access elements with different indices in each cell.

### Allocate an `Int64` object

The address of the last created chunk is recorded in the global variable `lastChunk`. When an `Int64` object is created,
its `lastIndex` is increased. If it reachs the number of int64 in the cell,
it means that this chunk has been totally allocated and a new chunk needs to be created.

### Access an `Int64` object

Please first understand Go's [GMP](https://www.google.com/search?q=golang+GMP) scheduling model.

The best performance is to get the current subscript of `M` in Go and directly access the corresponding `cell`,
so that there will be no conflict between different `M`s, and even avoid using atomic operations.

But I haven't found a way to get the `M`'s subscript.

Therefore, this implementation uses the hash of the address of `M` as the subscript to access the cell,
and the measured effect is also quite good.

As long as the number of cells in each chunk is larger than the common number of CPU cores,
the impact of hash collisions can be reduced, so that different M will have a high probability
of accessing different cells.

When increasing the value of an `Int64` object, the hash of current `M`'s' address is used as the
subscript to obtain the corresponding cell in the chunk.
Then use the `Int64.index` member as a subscript to access the int64 array in this cell.

When reading, traverse the value indexed by `Int64.index` in all cell arrays and accumulated the value.

<!-- gomarkdoc:embed:start -->

<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# atomiccounter

```go
import "github.com/chen3feng/atomiccounter"
```

Package atomiccounter provides an atomic counter for high throughput concurrent writing and rare reading scenario.

<details><summary>Example</summary>
<p>

```go
package main

import (
	"fmt"
	"github.com/chen3feng/atomiccounter"
	"sync"
)

func main() {
	counter := atomiccounter.MakeInt64()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			counter.Inc()
			wg.Done()
		}()

	}
	wg.Wait()
	fmt.Println(counter.Read())
	counter.Set(0)
	fmt.Println(counter.Read())
	counter.Add(10)
	fmt.Println(counter.Read())
}
```

#### Output

```
100
0
10
```

</p>
</details>

## Index

- [type Int64](<#type-int64>)
  - [func MakeInt64() Int64](<#func-makeint64>)
  - [func (c *Int64) Add(n int64)](<#func-int64-add>)
  - [func (c *Int64) Inc()](<#func-int64-inc>)
  - [func (c *Int64) Read() int64](<#func-int64-read>)
  - [func (c *Int64) Set(n int64)](<#func-int64-set>)
  - [func (c *Int64) Swap(n int64) int64](<#func-int64-swap>)


## type [Int64](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L19-L22>)

Int64 is an int64 atomic counter.

```go
type Int64 struct {
    // contains filtered or unexported fields
}
```

### func [MakeInt64](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L57>)

```go
func MakeInt64() Int64
```

MakeInt64 creates a new Int64 object. Int64 objects must be created by this function, simply initialized doesn't work.

### func \(\*Int64\) [Add](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L72>)

```go
func (c *Int64) Add(n int64)
```

Add adds n to the counter.

### func \(\*Int64\) [Inc](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L78>)

```go
func (c *Int64) Inc()
```

Inc adds 1 to the counter.

### func \(\*Int64\) [Read](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L91>)

```go
func (c *Int64) Read() int64
```

Read return the current value. it is a little slow so it should not be called frequently. Th result is not guaranteed to be accurate in race conditions.

### func \(\*Int64\) [Set](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L83>)

```go
func (c *Int64) Set(n int64)
```

Set set the value of the counter to n.

### func \(\*Int64\) [Swap](<https://github.com/chen3feng/atomiccounter/blob/master/int64.go#L101>)

```go
func (c *Int64) Swap(n int64) int64
```

Swap returns the current value and swap it with n.



Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)


<!-- gomarkdoc:embed:end -->
