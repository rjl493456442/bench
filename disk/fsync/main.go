// Package main benchmarks fsync performance across different dirty data sizes
// and concurrency levels, measuring the time it takes to flush written data
// from the OS buffer cache to the storage device.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// dirtySizes lists the amounts of data written before each fsync call.
var dirtySizes = []int{
	4 * 1024,          // 4 KB  – single filesystem block
	8 * 1024,          // 8 KB
	16 * 1024,         // 16 KB
	32 * 1024,         // 32 KB
	64 * 1024,         // 64 KB
	128 * 1024,        // 128 KB
	256 * 1024,        // 256 KB
	512 * 1024,        // 512 KB
	1 * 1024 * 1024,   // 1 MB
	2 * 1024 * 1024,   // 2 MB
	4 * 1024 * 1024,   // 4 MB
	8 * 1024 * 1024,   // 8 MB
	16 * 1024 * 1024,  // 16 MB
	32 * 1024 * 1024,  // 32 MB
	64 * 1024 * 1024,  // 64 MB
	128 * 1024 * 1024, // 128 MB
}

// ── Result type ───────────────────────────────────────────────────────────────

// result holds raw measurements for one (dirtySize, queueDepth) benchmark run.
type result struct {
	dirtySize  int
	queueDepth int
	fsyncs     int64
	bytes      int64
	duration   time.Duration   // wall-clock time including writes and fsyncs
	latencies  []time.Duration // per-fsync latencies, sorted ascending after run
}

func (r *result) throughputMBps() float64 {
	if r.duration == 0 {
		return 0
	}
	return float64(r.bytes) / r.duration.Seconds() / (1024 * 1024)
}

func (r *result) opsPerSec() float64 {
	if r.duration == 0 {
		return 0
	}
	return float64(r.fsyncs) / r.duration.Seconds()
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

// percentile returns the p-th percentile latency (p in [0, 100]).
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

// ── Data buffer ───────────────────────────────────────────────────────────────

// initBuffer creates a buffer filled with pseudo-random bytes to defeat
// filesystem compression.
func initBuffer(size int) []byte {
	buf := make([]byte, size)
	rng := rand.New(rand.NewPCG(0xdeadbeef, 0xcafebabe))
	for i := 0; i+8 <= len(buf); i += 8 {
		binary.LittleEndian.PutUint64(buf[i:], rng.Uint64())
	}
	return buf
}

// ── Benchmark core ────────────────────────────────────────────────────────────

// runBench performs a benchmark for a given dirty size and queue depth.
// For QD=1, a single goroutine writes dirtySize bytes then fsyncs, repeated
// `iters` times.  For QD>1, multiple goroutines do the same concurrently on
// separate files.
func runBench(dir string, dirtySize int, iters int, queueDepth int, dataBuf []byte) (*result, error) {
	if queueDepth > 1 {
		return runBenchConcurrent(dir, dirtySize, iters, queueDepth, dataBuf)
	}

	// Create benchmark file pre-sized to dirtySize.
	path := filepath.Join(dir, "bench_fsync_0.tmp")
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		f.Close()
		os.Remove(path)
	}()

	// Pre-extend file to dirtySize so overwrites don't change metadata (size).
	if _, err := f.Write(dataBuf[:dirtySize]); err != nil {
		return nil, fmt.Errorf("pre-extend: %w", err)
	}
	if err := syncFile(f); err != nil {
		return nil, fmt.Errorf("initial sync: %w", err)
	}

	r := &result{
		dirtySize:  dirtySize,
		queueDepth: 1,
		latencies:  make([]time.Duration, 0, iters),
	}

	start := time.Now()

	for i := 0; i < iters; i++ {
		// Seek to beginning and overwrite with dirty data.
		if _, err := f.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("seek: %w", err)
		}
		if _, err := f.Write(dataBuf[:dirtySize]); err != nil {
			return nil, fmt.Errorf("write: %w", err)
		}

		// Measure fsync latency only.
		t0 := time.Now()
		if err := syncFile(f); err != nil {
			return nil, fmt.Errorf("fsync: %w", err)
		}
		latency := time.Since(t0)

		r.fsyncs++
		r.bytes += int64(dirtySize)
		r.latencies = append(r.latencies, latency)
	}

	r.duration = time.Since(start)

	sort.Slice(r.latencies, func(i, j int) bool {
		return r.latencies[i] < r.latencies[j]
	})
	return r, nil
}

