//go:build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// StorageInfo holds detected storage device information.
type StorageInfo struct {
	Device    string // e.g. "nvme1n1"
	Model     string // e.g. "Samsung SSD 990 PRO 4TB"
	Serial    string
	Firmware  string
	Transport string // "NVMe" or "sata" etc.
	// PCIe / NVMe link fields
	LinkSpeed    string // e.g. "16.0 GT/s PCIe"
	LinkWidth    string // e.g. "4"
	MaxLinkSpeed string
	MaxLinkWidth string
}

// PCIeGen converts a GT/s speed string to a PCIe generation label.
func (s *StorageInfo) PCIeGen() string {
	table := map[string]string{
		"2.5":  "1.0",
		"5.0":  "2.0",
		"8.0":  "3.0",
		"16.0": "4.0",
		"32.0": "5.0",
		"64.0": "6.0",
	}
	// Extract the leading number before " GT/s"
	speed := strings.TrimSpace(s.LinkSpeed)
	speed = strings.TrimSuffix(speed, " GT/s PCIe")
	speed = strings.TrimSuffix(speed, " GT/s")
	speed = strings.TrimSpace(speed)
	if gen, ok := table[speed]; ok {
		return gen
	}
	return ""
}

// Interface returns a human-readable interface description, e.g. "NVMe PCIe 4.0 x4".
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

// detectStorage resolves which block device backs filePath and reads its
// model, serial, firmware, and PCIe link information from sysfs.
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
		// sysBlock is a symlink into something like:
		//   /sys/devices/.../nvmeN/nvme/nvmeNnM
		// Resolving it and going up two levels reaches the nvmeN controller dir.
		resolved, err := filepath.EvalSymlinks(sysBlock)
		if err == nil {
			// resolved = .../pci-addr/nvme/nvmeN
			// Go up twice: nvmeN -> nvme -> pci-addr (PCI device dir)
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

// resolveBlockDisk finds which block device (disk, not partition) a file lives on
// by parsing /proc/self/mountinfo, then strips any partition suffix.
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

	// mountinfo fields (space-separated):
	//  0:mountID 1:parentID 2:major:minor 3:root 4:mountpoint 5:mountopts
	//  -- optional fields -- separator:"-" 7:fstype 8:source 9:superopts
	var bestMount, bestSource string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		mountpoint := fields[4]
		// Find the separator "-" in optional fields
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
		// Strip subvolume: e.g. "/dev/sda1[/subvol]"
		if idx := strings.Index(source, "["); idx >= 0 {
			source = source[:idx]
		}

		// Keep the longest matching mountpoint prefix.
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

	devName := filepath.Base(bestSource)
	return stripPartition(devName), nil
}

var (
	reNVMePartition  = regexp.MustCompile(`^(nvme\d+n\d+)p\d+$`)
	reMMCPartition   = regexp.MustCompile(`^(mmcblk\d+)p\d+$`)
	reSATAPartition  = regexp.MustCompile(`^([a-z]+)\d+$`)
)

// stripPartition removes the partition suffix from a block device name.
//   nvme0n1p2 → nvme0n1
//   mmcblk0p1 → mmcblk0
//   sda3      → sda
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

// readSysFile reads a sysfs attribute file and returns its trimmed contents,
// or an empty string if the file cannot be read.
func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
