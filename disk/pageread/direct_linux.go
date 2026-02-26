//go:build linux

package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
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

// dropPageCache evicts the cached pages for the named file from the kernel
// page cache using posix_fadvise(POSIX_FADV_DONTNEED).
//
// Unlike madvise(MADV_DONTNEED) on a MAP_SHARED mapping — which only removes
// the pages from the calling process's page table without touching the kernel
// page cache — posix_fadvise(FADV_DONTNEED) calls invalidate_mapping_pages()
// inside the kernel, which directly walks the file's page cache and releases
// its clean pages synchronously. Since the benchmark dataset is never written
// during a run, all pages are clean and the drop is immediate.
//
// No elevated privileges are required.
//
// Fallback: write "3" to /proc/sys/vm/drop_caches (requires CAP_SYS_ADMIN).
func dropPageCache(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// posix_fadvise(POSIX_FADV_DONTNEED) – no root required.
	// unix.Fadvise is available for all Linux architectures via golang.org/x/sys.
	if err := unix.Fadvise(int(f.Fd()), 0, 0, unix.FADV_DONTNEED); err == nil {
		return nil
	}

	// Fallback: global cache drop (requires root / CAP_SYS_ADMIN).
	syscall.Sync()
	df, err := os.OpenFile("/proc/sys/vm/drop_caches", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("fadvise unavailable and drop_caches requires root: %w", err)
	}
	defer df.Close()
	_, err = df.WriteString("3\n")
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
