//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// F_NOCACHE (0x30 = 48) instructs the kernel to not cache data for this fd.
// This is the macOS equivalent of O_DIRECT on Linux.
const fNOCACHE = 48

// openDirect opens the named file for reading with the page cache disabled via
// fcntl(F_NOCACHE). Every read will go directly to the storage device.
func openDirect(name string) (*os.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	// Disable page-cache caching for this file descriptor.
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), uintptr(fNOCACHE), 1); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("fcntl(F_NOCACHE): %w", errno)
	}
	return f, nil
}

// newBuffer allocates a page-aligned buffer of the given size via mmap.
// mmap-backed memory is always page-aligned, which is required on Linux
// (kept here for API consistency across platforms).
func newBuffer(size int) []byte {
	buf, err := syscall.Mmap(
		-1, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE,
	)
	if err != nil {
		panic(fmt.Sprintf("mmap(%d): %v", size, err))
	}
	return buf
}

// freeBuffer releases the mmap-allocated buffer.
func freeBuffer(buf []byte) {
	_ = syscall.Munmap(buf)
}

// dropPageCache attempts to flush the macOS unified buffer cache via `purge`.
// The `purge` binary is installed by default on macOS but may require
// administrator privileges on some versions.
func dropPageCache() error {
	return exec.Command("purge").Run()
}

// directIOMethod returns a human-readable name for the direct-I/O mechanism.
func directIOMethod() string {
	return "fcntl(F_NOCACHE)"
}

// pageCacheDropMethod returns the name of the page-cache clearing mechanism.
func pageCacheDropMethod() string {
	return "purge"
}
