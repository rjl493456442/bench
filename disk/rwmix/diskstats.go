package main

import "time"

// diskStat is a snapshot of a block device's cumulative counters, mirroring the
// fields of /proc/diskstats that iostat consumes.  Sector counts are always in
// 512-byte units (kernel convention), independent of the device block size.
type diskStat struct {
	rdIOs     uint64 // reads completed
	rdSectors uint64 // sectors read (512 B each)
	rdTicks   uint64 // ms spent servicing reads (includes queueing)
	wrIOs     uint64 // writes completed
	wrSectors uint64 // sectors written (512 B each)
	wrTicks   uint64 // ms spent servicing writes (includes queueing)
	ioTicks   uint64 // ms during which the device had I/O in flight
	weighted  uint64 // weighted ms (Σ in-flight × time) → average queue depth
}

// DiskDelta holds device-level rates derived from two snapshots, matching the
// columns iostat reports (r/s, rMB/s, r_await, …, aqu-sz, %util).
type DiskDelta struct {
	ReadIOPS  float64
	WriteIOPS float64
	ReadMBps  float64
	WriteMBps float64

	// Average time a request waited, from entering the queue to completion (ms).
	AvgReadWaitMs  float64
	AvgWriteWaitMs float64

	// AvgQueueDepth (iostat aqu-sz): mean number of requests in flight.
	AvgQueueDepth float64

	// UtilPct: fraction of wall time the device had at least one request in
	// flight.  Unreliable on multi-queue NVMe (can saturate below true 100%).
	UtilPct float64
}

// computeDiskDelta derives per-second rates and average wait times from two
// counter snapshots taken `elapsed` apart.
func computeDiskDelta(before, after diskStat, elapsed time.Duration) DiskDelta {
	secs := elapsed.Seconds()
	ms := float64(elapsed) / float64(time.Millisecond)

	drIOs := float64(after.rdIOs - before.rdIOs)
	dwIOs := float64(after.wrIOs - before.wrIOs)

	d := DiskDelta{}
	if secs > 0 {
		d.ReadIOPS = drIOs / secs
		d.WriteIOPS = dwIOs / secs
		d.ReadMBps = float64(after.rdSectors-before.rdSectors) * 512 / secs / (1024 * 1024)
		d.WriteMBps = float64(after.wrSectors-before.wrSectors) * 512 / secs / (1024 * 1024)
	}
	if drIOs > 0 {
		d.AvgReadWaitMs = float64(after.rdTicks-before.rdTicks) / drIOs
	}
	if dwIOs > 0 {
		d.AvgWriteWaitMs = float64(after.wrTicks-before.wrTicks) / dwIOs
	}
	if ms > 0 {
		d.AvgQueueDepth = float64(after.weighted-before.weighted) / ms
		d.UtilPct = float64(after.ioTicks-before.ioTicks) / ms * 100
	}
	return d
}
