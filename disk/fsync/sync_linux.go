//go:build linux

package main

import (
	"os"
	"syscall"
)

// syncFile calls fdatasync(2) which flushes file data to the storage device
// without the metadata flush overhead of fsync(2). Since the benchmark
// overwrites existing data without changing the file size, metadata flush
// is unnecessary — this matches the standard database sync path.
func syncFile(f *os.File) error {
	return syscall.Fdatasync(int(f.Fd()))
}

// syncMethod returns a human-readable name for the sync mechanism.
func syncMethod() string {
	return "fdatasync"
}
