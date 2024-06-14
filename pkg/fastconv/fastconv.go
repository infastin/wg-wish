package fastconv

import (
	"unsafe"
)

// Quickly convert a slice of bytes to a string.
// The bytes argument can be nil.
// Must not be used if the passed slice may change.
func String(bytes []byte) string {
	return unsafe.String(unsafe.SliceData(bytes), len(bytes))
}

// Quickly convert a string to a slice of bytes.
// The str argument can be nil.
// Must not be used if the returned slice may change.
func Bytes(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}
