//go:build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// ── Direct I/O ────────────────────────────────────────────────────────────────

func openDirect(name string) (*os.File, error) {
	fd, err := syscall.Open(name, syscall.O_RDONLY|syscall.O_DIRECT, 0)
	if err != nil {
		return nil, fmt.Errorf("open O_DIRECT %q: %w", name, err)
	}
	return os.NewFile(uintptr(fd), name), nil
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

func dropPageCache(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := unix.Fadvise(int(f.Fd()), 0, 0, unix.FADV_DONTNEED); err == nil {
		return nil
	}
	syscall.Sync()
	df, err := os.OpenFile("/proc/sys/vm/drop_caches", os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("fadvise unavailable and drop_caches requires root: %w", err)
	}
	defer df.Close()
	_, err = df.WriteString("3\n")
	return err
}

func directIOMethod() string { return "O_DIRECT" }

// ── CPU pinning ───────────────────────────────────────────────────────────────

func pinCPU(cpu int) error {
	runtime.LockOSThread()
	var set unix.CPUSet
	set.Set(cpu)
	return unix.SchedSetaffinity(0, &set)
}

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
		User:       time.Duration(ru.Utime.Nano()),
		Sys:        time.Duration(ru.Stime.Nano()),
		VolCtxSw:   int64(ru.Nvcsw),
		InvolCtxSw: int64(ru.Nivcsw),
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

// ── Kernel version ────────────────────────────────────────────────────────────

func kernelVersion() string {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return "unknown"
	}
	return unix.ByteSliceToString(uts.Release[:])
}

// ── Storage detection ─────────────────────────────────────────────────────────

type StorageInfo struct {
	Device, Model, Serial, Firmware, Transport         string
	LinkSpeed, LinkWidth, MaxLinkSpeed, MaxLinkWidth string
}

func (s *StorageInfo) PCIeGen() string {
	table := map[string]string{
		"2.5": "1.0", "5.0": "2.0", "8.0": "3.0",
		"16.0": "4.0", "32.0": "5.0", "64.0": "6.0",
	}
	speed := strings.TrimSpace(s.LinkSpeed)
	speed = strings.TrimSuffix(speed, " GT/s PCIe")
	speed = strings.TrimSuffix(speed, " GT/s")
	speed = strings.TrimSpace(speed)
	if gen, ok := table[speed]; ok {
		return gen
	}
	return ""
}

func (s *StorageInfo) Interface() string {
	if s.Transport == "NVMe" {
		gen := s.PCIeGen()
		width := strings.TrimSpace(s.LinkWidth)
		if gen != "" && width != "" {
			return fmt.Sprintf("NVMe PCIe %s x%s", gen, width)
		}
		if gen != "" {
			return "NVMe PCIe " + gen
		}
		return "NVMe"
	}
	if s.Transport != "" {
		return strings.ToUpper(s.Transport)
	}
	return "Unknown"
}

func detectStorage(filePath string) (*StorageInfo, error) {
	disk, err := resolveBlockDisk(filePath)
	if err != nil {
		return nil, err
	}
	info := &StorageInfo{Device: disk}
	sysBlock := "/sys/block/" + disk + "/device"
	info.Model = readSysFile(sysBlock + "/model")
	info.Serial = readSysFile(sysBlock + "/serial")
	info.Firmware = readSysFile(sysBlock + "/firmware_rev")
	if strings.HasPrefix(disk, "nvme") {
		info.Transport = "NVMe"
		resolved, err := filepath.EvalSymlinks(sysBlock)
		if err == nil {
			pciDir := filepath.Dir(filepath.Dir(resolved))
			info.LinkSpeed = readSysFile(pciDir + "/current_link_speed")
			info.LinkWidth = readSysFile(pciDir + "/current_link_width")
			info.MaxLinkSpeed = readSysFile(pciDir + "/max_link_speed")
			info.MaxLinkWidth = readSysFile(pciDir + "/max_link_width")
		}
	} else {
		info.Transport = readSysFile(filepath.Dir(sysBlock) + "/device/../transport")
		if info.Transport == "" {
			info.Transport = readSysFile(sysBlock + "/../transport")
		}
	}
	return info, nil
}

func resolveBlockDisk(filePath string) (string, error) {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", fmt.Errorf("open mountinfo: %w", err)
	}
	defer f.Close()
	var bestMount, bestSource string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		mountpoint := fields[4]
		sepIdx := -1
		for i, v := range fields {
			if v == "-" {
				sepIdx = i
				break
			}
		}
		if sepIdx < 0 || sepIdx+2 >= len(fields) {
			continue
		}
		source := fields[sepIdx+2]
		if idx := strings.Index(source, "["); idx >= 0 {
			source = source[:idx]
		}
		if strings.HasPrefix(abs, mountpoint+"/") || abs == mountpoint {
			if len(mountpoint) > len(bestMount) {
				bestMount = mountpoint
				bestSource = source
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan mountinfo: %w", err)
	}
	if bestSource == "" {
		return "", fmt.Errorf("could not determine device for %s", filePath)
	}
	return stripPartition(filepath.Base(bestSource)), nil
}

var (
	reNVMePartition = regexp.MustCompile(`^(nvme\d+n\d+)p\d+$`)
	reMMCPartition  = regexp.MustCompile(`^(mmcblk\d+)p\d+$`)
	reSATAPartition = regexp.MustCompile(`^([a-z]+)\d+$`)
)

func stripPartition(dev string) string {
	if m := reNVMePartition.FindStringSubmatch(dev); m != nil {
		return m[1]
	}
	if m := reMMCPartition.FindStringSubmatch(dev); m != nil {
		return m[1]
	}
	if m := reSATAPartition.FindStringSubmatch(dev); m != nil {
		return m[1]
	}
	return dev
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
