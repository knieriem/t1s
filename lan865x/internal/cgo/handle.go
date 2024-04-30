//go:build !tinygo

package cgo

import (
	"runtime/cgo"
)

type Handle = cgo.Handle

func NewHandle(v any) Handle {
	return cgo.NewHandle(v)
}
