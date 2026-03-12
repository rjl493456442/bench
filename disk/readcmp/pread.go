package main

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"
)

// PreadEngine benchmarks concurrent pread(2) using a goroutine pool.
type PreadEngine struct{}

func (e *PreadEngine) Name() string { return "pread" }

func (e *PreadEngine) Features() map[string]bool {
	return map[string]bool{"goroutine_pool": true}
}

// Run executes the pread benchmark.  offsets describes the full I/O workload
// (one read per entry); every entry is measured.
func (e *PreadEngine) Run(cfg RunConfig, offsets []int64) (*RunResult, error) {
	var f *os.File
	var err error
	if cfg.DirectIO {
		f, err = openDirect(cfg.FilePath)
	} else {
		f, err = openBuffered(cfg.FilePath)
	}
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	qd := cfg.QueueDepth
	if qd < 1 {
		qd = 1
	}
	bs := cfg.BlockSize

	// Allocate one aligned buffer per worker.
	bufs := make([][]byte, qd)
	for i := range bufs {
		bufs[i] = newBuffer(bs)
	}
	defer func() {
		for _, b := range bufs {
			freeBuffer(b)
		}
	}()

	latencies := make([]time.Duration, len(offsets))

	before := getProcessMetrics()
	start := time.Now()

	errors := e.runPool(f, bufs, offsets, qd, bs, latencies)

	duration := time.Since(start)
	after := getProcessMetrics()
	delta := after.Sub(before)

	res := &RunResult{
		EngineName: e.Name(),
		Config:     cfg,
		Reads:      int64(len(offsets)) - errors,
		Bytes:      (int64(len(offsets)) - errors) * int64(bs),
		Duration:   duration,
		Latencies:  latencies,
		CPUUser:    delta.User,
		CPUSys:     delta.Sys,
		VolCtxSw:   delta.VolCtxSw,
		InvolCtxSw: delta.InvolCtxSw,
		Errors:     errors,
	}
	res.SortLatencies()
	return res, nil
}

// runPool fans out pread calls across qd goroutines and returns the error count.
// If latencies is non-nil, per-read latencies are recorded at the matching index.
func (e *PreadEngine) runPool(f *os.File, bufs [][]byte, offsets []int64, qd, bs int, latencies []time.Duration) int64 {
	type workItem struct {
		idx    int
		offset int64
	}
	work := make(chan workItem, qd*2)

	var wg sync.WaitGroup
	var errCount int64
	var errMu sync.Mutex

	for w := 0; w < qd; w++ {
		wg.Add(1)
		go func(buf []byte) {
			defer wg.Done()
			for item := range work {
				var t0 time.Time
				if latencies != nil {
					t0 = time.Now()
				}
				_, err := f.ReadAt(buf, item.offset)
				if latencies != nil {
					latencies[item.idx] = time.Since(t0)
				}
				if err != nil {
					errMu.Lock()
					errCount++
					errMu.Unlock()
				}
			}
		}(bufs[w])
	}

	for i, off := range offsets {
		work <- workItem{idx: i, offset: off}
	}
	close(work)
	wg.Wait()

	return errCount
}

// bufAddr returns the address of the first byte for use in io_uring SQEs.
func bufAddr(b []byte) uintptr {
	return uintptr(unsafe.Pointer(&b[0]))
}
