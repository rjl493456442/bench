//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const fNOCACHE = 48

func openDirect(name string) (*os.File, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	if _, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), uintptr(fNOCACHE), 1); errno != 0 {
		f.Close()
		return nil, fmt.Errorf("fcntl(F_NOCACHE): %w", errno)
	}
	return f, nil
}

func openBuffered(name string) (*os.File, error) {
	return os.Open(name)
}

func newBuffer(size int) []byte {
	buf, err := syscall.Mmap(-1, 0, size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANON|syscall.MAP_PRIVATE)
	if err != nil {
		panic(fmt.Sprintf("mmap(%d): %v", size, err))
	}
	return buf
}

func freeBuffer(buf []byte) {
	_ = syscall.Munmap(buf)
}

func dropPageCache(_ string) error {
	return exec.Command("purge").Run()
}

func directIOMethod() string { return "fcntl(F_NOCACHE)" }

func pinCPU(_ int) error { return nil } // not supported on macOS

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

// StorageInfo stubs for macOS.
type StorageInfo struct {
	Device, Model, Serial, Firmware, Transport string
	LinkSpeed, LinkWidth, MaxLinkSpeed, MaxLinkWidth string
}

func (s *StorageInfo) Interface() string { return "" }
func (s *StorageInfo) PCIeGen() string   { return "" }
func detectStorage(_ string) (*StorageInfo, error) { return nil, nil }
