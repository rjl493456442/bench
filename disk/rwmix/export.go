package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// exportMarkdown writes the full results document.
func exportMarkdown(outPath string, cfg Config, results []*MixResult, storage *StorageInfo) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	p := func(format string, args ...any) { fmt.Fprintf(f, format+"\n", args...) }

	// ── Environment ──────────────────────────────────────────────────────
	p("# Concurrent Read/Write Performance Benchmark")
	p("")
	p("Reads and writes run in independent goroutine pools (the **R:W mix**)")
	p("competing for one device, so each side's limit can be measured while the")
	p("other side loads the disk.  Writes are made durable with `%s`.", syncMethod())
	p("")
	p("## Test Environment")
	p("")
	p("| Parameter | Value |")
	p("|-----------|-------|")
	p("| Date | %s |", time.Now().Format("2006-01-02 15:04:05"))
	p("| OS | %s |", runtime.GOOS)
	p("| Architecture | %s |", runtime.GOARCH)
	p("| Kernel | %s |", kernelVersion())
	p("| Data File | `%s` |", cfg.FilePath)
	p("| File Size | %.1f GB |", float64(cfg.FileSize)/(1<<30))
	ioMode := "buffered"
	if cfg.DirectIO {
		ioMode = directIOMethod()
	}
	p("| I/O Mode | %s |", ioMode)
	p("| Duration / mix | %s |", cfg.Duration)
	p("| Read | %s, %s blocks |", cfg.ReadMode, fmtSize(cfg.ReadBlock))
	p("| Write | %s, %s blocks |", cfg.WriteMode, fmtSize(cfg.WriteBlock))
	p("| Sync | `%s` every %d write(s) |", syncMethod(), cfg.FsyncEvery)
	if storage != nil {
		p("| Storage Device | %s (`%s`) |", storage.Model, storage.Device)
		p("| Interface | %s |", storage.Interface())
		if storage.Firmware != "" {
			p("| Firmware | %s |", storage.Firmware)
		}
	}
	p("")

	// ── Application-measured results ─────────────────────────────────────
	p("## Application-Measured Throughput & Latency")
	p("")
	p("Measured inside the benchmark process (per-op wall time).  Write latency")
	p("includes the `%s` cost on the op that triggers it.", syncMethod())
	p("")
	p("| Mix (R:W) | Rd MB/s | Rd IOPS | Rd Avg | Rd P99 | Wr MB/s | Wr IOPS | Wr Avg | Wr P99 | CPU |")
	p("|-----------|--------:|--------:|-------:|-------:|--------:|--------:|-------:|-------:|----:|")
	for _, r := range results {
		p("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |",
			r.Mix,
			sideMBps(&r.Read, r.Elapsed, r.Mix.Readers),
			sideIOPS(&r.Read, r.Elapsed, r.Mix.Readers),
			sideLat(r.Mix.Readers, r.Read.Avg()),
			sideLat(r.Mix.Readers, r.Read.Pct(99)),
			sideMBps(&r.Write, r.Elapsed, r.Mix.Writers),
			sideIOPS(&r.Write, r.Elapsed, r.Mix.Writers),
			sideLat(r.Mix.Writers, r.Write.Avg()),
			sideLat(r.Mix.Writers, r.Write.Pct(99)),
			fmtDuration(r.CPUUser+r.CPUSys),
		)
	}
	p("")

	// ── Device-level (iostat-style) results ──────────────────────────────
	hasDisk := false
	for _, r := range results {
		if r.Disk != nil {
			hasDisk = true
			break
		}
	}
	if hasDisk {
		p("## Device-Level Counters (from `/proc/diskstats`)")
		p("")
		p("Measured by the kernel for the whole device, the way `iostat` reports.")
		p("`r_await`/`w_await` are the mean time a request spends from queue entry")
		p("to completion; `aqu-sz` is the mean number of requests in flight.")
		p("")
		p("| Mix (R:W) | r/s | rMB/s | r_await | w/s | wMB/s | w_await | aqu-sz | %%util |")
		p("|-----------|----:|------:|--------:|----:|------:|--------:|-------:|------:|")
		for _, r := range results {
			d := r.Disk
			if d == nil {
				p("| %s | – | – | – | – | – | – | – | – |", r.Mix)
				continue
			}
			p("| %s | %s | %.1f | %s | %s | %.1f | %s | %.2f | %.1f |",
				r.Mix,
				fmtFloat(d.ReadIOPS), d.ReadMBps, fmtMs(d.AvgReadWaitMs),
				fmtFloat(d.WriteIOPS), d.WriteMBps, fmtMs(d.AvgWriteWaitMs),
				d.AvgQueueDepth, d.UtilPct,
			)
		}
		p("")
	}

	// ── Observations ─────────────────────────────────────────────────────
	p("## Observations")
	p("")
	writeObservations(p, results)
	p("")

	// ── Methodology ──────────────────────────────────────────────────────
	p("## Methodology")
	p("")
	p("1. **Dataset**: a single %.1f GB file of pseudo-random bytes (PCG-64,",
		float64(cfg.FileSize)/(1<<30))
	p("   deterministic seed) defeats compression/dedup and is large enough that")
	p("   an NVMe SSD cannot serve it from its SLC write cache, so steady-state")
	p("   numbers are not inflated.")
	p("2. **Mix**: each `R:W` entry runs R reader goroutines and W writer")
	p("   goroutines concurrently against the same file for the configured")
	p("   duration.  Readers and writers share one fd each and issue")
	p("   `pread`/`pwrite` at block-aligned offsets.")
	p("3. **Durability**: writers call `%s` every %d write(s); the cost is charged",
		syncMethod(), cfg.FsyncEvery)
	p("   to the triggering op so per-op write latency reflects a durable write,")
	p("   and wall-time throughput always includes the sync cost.")
	if cfg.DirectIO {
		p("4. **Direct I/O** (`%s`) bypasses the page cache so reads and writes hit",
			directIOMethod())
		p("   the device, not RAM.")
	} else {
		p("4. **Buffered I/O**: reads may be served from the page cache; only the")
		p("   `%s` calls force writes to the device.", syncMethod())
	}
	p("5. **Latency** percentiles come from a per-worker uniform reservoir (up to")
	p("   %d samples each); the average is exact over all ops.", maxSamplesPerWorker)
	p("6. **Device counters** are sampled from `/proc/diskstats` immediately")
	p("   before and after each timed window, then divided by the elapsed time —")
	p("   the same arithmetic `iostat` performs.")
	p("")
	p("---")
	p("")
	p("*Generated by [bench/disk/rwmix](https://github.com/rjl493456442/bench)*")

	return nil
}

