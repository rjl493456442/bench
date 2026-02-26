// Package main benchmarks sequential disk read performance across different
// page sizes, bypassing the OS page cache to measure true disk I/O latency
// and throughput.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// pageSizes lists the read buffer sizes to benchmark, from 4 KB to 1 MB.
var pageSizes = []int{
	4 * 1024,         // 4 KB  – typical filesystem block
	8 * 1024,         // 8 KB  – PostgreSQL default page size
	16 * 1024,        // 16 KB
	32 * 1024,        // 32 KB
	64 * 1024,        // 64 KB
	128 * 1024,       // 128 KB
	256 * 1024,       // 256 KB
	512 * 1024,       // 512 KB
	1 * 1024 * 1024,  // 1 MB
}

const dataFileName = "bench_pageread.bin"

// result holds raw measurements for one page-size benchmark run.
type result struct {
	pageSize  int
	reads     int64
	bytes     int64
	duration  time.Duration
	latencies []time.Duration // per-read latencies, sorted ascending after run
}

func (r *result) throughputMBps() float64 {
	if r.duration == 0 {
		return 0
	}
	return float64(r.bytes) / r.duration.Seconds() / (1024 * 1024)
}

func (r *result) iops() float64 {
	if r.duration == 0 {
		return 0
	}
	return float64(r.reads) / r.duration.Seconds()
}

func (r *result) avg() time.Duration {
	if len(r.latencies) == 0 {
		return 0
	}
	var total int64
	for _, l := range r.latencies {
		total += int64(l)
	}
	return time.Duration(total / int64(len(r.latencies)))
}

// percentile returns the p-th percentile latency (p in [0,100]).
func (r *result) percentile(p float64) time.Duration {
	n := len(r.latencies)
	if n == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100*float64(n))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return r.latencies[idx]
}

func (r *result) min() time.Duration {
	if len(r.latencies) == 0 {
		return 0
	}
	return r.latencies[0]
}

func (r *result) max() time.Duration {
	if len(r.latencies) == 0 {
		return 0
	}
	return r.latencies[len(r.latencies)-1]
}

// initDataset creates (or verifies) the benchmark data file.
// It fills the file with deterministic pseudo-random bytes so that
// filesystem compression cannot skew results.
func initDataset(path string, size int64) error {
	if info, err := os.Stat(path); err == nil && info.Size() == size {
		log.Printf("Dataset already exists: %s (%.0f MB) – skipping creation", path, float64(size)/(1024*1024))
		return nil
	}
	log.Printf("Initializing dataset: %s (%.0f MB)...", path, float64(size)/(1024*1024))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	defer f.Close()

	const chunkSize = 4 * 1024 * 1024 // 4 MB write chunks
	chunk := make([]byte, chunkSize)
	rng := rand.New(rand.NewPCG(0xdeadbeef, 0xcafebabe))

	var written int64
	for written < size {
		// Fill chunk with pseudo-random 64-bit words.
		for i := 0; i+8 <= len(chunk); i += 8 {
			binary.LittleEndian.PutUint64(chunk[i:], rng.Uint64())
		}
		n := int64(len(chunk))
		if written+n > size {
			n = size - written
		}
		if _, err := f.Write(chunk[:n]); err != nil {
			return fmt.Errorf("write dataset: %w", err)
		}
		written += n
		if written%(64*1024*1024) == 0 {
			log.Printf("  Written %d / %.0f MB", written/(1024*1024), float64(size)/(1024*1024))
		}
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync dataset: %w", err)
	}
	log.Printf("Dataset ready: %s", path)
	return nil
}

// runPass performs one full sequential read of the data file using buffers of
// pageSize bytes, recording per-read latency. The returned result has its
// latency slice sorted ascending.
func runPass(path string, pageSize int) (*result, error) {
	f, err := openDirect(path)
	if err != nil {
		return nil, fmt.Errorf("openDirect: %w", err)
	}
	defer f.Close()

	buf := newBuffer(pageSize)
	defer freeBuffer(buf)

	r := &result{pageSize: pageSize}

	overallStart := time.Now()
	for {
		t0 := time.Now()
		n, readErr := f.Read(buf)
		latency := time.Since(t0)

		if n > 0 {
			r.reads++
			r.bytes += int64(n)
			r.latencies = append(r.latencies, latency)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read (offset %d): %w", r.bytes, readErr)
		}
	}
	r.duration = time.Since(overallStart)

	sort.Slice(r.latencies, func(i, j int) bool {
		return r.latencies[i] < r.latencies[j]
	})
	return r, nil
}

