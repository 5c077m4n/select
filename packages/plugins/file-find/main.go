package main

//#include <stdlib.h>
import "C"

import (
	"encoding/json"
	"log"
	"path/filepath"
	"unsafe"
)

// ptrToString returns a string from WebAssembly compatible numeric types representing its pointer and length.
func ptrToString(ptr uint32, size uint32) string {
	unsafePtr := unsafe.Pointer(uintptr(ptr))
	return unsafe.String((*byte)(unsafePtr), size)
}

// stringToPtr returns a pointer and size pair for the given string in a way compatible with WebAssembly numeric types.
// The returned pointer aliases the string hence the string must be kept alive until ptr is no longer needed.
func stringToPtr(s string) (uint32, uint32) {
	ptr := unsafe.Pointer(unsafe.StringData(s))
	return uint32(uintptr(ptr)), uint32(len(s))
}

// stringToLeakedPtr returns a pointer and size pair for the given string in a way compatible with WebAssembly numeric types.
// The pointer is not automatically managed by TinyGo hence it must be freed by the host.
func stringToLeakedPtr(s string) (uint32, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return uint32(uintptr(ptr)), uint32(size)
}

func _search(glob string) string {
	matches, err := filepath.Glob(glob)
	if err != nil {
		log.Panicf("Could not complete the file search: %v", err)
	}

	jsonb, err := json.Marshal(matches)
	if err != nil {
		log.Panic(err)
	}

	return string(jsonb)
}

// search is a WebAssembly export that accepts a string pointer (linear memory offset) and returns a pointer/size pair packed into a uint64.
//
// NOTE: This uses a uint64 instead of two result values for compatibility with WebAssembly 1.0.
//
//export search
func search(ptr, size uint32) (ptrSize uint64) {
	name := ptrToString(ptr, size)
	g := _search(name)
	ptr, size = stringToLeakedPtr(g)
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

func main() {}
