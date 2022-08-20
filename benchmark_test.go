package atomiccounter_test

import (
	"sync/atomic"
	"testing"

	"github.com/chen3feng/atomiccounter"
)

//go:noinline
func add(count *int64, n int) {
	for i := 0; i < n; i++ {
		*count++
	}
}

func atomicAdd(count *int64, n int) {
	for i := 0; i < n; i++ {
		atomic.AddInt64(count, 1)
	}
}

const (
	batchSize = 100
)

func BenchmarkNonAtomicAdd(b *testing.B) {
	count := int64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			add(&count, batchSize)
		}
	})
}

func BenchmarkAtomicAdd(b *testing.B) {
	count := int64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomicAdd(&count, batchSize)
		}
	})
}

func BenchmarkCounter(b *testing.B) {
	counter := atomiccounter.NewInt64()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < batchSize; i++ {
				counter.Add(1)
			}
		}
	})
}

func BenchmarkNonAtomicRead(b *testing.B) {
	count := int64(0)
	for i := 0; i < b.N; i++ {
		_ = count
	}
}

func BenchmarkAtomicRead(b *testing.B) {
	count := int64(0)
	for i := 0; i < b.N; i++ {
		_ = atomic.LoadInt64(&count)
	}
}

func BenchmarkCounterRead(b *testing.B) {
	counter := atomiccounter.NewInt64()
	for i := 0; i < b.N; i++ {
		_ = counter.Read()
	}
}
