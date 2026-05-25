//go:build linux

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// readDiskStat returns the current counters for the named block device (e.g.
// "nvme0n1") by parsing /proc/diskstats.  The bool is false when the device is
// not found or the line is malformed.
func readDiskStat(device string) (diskStat, bool) {
	if device == "" {
		return diskStat{}, false
	}
	f, err := os.Open("/proc/diskstats")
	if err != nil {
		return diskStat{}, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// Need at least the classic 14 fields (3 header + 11 stats).
		if len(fields) < 14 || fields[2] != device {
			continue
		}
		u := func(i int) uint64 {
			v, _ := strconv.ParseUint(fields[i], 10, 64)
			return v
		}
		return diskStat{
			rdIOs:     u(3),
			rdSectors: u(5),
			rdTicks:   u(6),
			wrIOs:     u(7),
			wrSectors: u(9),
			wrTicks:   u(10),
			ioTicks:   u(12),
			weighted:  u(13),
		}, true
	}
	return diskStat{}, false
}
