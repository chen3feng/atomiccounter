//go:build !(darwin && arm64)

package atomiccounter

import (
	"golang.org/x/sys/cpu"
	"unsafe"
)

const cacheLineSize = unsafe.Sizeof(cpu.CacheLinePad{})