// runBenchConcurrent performs a benchmark with multiple goroutines, each
// operating on its own file.  This measures how fsync latency scales under
// device contention.
func runBenchConcurrent(dir string, dirtySize int, iters int, queueDepth int, dataBuf []byte) (*result, error) {
	// Create per-worker files.
	files := make([]*os.File, queueDepth)
	paths := make([]string, queueDepth)
	for w := 0; w < queueDepth; w++ {
		paths[w] = filepath.Join(dir, fmt.Sprintf("bench_fsync_%d.tmp", w))
		f, err := os.Create(paths[w])
		if err != nil {
			for j := 0; j < w; j++ {
				files[j].Close()
				os.Remove(paths[j])
			}
			return nil, fmt.Errorf("create %s: %w", paths[w], err)
		}
		files[w] = f

		// Pre-extend.
		if _, err := f.Write(dataBuf[:dirtySize]); err != nil {
			f.Close()
			return nil, fmt.Errorf("pre-extend worker %d: %w", w, err)
		}
		if err := syncFile(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("initial sync worker %d: %w", w, err)
		}
	}
	defer func() {
		for w := 0; w < queueDepth; w++ {
			files[w].Close()
			os.Remove(paths[w])
		}
	}()

	// Divide iterations among workers so total fsync count stays ~constant.
	itersPerWorker := iters / queueDepth
	if itersPerWorker < 1 {
		itersPerWorker = 1
	}
	totalIters := itersPerWorker * queueDepth

	// Pre-allocate per-iteration latency slots; each worker owns its indices.
	latencies := make([]time.Duration, totalIters)

	var wg sync.WaitGroup
	var syncErr error
	var errOnce sync.Once

	// Barrier so all workers start simultaneously.
	var barrier sync.WaitGroup
	barrier.Add(1)

	for w := 0; w < queueDepth; w++ {
		wg.Add(1)
		go func(workerID int, f *os.File) {
			defer wg.Done()
			barrier.Wait()

			base := workerID * itersPerWorker
			for i := 0; i < itersPerWorker; i++ {
				if _, err := f.Seek(0, 0); err != nil {
					errOnce.Do(func() { syncErr = fmt.Errorf("seek worker %d: %w", workerID, err) })
					return
				}
				if _, err := f.Write(dataBuf[:dirtySize]); err != nil {
					errOnce.Do(func() { syncErr = fmt.Errorf("write worker %d: %w", workerID, err) })
					return
				}

				t0 := time.Now()
				if err := syncFile(f); err != nil {
					errOnce.Do(func() { syncErr = fmt.Errorf("fsync worker %d: %w", workerID, err) })
					return
				}
				latencies[base+i] = time.Since(t0)
			}
		}(w, files[w])
	}

	start := time.Now()
	barrier.Done() // release all workers
	wg.Wait()
	duration := time.Since(start)

	if syncErr != nil {
		return nil, syncErr
	}

	// Aggregate results.
	r := &result{
		dirtySize:  dirtySize,
		queueDepth: queueDepth,
		duration:   duration,
		latencies:  make([]time.Duration, 0, totalIters),
	}
	for i := 0; i < totalIters; i++ {
		r.fsyncs++
		r.bytes += int64(dirtySize)
		r.latencies = append(r.latencies, latencies[i])
	}

	sort.Slice(r.latencies, func(i, j int) bool {
		return r.latencies[i] < r.latencies[j]
	})
	return r, nil
}

