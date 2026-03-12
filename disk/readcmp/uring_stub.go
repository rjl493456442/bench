//go:build !linux

package main

import "fmt"

// UringEngine is not available on non-Linux platforms.
type UringEngine struct{}

func newUringEngine() (Engine, error) {
	return nil, fmt.Errorf("io_uring is only available on Linux")
}

func (e *UringEngine) Name() string                                    { return "uring" }
func (e *UringEngine) Run(_ RunConfig, _ []int64) (*RunResult, error)  { return nil, fmt.Errorf("unsupported") }
func (e *UringEngine) Features() map[string]bool                       { return nil }

func probeIOUring() (bool, uint32)        { return false, 0 }
func ioUringFeaturesString(_ uint32) string { return "unsupported" }
