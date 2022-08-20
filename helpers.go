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

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
