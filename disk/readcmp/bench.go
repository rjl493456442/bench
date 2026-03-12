// Package main benchmarks io_uring versus conventional concurrent pread under
// identical workloads, measuring read latency, throughput, CPU utilisation,
// syscall overhead, and tail behaviour across queue depths, block sizes,
// access patterns, and concurrency levels.
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"time"
)

// ── Access mode ───────────────────────────────────────────────────────────────

type AccessMode uint8

const (
	ModeSeq  AccessMode = iota // sequential: start-to-finish
	ModeRand                   // random: uniformly random aligned offsets
)

func (m AccessMode) String() string {
	if m == ModeSeq {
		return "sequential"
	}
	return "random"
}

func (m AccessMode) Short() string {
	if m == ModeSeq {
		return "seq"
	}
	return "rand"
}

// ── Run configuration ─────────────────────────────────────────────────────────

// RunConfig describes a single benchmark run.
type RunConfig struct {
	FilePath   string
	FileSize   int64
	BlockSize  int
	QueueDepth int
	Mode       AccessMode
	DirectIO   bool

	// io_uring features (ignored by pread engine)
	RegisteredFiles   bool
	RegisteredBuffers bool
	SQPoll            bool
	BatchSize         int // 0 = submit individually
}

// ── Engine interface ──────────────────────────────────────────────────────────

// Engine is the benchmark engine interface.  Both the pread goroutine-pool
// engine and the io_uring engine implement it.
type Engine interface {
	Name() string
	Run(cfg RunConfig, offsets []int64) (*RunResult, error)
	Features() map[string]bool
}

// ── Run result ────────────────────────────────────────────────────────────────

// RunResult holds measurements from a single benchmark run.
type RunResult struct {
	EngineName string
	Config     RunConfig
	Reads      int64
	Bytes      int64
	Duration   time.Duration
	Latencies  []time.Duration // sorted ascending after SortLatencies()

	// System metrics (best-effort, zero when unavailable)
	CPUUser    time.Duration
	CPUSys     time.Duration
	VolCtxSw   int64
	InvolCtxSw int64
	Errors     int64
}

func (r *RunResult) ThroughputMBps() float64 {
	if r.Duration == 0 {
		return 0
	}
	return float64(r.Bytes) / r.Duration.Seconds() / (1024 * 1024)
}

func (r *RunResult) IOPS() float64 {
	if r.Duration == 0 {
		return 0
	}
	return float64(r.Reads) / r.Duration.Seconds()
}

func (r *RunResult) AvgLat() time.Duration {
	if len(r.Latencies) == 0 {
		return 0
	}
	var total int64
	for _, l := range r.Latencies {
		total += int64(l)
	}
	return time.Duration(total / int64(len(r.Latencies)))
}

func (r *RunResult) Pct(p float64) time.Duration {
	n := len(r.Latencies)
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
	return r.Latencies[idx]
}

func (r *RunResult) MinLat() time.Duration {
	if len(r.Latencies) == 0 {
		return 0
	}
	return r.Latencies[0]
}

func (r *RunResult) MaxLat() time.Duration {
	if len(r.Latencies) == 0 {
		return 0
	}
	return r.Latencies[len(r.Latencies)-1]
}

func (r *RunResult) CPUPerOp() time.Duration {
	total := r.CPUUser + r.CPUSys
	if r.Reads == 0 {
		return 0
	}
	return time.Duration(int64(total) / r.Reads)
}

func (r *RunResult) SortLatencies() {
	sort.Slice(r.Latencies, func(i, j int) bool {
		return r.Latencies[i] < r.Latencies[j]
	})
}

// ── Offset generation ─────────────────────────────────────────────────────────

// generateOffsets creates a deterministic sequence of block-aligned offsets.
// Both engines receive the same slice so the comparison is fair.
func generateOffsets(fileSize int64, blockSize int, mode AccessMode, count int) []int64 {
	nBlocks := int(fileSize / int64(blockSize))
	if nBlocks == 0 {
		return nil
	}
	offsets := make([]int64, count)
	switch mode {
	case ModeSeq:
		for i := range offsets {
			offsets[i] = int64(i%nBlocks) * int64(blockSize)
		}
	case ModeRand:
		perm := make([]int64, nBlocks)
		for i := range perm {
			perm[i] = int64(i) * int64(blockSize)
		}
		rng := rand.New(rand.NewPCG(0xdeadbeef, 0xcafebabe))
		rng.Shuffle(len(perm), func(i, j int) {
			perm[i], perm[j] = perm[j], perm[i]
		})
		for i := range offsets {
			offsets[i] = perm[i%nBlocks]
		}
	}
	return offsets
}

// ── Dataset ───────────────────────────────────────────────────────────────────

const dataFileName = "bench_readcmp.bin"

// initDataset creates or verifies the benchmark data file filled with
// deterministic pseudo-random bytes to defeat filesystem compression.
func initDataset(path string, size int64) error {
	if info, err := os.Stat(path); err == nil && info.Size() == size {
		log.Printf("Dataset already exists: %s (%.0f MB)", path, float64(size)/(1024*1024))
		return nil
	}
	log.Printf("Initializing dataset: %s (%.0f MB)...", path, float64(size)/(1024*1024))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	defer f.Close()

	const chunkSize = 4 * 1024 * 1024
	chunk := make([]byte, chunkSize)
	rng := rand.New(rand.NewPCG(0xdeadbeef, 0xcafebabe))

	var written int64
	for written < size {
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

func parseSize(s string) (int, error) {
	s2 := s
	mul := 1
	switch {
	case len(s) > 1 && (s[len(s)-1] == 'k' || s[len(s)-1] == 'K'):
		mul = 1024
		s2 = s[:len(s)-1]
	case len(s) > 1 && (s[len(s)-1] == 'm' || s[len(s)-1] == 'M'):
		mul = 1024 * 1024
		s2 = s[:len(s)-1]
	}
	var n int
	if _, err := fmt.Sscanf(s2, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return n * mul, nil
}