// ── Cell helpers ──────────────────────────────────────────────────────────────

// A side with zero workers contributes nothing; show "–" instead of 0.
func sideMBps(s *SideResult, d time.Duration, workers int) string {
	if workers == 0 {
		return "–"
	}
	return fmt.Sprintf("%.1f", s.MBps(d))
}

func sideIOPS(s *SideResult, d time.Duration, workers int) string {
	if workers == 0 {
		return "–"
	}
	return fmtFloat(s.IOPS(d))
}

func sideLat(workers int, lat time.Duration) string {
	if workers == 0 {
		return "–"
	}
	return fmtDuration(lat)
}

func fmtMs(ms float64) string {
	if ms <= 0 {
		return "–"
	}
	if ms < 1 {
		return fmt.Sprintf("%.0f µs", ms*1000)
	}
	return fmt.Sprintf("%.2f ms", ms)
}

// writeObservations highlights the peak read and peak write points and how the
// other side's presence affects them.
func writeObservations(p func(string, ...any), results []*MixResult) {
	var bestRead, bestWrite *MixResult
	for _, r := range results {
		if r.Mix.Readers > 0 && (bestRead == nil || r.Read.MBps(r.Elapsed) > bestRead.Read.MBps(bestRead.Elapsed)) {
			bestRead = r
		}
		if r.Mix.Writers > 0 && (bestWrite == nil || r.Write.MBps(r.Elapsed) > bestWrite.Write.MBps(bestWrite.Elapsed)) {
			bestWrite = r
		}
	}
	if bestRead != nil {
		p("- **Peak read**: %.1f MB/s (%s IOPS) at mix %s.",
			bestRead.Read.MBps(bestRead.Elapsed), fmtFloat(bestRead.Read.IOPS(bestRead.Elapsed)), bestRead.Mix)
	}
	if bestWrite != nil {
		p("- **Peak write**: %.1f MB/s (%s IOPS) at mix %s.",
			bestWrite.Write.MBps(bestWrite.Elapsed), fmtFloat(bestWrite.Write.IOPS(bestWrite.Elapsed)), bestWrite.Mix)
	}

	// Quantify read interference: compare a read-only mix against the mix with
	// the same reader count but writers added, if both exist.
	for _, base := range results {
		if base.Mix.Writers != 0 || base.Mix.Readers == 0 {
			continue
		}
		for _, mixed := range results {
			if mixed.Mix.Readers == base.Mix.Readers && mixed.Mix.Writers > 0 {
				bMBps := base.Read.MBps(base.Elapsed)
				mMBps := mixed.Read.MBps(mixed.Elapsed)
				if bMBps > 0 {
					p("- Adding %d writer(s) to %d reader(s) drops read throughput "+
						"from %.1f to %.1f MB/s (**%.0f%%** of isolated), read avg latency %s → %s.",
						mixed.Mix.Writers, base.Mix.Readers, bMBps, mMBps, mMBps/bMBps*100,
						fmtDuration(base.Read.Avg()), fmtDuration(mixed.Read.Avg()))
				}
			}
		}
	}
}
