//go:build linux

package main

import (
	"fmt"
	"os"
	"syscall"
)

// openDirect opens the named file for reading with O_DIRECT, which instructs
// the kernel to bypass the page cache and transfer data directly between the
// storage device and the user-space buffer.
//
// O_DIRECT imposes alignment requirements: the buffer address, file offset, and
// transfer length must all be multiples of the device's logical block size
// (commonly 512 B, but 4 KiB for many SSDs). newBuffer satisfies this via mmap.
func openDirect(name string) (*os.File, error) {
	fd, err := syscall.Open(name, syscall.O_RDONLY|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, fmt.Errorf("open O_DIRECT %q: %w", name, err)
	}
	return os.NewFile(uintptr(fd), name), nil
}

// newBuffer allocates a page-aligned buffer of the given size via anonymous
// mmap. mmap always returns page-aligned memory, satisfying O_DIRECT's
// alignment requirements even for large transfers.
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

// dropPageCache writes "3" to /proc/sys/vm/drop_caches to evict all page cache,
// dentry, and inode caches. This requires root privileges.
func dropPageCache() error {
	// Flush dirty pages to disk before dropping.
	syscall.Sync()

	f, err := os.OpenFile("/proc/sys/vm/drop_caches", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("drop_caches (run as root?): %w", err)
	}
	defer f.Close()
	_, err = f.WriteString("3\n")
	return err
}

// directIOMethod returns a human-readable name for the direct-I/O mechanism.
func directIOMethod() string {
	return "O_DIRECT"
}

// pageCacheDropMethod returns the name of the page-cache clearing mechanism.
func pageCacheDropMethod() string {
	return "echo 3 > /proc/sys/vm/drop_caches"
}
