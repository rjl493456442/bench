//go:build darwin

package main

// readDiskStat is unavailable on macOS: there is no /proc/diskstats and the
// `iostat` output is not per-device-cumulative in a usable way.  Device-level
// metrics are simply omitted; the app-level measurements still apply.
func readDiskStat(_ string) (diskStat, bool) {
	return diskStat{}, false
}