// benchmark runs one or more passes for a (dirtySize, queueDepth) tuple
// and aggregates all passes into a single result.
func benchmark(dir string, dirtySize int, iters int, passes int, queueDepth int, dataBuf []byte) (*result, error) {
	combined := &result{dirtySize: dirtySize, queueDepth: queueDepth}

	for pass := 1; pass <= passes; pass++ {
		log.Printf("  [%s/QD%d] pass %d/%d...",
			fmtSize(dirtySize), queueDepth, pass, passes)

		r, err := runBench(dir, dirtySize, iters, queueDepth, dataBuf)
		if err != nil {
			return nil, err
		}
		log.Printf("  [%s/QD%d] pass %d/%d – %.2f MB/s, %.0f fsync/s, avg %s",
			fmtSize(dirtySize), queueDepth, pass, passes,
			r.throughputMBps(), r.opsPerSec(), fmtDuration(r.avg()))

		combined.fsyncs += r.fsyncs
		combined.bytes += r.bytes
		combined.duration += r.duration
		combined.latencies = append(combined.latencies, r.latencies...)
	}

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

func writeResultTable(p func(string, ...any), results []*result) {
	p("| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |")
	p("|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|")
	for _, r := range results {
		p("| %s | %s | %.2f MB/s | %s | %s | %s | %s | %s | %s | %s | %s |",
			fmtSize(r.dirtySize),
			fmtFloat(float64(r.fsyncs)),
			r.throughputMBps(),
			fmtFloat(r.opsPerSec()),
			fmtDuration(r.avg()),
			fmtDuration(r.percentile(50)),
			fmtDuration(r.percentile(90)),
			fmtDuration(r.percentile(99)),
			fmtDuration(r.percentile(99.9)),
			fmtDuration(r.min()),
			fmtDuration(r.max()),
		)
	}
}

func writeObservations(p func(string, ...any), results []*result) {
	if len(results) == 0 {
		return
	}

	// Find best throughput.
	best := results[0]
	for _, r := range results[1:] {
		if r.throughputMBps() > best.throughputMBps() {
			best = r
		}
	}
	p("- Peak effective throughput **%.2f MB/s** achieved with **%s** dirty size.",
		best.throughputMBps(), fmtSize(best.dirtySize))

	// Find lowest latency.
	fastest := results[0]
	for _, r := range results[1:] {
		if r.avg() < fastest.avg() {
			fastest = r
		}
	}
	p("- Lowest average fsync latency **%s** with **%s** dirty size.",
		fmtDuration(fastest.avg()), fmtSize(fastest.dirtySize))

	first, last := results[0], results[len(results)-1]
	ratio := last.avg().Seconds() / first.avg().Seconds()
	p("- Fsync latency with %s dirty data is **%.1fx** the latency with %s.",
		fmtSize(last.dirtySize), ratio, fmtSize(first.dirtySize))
}

func exportMarkdown(outPath, dir string, iters, passes int, allResults map[int][]*result, queueDepths []int, storage *StorageInfo) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	p := func(format string, args ...any) {
		fmt.Fprintf(f, format+"\n", args...)
	}

	multiQD := len(queueDepths) > 1

	qdStrs := make([]string, len(queueDepths))
	for i, qd := range queueDepths {
		qdStrs[i] = strconv.Itoa(qd)
	}

	p("# Disk Fsync Performance Benchmark")
	p("")
	p("## Test Environment")
	p("")
	p("| Parameter | Value |")
	p("|-----------|-------|")
	p("| Date | %s |", time.Now().Format("2006-01-02 15:04:05"))
	p("| OS | %s |", runtime.GOOS)
	p("| Architecture | %s |", runtime.GOARCH)
	p("| Directory | `%s` |", dir)
	p("| Iterations per Size | %d |", iters)
	p("| Queue Depths | %s |", strings.Join(qdStrs, ", "))
	p("| Passes | %d |", passes)
	p("| Sync Method | `%s` |", syncMethod())
	if storage != nil {
		p("| Storage Device | %s (`%s`) |", storage.Model, storage.Device)
		p("| Interface | %s |", storage.Interface())
		if storage.Firmware != "" {
			p("| Firmware | %s |", storage.Firmware)
		}
	}
	p("")
	p("> Latency values are per individual `%s` call.  ", syncMethod())
	p("> Throughput = dirty_size x fsync_count / wall_time (includes write overhead).")

	for _, qd := range queueDepths {
		rs, ok := allResults[qd]
		if !ok {
			continue
		}
		p("")
		title := "Fsync Results"
		if multiQD {
			title += fmt.Sprintf(" (QD=%d)", qd)
		}
		p("## %s", title)
		p("")
		writeResultTable(p, rs)
		p("")
		p("### Observations")
		p("")
		writeObservations(p, rs)
	}

	p("")
	p("## Methodology")
	p("")
	p("1. **Write pattern**: for each dirty size, a temporary file is pre-created")
	p("   and pre-extended to the target size. Each iteration overwrites the file")
	p("   from offset 0 with pseudo-random bytes (PCG-64 RNG, deterministic seed)")
	p("   to defeat filesystem-level compression, then calls `%s`.", syncMethod())
	p("2. **Latency measurement**: only the `%s` call is timed; the preceding", syncMethod())
	p("   `write(2)` is excluded from the measured latency.")
	p("3. **Concurrent mode**: when QD > 1, each of QD goroutines operates on its")
	p("   own temporary file, performing write+fsync iterations concurrently.")
	p("   This measures how fsync latency scales under device contention.")
	p("4. **Aggregation**: when multiple passes are requested, latency samples are")
	p("   pooled and throughput / ops-per-sec are averaged across all passes.")
	p("")
	p("---")
	p("")
	p("*Generated by [bench/disk/fsync](https://github.com/rjl493456442/bench)*")

	return nil
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	var (
		dir    = flag.String("dir", ".", "directory for temporary benchmark files")
		output = flag.String("output", "results.md", "path for the markdown results file")
		iters  = flag.Int("iters", 100, "number of write+fsync iterations per dirty size")
		passes = flag.Int("passes", 1, "number of passes per (dirty size, QD) pair")
		qdFlag = flag.String("qd", "1", `queue depth(s), comma-separated (e.g. "1,4,8")`)
	)
	flag.Parse()

	// Parse queue depth flag.
	var queueDepths []int
	for _, s := range strings.Split(*qdFlag, ",") {
		s = strings.TrimSpace(s)
		qd, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("invalid queue depth %q: %v", s, err)
		}
		if qd <= 0 {
			log.Fatalf("queue depth must be > 0, got %d", qd)
		}
		queueDepths = append(queueDepths, qd)
	}

	// Pre-generate random data buffer sized to the largest dirty size.
	maxSize := dirtySizes[len(dirtySizes)-1]
	dataBuf := initBuffer(maxSize)

	qdStrs := make([]string, len(queueDepths))
	for i, qd := range queueDepths {
		qdStrs[i] = strconv.Itoa(qd)
	}
	sizeStrs := make([]string, len(dirtySizes))
	for i, s := range dirtySizes {
		sizeStrs[i] = fmtSize(s)
	}

	log.Printf("Starting fsync benchmark")
	log.Printf("  Dir    : %s", *dir)
	log.Printf("  Sizes  : %s", strings.Join(sizeStrs, ", "))
	log.Printf("  QD     : %s", strings.Join(qdStrs, ", "))
	log.Printf("  Iters  : %d", *iters)
	log.Printf("  Passes : %d", *passes)
	log.Printf("  Method : %s", syncMethod())

	allResults := make(map[int][]*result) // qd -> results

	for _, qd := range queueDepths {
		log.Printf("=== Fsync benchmark (QD=%d) ===", qd)
		for _, ds := range dirtySizes {
			log.Printf("Benchmarking %s dirty data (QD=%d)...", fmtSize(ds), qd)
			r, err := benchmark(*dir, ds, *iters, *passes, qd, dataBuf)
			if err != nil {
				log.Fatalf("benchmark [%s/QD%d]: %v", fmtSize(ds), qd, err)
			}
			allResults[qd] = append(allResults[qd], r)
			log.Printf("  => %.2f MB/s, %.0f fsync/s, avg %s, p99 %s",
				r.throughputMBps(), r.opsPerSec(),
				fmtDuration(r.avg()), fmtDuration(r.percentile(99)))
		}
	}

	// Resolve dir to absolute path for storage detection.
	absDir, _ := filepath.Abs(*dir)
	storage, storageErr := detectStorage(absDir)
	if storageErr != nil {
		log.Printf("Warning: storage detection failed: %v", storageErr)
		storage = nil
	}

	if err := exportMarkdown(*output, absDir, *iters, *passes, allResults, queueDepths, storage); err != nil {
		log.Fatalf("export markdown: %v", err)
	}
	log.Printf("Results written to: %s", *output)
}
