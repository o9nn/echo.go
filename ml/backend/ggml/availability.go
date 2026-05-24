//go:build cgo

package ggml

// Available reports whether the native GGML backend was compiled into this build.
func Available() bool { return true }