// benchmark runs one or more passes for a given page size, aggregating results.
func benchmark(path string, pageSize, passes int) (*result, error) {
	combined := &result{pageSize: pageSize}

	for pass := 1; pass <= passes; pass++ {
		log.Printf("  [%s] pass %d/%d – clearing page cache...",
			fmtSize(pageSize), pass, passes)

		if err := dropPageCache(path); err != nil {
			log.Printf("    Warning: could not drop page cache: %v", err)
			log.Printf("    Continuing; direct I/O will still bypass the cache.")
		}

		r, err := runPass(path, pageSize)
		if err != nil {
			return nil, err
		}
		log.Printf("  [%s] pass %d/%d – %.2f MB/s, %.0f IOPS, avg %s",
			fmtSize(pageSize), pass, passes,
			r.throughputMBps(), r.iops(), fmtDuration(r.avg()))

		combined.reads += r.reads
		combined.bytes += r.bytes
		combined.duration += r.duration
		combined.latencies = append(combined.latencies, r.latencies...)
	}

	// Re-sort the combined latency slice.
	sort.Slice(combined.latencies, func(i, j int) bool {
		return combined.latencies[i] < combined.latencies[j]
	})
	return combined, nil
}

// ── Formatting helpers ────────────────────────────────────────────────────────

func fmtSize(bytes int) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%d MB", bytes/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%d KB", bytes/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func fmtDuration(d time.Duration) string {
	us := float64(d.Nanoseconds()) / 1e3
	switch {
	case us < 1000:
		return fmt.Sprintf("%.1f µs", us)
	case us < 1e6:
		return fmt.Sprintf("%.2f ms", us/1e3)
	default:
		return fmt.Sprintf("%.2f s", us/1e6)
	}
}

func fmtFloat(f float64) string {
	switch {
	case f >= 1e6:
		return fmt.Sprintf("%.2fM", f/1e6)
	case f >= 1e3:
		return fmt.Sprintf("%.2fK", f/1e3)
	default:
		return fmt.Sprintf("%.2f", f)
	}
}

// ── Markdown export ───────────────────────────────────────────────────────────

