// Package main benchmarks concurrent read and write performance on a single
// storage device, measuring how far each side can be pushed when reads and
// writes compete for the same device.  Reads and writes run in independent
// goroutine pools (the R:W "mix"); writes are made durable with periodic
// fdatasync/F_FULLFSYNC, and device-level counters (iostat-style) are sampled
// around each run to expose queue depth and per-side wait times.
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ── Access mode ───────────────────────────────────────────────────────────────

type AccessMode uint8

const (
	ModeSeq  AccessMode = iota // sequential: each worker streams its own region
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

// ── Mix and configuration ───────────────────────────────────────────────────

// Mix is one (readers, writers) concurrency scenario.
type Mix struct {
	Readers int
	Writers int
}

func (m Mix) String() string { return fmt.Sprintf("%d:%d", m.Readers, m.Writers) }

// Config describes a single mix run.
type Config struct {
	FilePath   string
	FileSize   int64
	Duration   time.Duration
	ReadBlock  int
	WriteBlock int
	ReadMode   AccessMode
	WriteMode  AccessMode
	DirectIO   bool
	FsyncEvery int // fdatasync after every N writes (>=1)
}

// ── Per-side results ──────────────────────────────────────────────────────────

// maxSamplesPerWorker bounds the reservoir of latency samples kept per worker
// so a long, high-IOPS run does not exhaust memory.  Percentiles computed from
// a uniform reservoir of this size are accurate to well within a microsecond.
const maxSamplesPerWorker = 1 << 16 // 65536

// sideAccum is a single worker's running tally.  Each worker owns its own
// instance, so no locking is needed on the hot path.
type sideAccum struct {
	ops    int64
	bytes  int64
	latSum int64           // nanoseconds, exact across all ops
	latN   int64           // number of ops contributing to latSum
	res    []time.Duration // reservoir sample of latencies
	seen   int64           // total latencies offered to the reservoir
	rng    *rand.Rand
}

func newSideAccum(seed uint64) *sideAccum {
	return &sideAccum{
		res: make([]time.Duration, 0, maxSamplesPerWorker),
		rng: rand.New(rand.NewPCG(seed, 0x9e3779b97f4a7c15)),
	}
}

// record adds one completed I/O of the given latency to the tally.
func (a *sideAccum) record(lat time.Duration, n int) {
	a.ops++
	a.bytes += int64(n)
	a.latSum += int64(lat)
	a.latN++
	// Reservoir sampling (Algorithm R) for representative percentiles.
	if len(a.res) < maxSamplesPerWorker {
		a.res = append(a.res, lat)
	} else if j := a.rng.Int64N(a.seen + 1); j < maxSamplesPerWorker {
		a.res[j] = lat
	}
	a.seen++
}

// SideResult aggregates one side (read or write) across all its workers.
type SideResult struct {
	Workers   int
	Ops       int64
	Bytes     int64
	latSum    int64
	latN      int64
	Latencies []time.Duration // sorted ascending after merge
}

func (s *SideResult) merge(a *sideAccum) {
	s.Ops += a.ops
	s.Bytes += a.bytes
	s.latSum += a.latSum
	s.latN += a.latN
	s.Latencies = append(s.Latencies, a.res...)
}

func (s *SideResult) sort() {
	sort.Slice(s.Latencies, func(i, j int) bool { return s.Latencies[i] < s.Latencies[j] })
}

func (s *SideResult) IOPS(d time.Duration) float64 {
	if d == 0 {
		return 0
	}
	return float64(s.Ops) / d.Seconds()
}

func (s *SideResult) MBps(d time.Duration) float64 {
	if d == 0 {
		return 0
	}
	return float64(s.Bytes) / d.Seconds() / (1024 * 1024)
}

func (s *SideResult) Avg() time.Duration {
	if s.latN == 0 {
		return 0
	}
	return time.Duration(s.latSum / s.latN)
}

func (s *SideResult) Pct(p float64) time.Duration {
	n := len(s.Latencies)
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
	return s.Latencies[idx]
}

// MixResult holds everything measured for one mix scenario.
type MixResult struct {
	Mix     Mix
	Elapsed time.Duration
	Read    SideResult
	Write   SideResult

	CPUUser    time.Duration
	CPUSys     time.Duration
	VolCtxSw   int64
	InvolCtxSw int64

	Disk *DiskDelta // device-level counters; nil when unavailable
}

// ── Benchmark core ────────────────────────────────────────────────────────────

// runMix runs one mix scenario: cfg.Mix.Readers reader goroutines and
// cfg.Mix.Writers writer goroutines hammer the file for cfg.Duration, then the
// per-side tallies and device counters are collected.
func runMix(cfg Config, mix Mix, device string) (*MixResult, error) {
	// Open shared descriptors: one read fd, one write fd.  pread/pwrite are
	// safe to issue concurrently against a shared fd.
	var (
		rf, wf *os.File
		err    error
	)
	if mix.Readers > 0 {
		if cfg.DirectIO {
			rf, err = openDirect(cfg.FilePath)
		} else {
			rf, err = openBuffered(cfg.FilePath)
		}
		if err != nil {
			return nil, fmt.Errorf("open read fd: %w", err)
		}
		defer rf.Close()
	}
	if mix.Writers > 0 {
		if cfg.DirectIO {
			wf, err = openDirectRW(cfg.FilePath)
		} else {
			wf, err = openBufferedRW(cfg.FilePath)
		}
		if err != nil {
			return nil, fmt.Errorf("open write fd: %w", err)
		}
		defer wf.Close()
	}

	readBlocks := cfg.FileSize / int64(cfg.ReadBlock)
	writeBlocks := cfg.FileSize / int64(cfg.WriteBlock)
	if mix.Readers > 0 && readBlocks == 0 {
		return nil, fmt.Errorf("file too small for read block %s", fmtSize(cfg.ReadBlock))
	}
	if mix.Writers > 0 && writeBlocks == 0 {
		return nil, fmt.Errorf("file too small for write block %s", fmtSize(cfg.WriteBlock))
	}

	var stop atomic.Bool
	var wg sync.WaitGroup
	var barrier sync.WaitGroup
	barrier.Add(1)

	readAccs := make([]*sideAccum, mix.Readers)
	writeAccs := make([]*sideAccum, mix.Writers)

	// Reader workers.
	for i := 0; i < mix.Readers; i++ {
		acc := newSideAccum(uint64(i)*2 + 1)
		readAccs[i] = acc
		wg.Add(1)
		go func(id int, acc *sideAccum) {
			defer wg.Done()
			buf := newBuffer(cfg.ReadBlock)
			defer freeBuffer(buf)
			cursor := int64(id) % readBlocks // sequential start offset
			barrier.Wait()
			for !stop.Load() {
				var blk int64
				if cfg.ReadMode == ModeRand {
					blk = acc.rng.Int64N(readBlocks)
				} else {
					blk = cursor
					cursor = (cursor + 1) % readBlocks
				}
				t0 := time.Now()
				n, err := rf.ReadAt(buf, blk*int64(cfg.ReadBlock))
				lat := time.Since(t0)
				if err != nil {
					log.Printf("  read error @%d: %v", blk*int64(cfg.ReadBlock), err)
					continue
				}
				acc.record(lat, n)
			}
		}(i, acc)
	}

	// Writer workers.
	for i := 0; i < mix.Writers; i++ {
		acc := newSideAccum(uint64(i)*2 + 2)
		writeAccs[i] = acc
		wg.Add(1)
		go func(id int, acc *sideAccum) {
			defer wg.Done()
			buf := newBuffer(cfg.WriteBlock)
			defer freeBuffer(buf)
			fillRandom(buf, uint64(id))
			cursor := int64(id) % writeBlocks
			var sinceSync int
			barrier.Wait()
			for !stop.Load() {
				var blk int64
				if cfg.WriteMode == ModeRand {
					blk = acc.rng.Int64N(writeBlocks)
				} else {
					blk = cursor
					cursor = (cursor + 1) % writeBlocks
				}
				t0 := time.Now()
				n, err := wf.WriteAt(buf, blk*int64(cfg.WriteBlock))
				if err != nil {
					log.Printf("  write error @%d: %v", blk*int64(cfg.WriteBlock), err)
					continue
				}
				sinceSync++
				// Make the write durable every FsyncEvery ops; the sync cost is
				// charged to the op that triggered it so per-op latency reflects
				// the true cost of a durable write.
				if sinceSync >= cfg.FsyncEvery {
					if err := syncFile(wf); err != nil {
						log.Printf("  fsync error: %v", err)
					}
					sinceSync = 0
				}
				acc.record(time.Since(t0), n)
			}
		}(i, acc)
	}

	// Measure the device counters and CPU strictly around the timed window.
	procBefore := getProcessMetrics()
	diskBefore, diskOK := readDiskStat(device)
	start := time.Now()
	barrier.Done() // release all workers simultaneously

	time.Sleep(cfg.Duration)
	stop.Store(true)
	wg.Wait()

	elapsed := time.Since(start)
	diskAfter, diskOK2 := readDiskStat(device)
	procAfter := getProcessMetrics()
	delta := procAfter.Sub(procBefore)

	res := &MixResult{
		Mix:        mix,
		Elapsed:    elapsed,
		CPUUser:    delta.User,
		CPUSys:     delta.Sys,
		VolCtxSw:   delta.VolCtxSw,
		InvolCtxSw: delta.InvolCtxSw,
	}
	res.Read.Workers = mix.Readers
	res.Write.Workers = mix.Writers
	for _, a := range readAccs {
		res.Read.merge(a)
	}
	for _, a := range writeAccs {
		res.Write.merge(a)
	}
	res.Read.sort()
	res.Write.sort()

	if diskOK && diskOK2 {
		d := computeDiskDelta(diskBefore, diskAfter, elapsed)
		res.Disk = &d
	}
	return res, nil
}

// ── Dataset ───────────────────────────────────────────────────────────────────

const dataFileName = "bench_rwmix.bin"

// fillRandom writes deterministic pseudo-random bytes into buf to defeat
// filesystem/device compression and dedup.
func fillRandom(buf []byte, seed uint64) {
	rng := rand.New(rand.NewPCG(0xdeadbeef^seed, 0xcafebabe))
	for i := 0; i+8 <= len(buf); i += 8 {
		binary.LittleEndian.PutUint64(buf[i:], rng.Uint64())
	}
}

// initDataset creates or verifies the benchmark data file.
func initDataset(path string, size int64) error {
	if info, err := os.Stat(path); err == nil && info.Size() == size {
		log.Printf("Dataset already exists: %s (%.1f GB)", path, float64(size)/(1<<30))
		return nil
	}
	log.Printf("Initializing dataset: %s (%.1f GB)...", path, float64(size)/(1<<30))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create dataset: %w", err)
	}
	defer f.Close()

	const chunkSize = 8 * 1024 * 1024
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
		if written%(1024*1024*1024) == 0 {
			log.Printf("  written %.0f / %.0f GB", float64(written)/(1<<30), float64(size)/(1<<30))
		}
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync dataset: %w", err)
	}
	log.Printf("Dataset ready: %s", path)
	return nil
}

// ── Formatting helpers ──────────────────────────────────────────────────────

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
	case len(s) > 1 && (s[len(s)-1] == 'g' || s[len(s)-1] == 'G'):
		mul = 1024 * 1024 * 1024
		s2 = s[:len(s)-1]
	}
	var n int
	if _, err := fmt.Sscanf(s2, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return n * mul, nil
}
