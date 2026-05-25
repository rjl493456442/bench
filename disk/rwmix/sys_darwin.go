//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// ── File opening ──────────────────────────────────────────────────────────────

const fNOCACHE = 48 // F_NOCACHE: bypass the unified buffer cache for this fd

func setNoCache(f *os.File) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), uintptr(fNOCACHE), 1); errno != 0 {
		return fmt.Errorf("fcntl(F_NOCACHE): %w", errno)
	}
	return nil
}

func openDirect(name string) (*os.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	if err := setNoCache(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func openDirectRW(name string) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	if err := setNoCache(f); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func openBuffered(name string) (*os.File, error)   { return os.Open(name) }
func openBufferedRW(name string) (*os.File, error) { return os.OpenFile(name, os.O_RDWR, 0) }

func directIOMethod() string { return "fcntl(F_NOCACHE)" }

// ── Aligned buffers ───────────────────────────────────────────────────────────

func newBuffer(size int) []byte {
	buf, err := syscall.Mmap(-1, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Sprintf("mmap(%d): %v", size, err))
	}
	return buf
}

func freeBuffer(buf []byte) { _ = syscall.Munmap(buf) }

// ── Durability ────────────────────────────────────────────────────────────────

const fFULLFSYNC = 51 // F_FULLFSYNC: flush the drive's write cache to stable media

// syncFile issues F_FULLFSYNC, the only macOS call that guarantees data reaches
// stable storage; plain fsync(2) only flushes to the drive's volatile cache.
func syncFile(f *os.File) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), uintptr(fFULLFSYNC), 0); errno != 0 {
		return fmt.Errorf("fcntl(F_FULLFSYNC): %w", errno)
	}
	return nil
}

func syncMethod() string { return "fcntl(F_FULLFSYNC)" }

// ── Page cache ────────────────────────────────────────────────────────────────

func dropPageCache(_ string) error { return exec.Command("purge").Run() }

// ── Process metrics ───────────────────────────────────────────────────────────

type procMetrics struct {
	User       time.Duration
	Sys        time.Duration
	VolCtxSw   int64
	InvolCtxSw int64
}

func getProcessMetrics() procMetrics {
	var ru syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	return procMetrics{
		User: time.Duration(ru.Utime.Nano()),
		Sys:  time.Duration(ru.Stime.Nano()),
	}
}

func (a procMetrics) Sub(b procMetrics) procMetrics {
	return procMetrics{
		User:       a.User - b.User,
		Sys:        a.Sys - b.Sys,
		VolCtxSw:   a.VolCtxSw - b.VolCtxSw,
		InvolCtxSw: a.InvolCtxSw - b.InvolCtxSw,
	}
}

func kernelVersion() string { return "darwin" }

// ── Storage detection (stubbed on macOS) ──────────────────────────────────────

type StorageInfo struct {
	Device, Model, Serial, Firmware, Transport       string
	LinkSpeed, LinkWidth, MaxLinkSpeed, MaxLinkWidth string
}

func (s *StorageInfo) Interface() string           { return "" }
func (s *StorageInfo) PCIeGen() string             { return "" }
func detectStorage(_ string) (*StorageInfo, error) { return nil, nil }
