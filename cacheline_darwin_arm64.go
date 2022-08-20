////go:build !(darwin && arm64)

package atomiccounter

// For apple M1 chip, the correct cache line size is 128 rather than 64 in x/sys/cpu:
//   % sysctl hw.cachelinesize
//   hw.cachelinesize: 128
const cacheLineSize = 128
