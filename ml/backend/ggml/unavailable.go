//go:build !cgo

package ggml

import "errors"

// ErrUnavailable reports that the GGML backend is not available in this build.
// Cognitive schedulers should treat this as a capability boundary and route work
// to another backend rather than failing the whole event loop.
var ErrUnavailable = errors.New("ggml backend unavailable: cgo is disabled")

// Available reports whether the native GGML backend was compiled into this build.
func Available() bool { return false }
