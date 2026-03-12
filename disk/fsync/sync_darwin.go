//go:build darwin

package main

import (
	"fmt"
	"os"
	"syscall"
)

// F_FULLFSYNC (51) asks the drive to flush its write-back cache to stable
// storage. This is the only reliable way to ensure durability on macOS;
// plain fsync(2) only flushes to the drive's volatile write cache.
const fFULLFSYNC = 51

// syncFile issues F_FULLFSYNC via fcntl, ensuring data reaches stable storage.
func syncFile(f *os.File) error {
	_, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), uintptr(fFULLFSYNC), 0)
	if errno != 0 {
		return fmt.Errorf("fcntl(F_FULLFSYNC): %w", errno)
	}
	return nil
}

// syncMethod returns a human-readable name for the sync mechanism.
func syncMethod() string {
	return "fcntl(F_FULLFSYNC)"
}
