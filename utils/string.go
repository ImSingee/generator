package utils

import (
	"unsafe"
)

func ToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}
