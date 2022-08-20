package atomiccounter

import (
	"unsafe"
)

//go:linkname getm runtime.getm
func getm() uintptr

//go:noescape
//go:linkname memhash runtime.memhash
func memhash(p unsafe.Pointer, h, s uintptr) uintptr

func threadHash() uint {
	m := getm()
	// #nosec G103
	return uint(memhash(unsafe.Pointer(&m), 0, unsafe.Sizeof(m)))
}