func exportMarkdown(outPath, dataFile string, fileSize int64, passes int, results []*result) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	p := func(format string, args ...any) {
		fmt.Fprintf(f, format+"\n", args...)
	}

	p("# Disk Read Performance Benchmark")
	p("")
	p("## Test Environment")
	p("")
	p("| Parameter | Value |")
	p("|-----------|-------|")
	p("| Date | %s |", time.Now().Format("2006-01-02 15:04:05"))
	p("| OS | %s |", runtime.GOOS)
	p("| Architecture | %s |", runtime.GOARCH)
	p("| Data File | `%s` |", dataFile)
	p("| File Size | %.0f MB |", float64(fileSize)/(1024*1024))
	p("| Passes per Page Size | %d |", passes)
	p("| Cache Bypass Method | `%s` |", directIOMethod())
	p("")
	p("## Results")
	p("")
	p("> Latency values are per individual `read(2)` syscall.  ")
	p("> Throughput and IOPS are averaged across all passes.")
	p("")

	// Column widths are fixed; use a separator that matches the header.
	header := "| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |"
	sep := "|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|"
	p(header)
	p(sep)

	for _, r := range results {
		p("| %s | %s | %.2f MB/s | %s | %s | %s | %s | %s | %s | %s | %s |",
			fmtSize(r.pageSize),
			fmtFloat(float64(r.reads)),
			r.throughputMBps(),
			fmtFloat(r.iops()),
			fmtDuration(r.avg()),
			fmtDuration(r.percentile(50)),
			fmtDuration(r.percentile(90)),
			fmtDuration(r.percentile(99)),
			fmtDuration(r.percentile(99.9)),
			fmtDuration(r.min()),
			fmtDuration(r.max()),
		)
	}

	p("")
	p("## Methodology")
	p("")
	p("1. **Dataset**: a `%.0f MB` file filled with pseudo-random bytes (PCG-64 RNG,", float64(fileSize)/(1024*1024))
	p("   deterministic seed) to defeat filesystem-level compression.")
	p("2. **Cache bypass**: `%s` is applied before each pass to ensure reads", directIOMethod())
	p("   come from disk, not the OS page cache.  Additionally, `%s` is", pageCacheDropMethod())
	p("   invoked prior to each pass (best-effort; a warning is printed if it fails).")
	p("3. **Measurement**: wall-clock time is recorded around each `read(2)` call.")
	p("   The file is read sequentially from start to finish.")
	p("4. **Aggregation**: when multiple passes are requested, latency samples are")
	p("   pooled and throughput / IOPS are averaged across all passes.")
	p("")
	p("## Observations")
	p("")

	// Find best throughput entry and the trend description.
	best := results[0]
	for _, r := range results[1:] {
		if r.throughputMBps() > best.throughputMBps() {
			best = r
		}
	}
	p("- Peak throughput **%.2f MB/s** achieved with **%s** page size.",
		best.throughputMBps(), fmtSize(best.pageSize))

	// Compare first vs last
	first, last := results[0], results[len(results)-1]
	ratio := last.throughputMBps() / first.throughputMBps()
	trend := "higher"
	if ratio < 1 {
		ratio = 1 / ratio
		trend = "lower"
	}
	p("- Throughput with %s pages is **%.1fx %s** than with %s pages,",
		fmtSize(last.pageSize), ratio, trend, fmtSize(first.pageSize))
	p("  reflecting reduced syscall overhead with larger buffers.")
	p("- Larger page sizes increase per-read latency but amortise it over more bytes,")
	p("  improving throughput until I/O device bandwidth is saturated.")

	// Summarise key stats as a mini-table.
	p("")
	p("### Quick Reference")
	p("")
	p("| Page Size | Throughput | Avg Latency | P99 Latency |")
	p("|----------:|-----------:|------------:|------------:|")
	for _, r := range results {
		p("| %s | %.2f MB/s | %s | %s |",
			fmtSize(r.pageSize),
			r.throughputMBps(),
			fmtDuration(r.avg()),
			fmtDuration(r.percentile(99)),
		)
	}
	p("")
	p("---")
	p("")
	p("*Generated by [bench/disk/pageread](https://github.com/rjl493456442/bench)*")

	return nil
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	var (
		dir      = flag.String("dir", ".", "directory where the benchmark data file is stored")
		sizeMB   = flag.Int("size", 512, "dataset size in MB (rounded down to nearest MB)")
		output   = flag.String("output", "results.md", "path for the markdown results file")
		skipInit = flag.Bool("skip-init", false, "skip dataset creation if a correct-size file already exists")
		passes   = flag.Int("passes", 1, "number of read passes per page size (results are aggregated)")
	)
	flag.Parse()

	// Align file size to 1 MB so every page size divides it evenly.
	fileSize := int64(*sizeMB) * 1024 * 1024

	dataPath := filepath.Join(*dir, dataFileName)

	if !*skipInit {
		if err := initDataset(dataPath, fileSize); err != nil {
			log.Fatalf("dataset init: %v", err)
		}
	} else {
		if _, err := os.Stat(dataPath); err != nil {
			log.Fatalf("dataset not found (--skip-init was set): %v", err)
		}
	}

	log.Printf("Starting benchmark")
	log.Printf("  File   : %s (%.0f MB)", dataPath, float64(fileSize)/(1024*1024))
	log.Printf("  Sizes  : %s", strings.Join(func() []string {
		ss := make([]string, len(pageSizes))
		for i, ps := range pageSizes {
			ss[i] = fmtSize(ps)
		}
		return ss
	}(), ", "))
	log.Printf("  Passes : %d", *passes)
	log.Printf("  Method : %s + %s (cache drop)", directIOMethod(), pageCacheDropMethod())

	results := make([]*result, 0, len(pageSizes))
	for _, ps := range pageSizes {
		log.Printf("Benchmarking %s pages...", fmtSize(ps))
		r, err := benchmark(dataPath, ps, *passes)
		if err != nil {
			log.Fatalf("benchmark [%s]: %v", fmtSize(ps), err)
		}
		results = append(results, r)
		log.Printf("  => %.2f MB/s, %.0f IOPS, avg %s, p99 %s",
			r.throughputMBps(), r.iops(),
			fmtDuration(r.avg()), fmtDuration(r.percentile(99)))
	}

	if err := exportMarkdown(*output, dataPath, fileSize, *passes, results); err != nil {
		log.Fatalf("export markdown: %v", err)
	}
	log.Printf("Results written to: %s", *output)
}
