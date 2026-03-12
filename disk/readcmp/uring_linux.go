//go:build linux

package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

// ── io_uring syscall wrappers ─────────────────────────────────────────────────

const (
	_SYS_IO_URING_SETUP    = 425
	_SYS_IO_URING_ENTER    = 426
	_SYS_IO_URING_REGISTER = 427
)

func sysIoUringSetup(entries uint32, params *ioUringParams) (int, error) {
	fd, _, errno := syscall.Syscall(
		_SYS_IO_URING_SETUP,
		uintptr(entries),
		uintptr(unsafe.Pointer(params)),
		0,
	)
	if errno != 0 {
		return -1, errno
	}
	return int(fd), nil
}

func sysIoUringEnter(fd int, toSubmit, minComplete, flags uint32) (int, error) {
	ret, _, errno := syscall.Syscall6(
		_SYS_IO_URING_ENTER,
		uintptr(fd),
		uintptr(toSubmit),
		uintptr(minComplete),
		uintptr(flags),
		0, 0,
	)
	if errno != 0 {
		return -1, errno
	}
	return int(ret), nil
}

func sysIoUringRegister(fd int, opcode uint32, arg unsafe.Pointer, nrArgs uint32) (int, error) {
	ret, _, errno := syscall.Syscall6(
		_SYS_IO_URING_REGISTER,
		uintptr(fd),
		uintptr(opcode),
		uintptr(arg),
		uintptr(nrArgs),
		0, 0,
	)
	if errno != 0 {
		return -1, errno
	}
	return int(ret), nil
}

// ── io_uring kernel ABI types ─────────────────────────────────────────────────

type ioSqringOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Flags       uint32
	Dropped     uint32
	Array       uint32
	Resv1       uint32
	UserAddr    uint64
}

type ioCqringOffsets struct {
	Head        uint32
	Tail        uint32
	RingMask    uint32
	RingEntries uint32
	Overflow    uint32
	Cqes        uint32
	Flags       uint32
	Resv1       uint32
	UserAddr    uint64
}

type ioUringParams struct {
	SqEntries    uint32
	CqEntries    uint32
	Flags        uint32
	SqThreadCPU  uint32
	SqThreadIdle uint32
	Features     uint32
	WqFd         uint32
	Resv         [3]uint32
	SqOff        ioSqringOffsets
	CqOff        ioCqringOffsets
}

// io_uring opcodes.
const (
	_IORING_OP_NOP  = 0
	_IORING_OP_READ = 22 // kernel 5.6+
)

// io_uring setup flags.
const (
	_IORING_SETUP_SQPOLL = 1 << 1
	_IORING_SETUP_SQ_AFF = 1 << 2
)

// io_uring enter flags.
const (
	_IORING_ENTER_GETEVENTS = 1 << 0
	_IORING_ENTER_SQ_WAKEUP = 1 << 1
)

// io_uring register opcodes.
const (
	_IORING_REGISTER_BUFFERS = 0
	_IORING_REGISTER_FILES   = 2
)

// SQ flags.
const _IORING_SQ_NEED_WAKEUP = 1 << 0

// mmap offsets.
const (
	_IORING_OFF_SQ_RING int64 = 0
	_IORING_OFF_CQ_RING int64 = 0x8000000
	_IORING_OFF_SQES    int64 = 0x10000000
)

// SQE/CQE sizes.
const (
	_sqeSize = 64
	_cqeSize = 16
)

// SQE byte offsets (struct io_uring_sqe, 64 bytes).
const (
	_sqeOpcode   = 0  // uint8
	_sqeFlags    = 1  // uint8
	_sqeIoprio   = 2  // uint16
	_sqeFd       = 4  // int32
	_sqeOff      = 8  // uint64
	_sqeAddr     = 16 // uint64
	_sqeLen      = 24 // uint32
	_sqeRWFlags  = 28 // uint32
	_sqeUserData = 32 // uint64
	_sqeBufIndex = 40 // uint16
)

// CQE byte offsets (struct io_uring_cqe, 16 bytes).
const (
	_cqeUserData = 0  // uint64
	_cqeRes      = 8  // int32
	_cqeFlags    = 12 // uint32
)

// io_uring feature bits.
const (
	_IORING_FEAT_SINGLE_MMAP  = 1 << 0
	_IORING_FEAT_NODROP       = 1 << 1
	_IORING_FEAT_SUBMIT_STABLE = 1 << 2
)

// ── io_uring ring management ──────────────────────────────────────────────────

type ioRing struct {
	fd     int
	params ioUringParams

	sqRingMem []byte
	cqRingMem []byte
	sqesMem   []byte

	// SQ ring pointers (into sqRingMem)
	sqHead    *uint32
	sqTail    *uint32
	sqMask    uint32
	sqEntries uint32
	sqFlags   *uint32
	sqArray   unsafe.Pointer // base of SQ array

	// CQ ring pointers (into cqRingMem)
	cqHead    *uint32
	cqTail    *uint32
	cqMask    uint32
	cqEntries uint32
	cqesBase  unsafe.Pointer // base of CQE array
}

func newRing(entries uint32, flags uint32) (*ioRing, error) {
	var params ioUringParams
	params.Flags = flags
	fd, err := sysIoUringSetup(entries, &params)
	if err != nil {
		return nil, fmt.Errorf("io_uring_setup(%d): %w", entries, err)
	}

	r := &ioRing{fd: fd, params: params}

	// Map SQ ring.
	sqSize := int(params.SqOff.Array + params.SqEntries*4)
	r.sqRingMem, err = syscall.Mmap(fd, _IORING_OFF_SQ_RING, sqSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE)
	if err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("mmap SQ ring: %w", err)
	}

	// Map CQ ring.
	cqSize := int(params.CqOff.Cqes + params.CqEntries*_cqeSize)
	r.cqRingMem, err = syscall.Mmap(fd, _IORING_OFF_CQ_RING, cqSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE)
	if err != nil {
		syscall.Munmap(r.sqRingMem)
		syscall.Close(fd)
		return nil, fmt.Errorf("mmap CQ ring: %w", err)
	}

	// Map SQEs array.
	sqesSize := int(params.SqEntries) * _sqeSize
	r.sqesMem, err = syscall.Mmap(fd, _IORING_OFF_SQES, sqesSize,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE)
	if err != nil {
		syscall.Munmap(r.cqRingMem)
		syscall.Munmap(r.sqRingMem)
		syscall.Close(fd)
		return nil, fmt.Errorf("mmap SQEs: %w", err)
	}

	// Resolve SQ ring pointers using unsafe.Add (the Go-blessed pattern).
	sqBase := unsafe.Pointer(&r.sqRingMem[0])
	r.sqHead = (*uint32)(unsafe.Add(sqBase, int(params.SqOff.Head)))
	r.sqTail = (*uint32)(unsafe.Add(sqBase, int(params.SqOff.Tail)))
	r.sqMask = *(*uint32)(unsafe.Add(sqBase, int(params.SqOff.RingMask)))
	r.sqEntries = *(*uint32)(unsafe.Add(sqBase, int(params.SqOff.RingEntries)))
	r.sqFlags = (*uint32)(unsafe.Add(sqBase, int(params.SqOff.Flags)))
	r.sqArray = unsafe.Add(sqBase, int(params.SqOff.Array))

	// Resolve CQ ring pointers.
	cqBase := unsafe.Pointer(&r.cqRingMem[0])
	r.cqHead = (*uint32)(unsafe.Add(cqBase, int(params.CqOff.Head)))
	r.cqTail = (*uint32)(unsafe.Add(cqBase, int(params.CqOff.Tail)))
	r.cqMask = *(*uint32)(unsafe.Add(cqBase, int(params.CqOff.RingMask)))
	r.cqEntries = *(*uint32)(unsafe.Add(cqBase, int(params.CqOff.RingEntries)))
	r.cqesBase = unsafe.Add(cqBase, int(params.CqOff.Cqes))

	return r, nil
}

// prepRead fills the next SQE with a READ.  Returns false if the SQ is full.
func (r *ioRing) prepRead(fd int32, bufPtr uintptr, length uint32, offset uint64, userData uint64) bool {
	tail := atomic.LoadUint32(r.sqTail)
	head := atomic.LoadUint32(r.sqHead)
	if tail-head >= r.sqEntries {
		return false
	}
	idx := tail & r.sqMask
	sqe := unsafe.Add(unsafe.Pointer(&r.sqesMem[0]), int(idx)*_sqeSize)

	// Zero the SQE first (some fields like personality, splice_fd_in matter).
	for i := 0; i < _sqeSize; i += 8 {
		*(*uint64)(unsafe.Add(sqe, i)) = 0
	}

	*(*uint8)(unsafe.Add(sqe, _sqeOpcode)) = _IORING_OP_READ
	*(*int32)(unsafe.Add(sqe, _sqeFd)) = fd
	*(*uint64)(unsafe.Add(sqe, _sqeOff)) = offset
	*(*uint64)(unsafe.Add(sqe, _sqeAddr)) = uint64(bufPtr)
	*(*uint32)(unsafe.Add(sqe, _sqeLen)) = length
	*(*uint64)(unsafe.Add(sqe, _sqeUserData)) = userData

	// Identity mapping in SQ array.
	*(*uint32)(unsafe.Add(r.sqArray, int(idx)*4)) = idx

	atomic.StoreUint32(r.sqTail, tail+1)
	return true
}

// submit submits pending SQEs.  In SQPOLL mode the kernel polls automatically
// and we only call enter if NEED_WAKEUP is set.
func (r *ioRing) submit(count uint32) (int, error) {
	if r.params.Flags&_IORING_SETUP_SQPOLL != 0 {
		if atomic.LoadUint32(r.sqFlags)&_IORING_SQ_NEED_WAKEUP == 0 {
			return int(count), nil
		}
		return sysIoUringEnter(r.fd, count, 0, _IORING_ENTER_SQ_WAKEUP)
	}
	return sysIoUringEnter(r.fd, count, 0, 0)
}

// submitAndWait submits and waits for at least minComplete CQEs.
func (r *ioRing) submitAndWait(count, minComplete uint32) (int, error) {
	return sysIoUringEnter(r.fd, count, minComplete, _IORING_ENTER_GETEVENTS)
}

// peekCQE returns the next CQE or ok=false if CQ is empty.
func (r *ioRing) peekCQE() (userData uint64, res int32, ok bool) {
	head := atomic.LoadUint32(r.cqHead)
	tail := atomic.LoadUint32(r.cqTail)
	if head == tail {
		return 0, 0, false
	}
	idx := head & r.cqMask
	cqe := unsafe.Add(r.cqesBase, int(idx)*_cqeSize)
	userData = *(*uint64)(unsafe.Add(cqe, _cqeUserData))
	res = *(*int32)(unsafe.Add(cqe, _cqeRes))
	return userData, res, true
}

// advanceCQ advances the CQ head by 1.
func (r *ioRing) advanceCQ() {
	atomic.StoreUint32(r.cqHead, atomic.LoadUint32(r.cqHead)+1)
}

// registerFiles registers fds for use with IOSQE_FIXED_FILE.
func (r *ioRing) registerFiles(fds []int32) error {
	_, err := sysIoUringRegister(r.fd, _IORING_REGISTER_FILES,
		unsafe.Pointer(&fds[0]), uint32(len(fds)))
	return err
}

func (r *ioRing) close() {
	syscall.Munmap(r.sqesMem)
	syscall.Munmap(r.cqRingMem)
	syscall.Munmap(r.sqRingMem)
	syscall.Close(r.fd)
}

// ── io_uring feature probe ────────────────────────────────────────────────────

func probeIOUring() (supported bool, features uint32) {
	var params ioUringParams
	fd, err := sysIoUringSetup(1, &params)
	if err != nil {
		return false, 0
	}
	syscall.Close(fd)
	return true, params.Features
}

func ioUringFeaturesString(feat uint32) string {
	var parts []string
	if feat&_IORING_FEAT_SINGLE_MMAP != 0 {
		parts = append(parts, "SINGLE_MMAP")
	}
	if feat&_IORING_FEAT_NODROP != 0 {
		parts = append(parts, "NODROP")
	}
	if feat&_IORING_FEAT_SUBMIT_STABLE != 0 {
		parts = append(parts, "SUBMIT_STABLE")
	}
	if len(parts) == 0 {
		return "none"
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += ", " + p
	}
	return result
}

// ── UringEngine ───────────────────────────────────────────────────────────────

// UringEngine benchmarks io_uring-based reads.
type UringEngine struct{}

func newUringEngine() (Engine, error) {
	ok, _ := probeIOUring()
	if !ok {
		return nil, fmt.Errorf("io_uring not supported on this kernel")
	}
	return &UringEngine{}, nil
}

func (e *UringEngine) Name() string { return "uring" }

func (e *UringEngine) Features() map[string]bool {
	ok, feat := probeIOUring()
	m := map[string]bool{
		"supported":     ok,
		"sqpoll":        ok,
		"single_mmap":   feat&_IORING_FEAT_SINGLE_MMAP != 0,
		"nodrop":        feat&_IORING_FEAT_NODROP != 0,
		"submit_stable": feat&_IORING_FEAT_SUBMIT_STABLE != 0,
	}
	return m
}

// Run executes the io_uring benchmark.  offsets describes the full I/O workload.
func (e *UringEngine) Run(cfg RunConfig, offsets []int64) (*RunResult, error) {
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
	fileFd := int32(f.Fd())

	qd := cfg.QueueDepth
	if qd < 1 {
		qd = 1
	}

	// Ring entries must be power of 2 and >= qd.
	ringEntries := nextPow2(uint32(qd))
	if ringEntries < 4 {
		ringEntries = 4
	}

	var flags uint32
	if cfg.SQPoll {
		flags |= _IORING_SETUP_SQPOLL
	}

	ring, err := newRing(ringEntries, flags)
	if err != nil {
		return nil, fmt.Errorf("newRing: %w", err)
	}
	defer ring.close()

	// Optional: register the file descriptor.
	if cfg.RegisteredFiles {
		if err := ring.registerFiles([]int32{fileFd}); err != nil {
			return nil, fmt.Errorf("register files: %w", err)
		}
	}

	bs := cfg.BlockSize

	// Allocate QD aligned I/O buffers.
	bufs := make([][]byte, qd)
	bufAddrs := make([]uintptr, qd)
	for i := range bufs {
		bufs[i] = newBuffer(bs)
		bufAddrs[i] = bufAddr(bufs[i])
	}
	defer func() {
		for _, b := range bufs {
			freeBuffer(b)
		}
	}()

	latencies := make([]time.Duration, len(offsets))
	submitTimes := make([]time.Time, qd) // per-slot submit timestamp

	before := getProcessMetrics()
	start := time.Now()

	errors := e.runLoop(ring, fileFd, bufAddrs, uint32(bs), offsets, qd, &loopMetrics{
		latencies:   latencies,
		submitTimes: submitTimes,
	})

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

type loopMetrics struct {
	latencies   []time.Duration
	submitTimes []time.Time
}

// runLoop drives the io_uring submission/completion loop.
// It keeps up to qd requests in-flight.  Each slot in bufAddrs corresponds
// to one in-flight request.  userData encodes (slot | completionIndex<<16).
func (e *UringEngine) runLoop(
	ring *ioRing,
	fd int32,
	bufAddrs []uintptr,
	blockSize uint32,
	offsets []int64,
	qd int,
	metrics *loopMetrics,
) int64 {
	n := len(offsets)
	submitted := 0
	completed := 0
	inflight := 0
	var errors int64
	freeSlots := make([]int, qd) // available buffer slots
	for i := range freeSlots {
		freeSlots[i] = qd - 1 - i
	}

	for completed < n {
		// Submit as many as possible up to QD.
		toSubmit := uint32(0)
		for submitted < n && inflight < qd && len(freeSlots) > 0 {
			slot := freeSlots[len(freeSlots)-1]
			freeSlots = freeSlots[:len(freeSlots)-1]

			userData := uint64(slot) | (uint64(submitted) << 16)
			ok := ring.prepRead(fd, bufAddrs[slot], blockSize,
				uint64(offsets[submitted]), userData)
			if !ok {
				// SQ full — flush what we have then retry.
				freeSlots = append(freeSlots, slot)
				break
			}
			if metrics != nil {
				metrics.submitTimes[slot] = time.Now()
			}
			submitted++
			inflight++
			toSubmit++
		}

		// Submit + wait for at least 1 completion.
		minWait := uint32(1)
		if toSubmit == 0 {
			minWait = 1 // just wait for completions
		}
		if _, err := ring.submitAndWait(toSubmit, minWait); err != nil {
			// EINTR is benign.
			if err != syscall.EINTR {
				errors++
			}
		}

		// Reap completions.
		for {
			userData, res, ok := ring.peekCQE()
			if !ok {
				break
			}
			ring.advanceCQ()

			slot := int(userData & 0xFFFF)
			idx := int(userData >> 16)

			if res < 0 {
				errors++
			}
			if metrics != nil && idx < len(metrics.latencies) {
				metrics.latencies[idx] = time.Since(metrics.submitTimes[slot])
			}
			freeSlots = append(freeSlots, slot)
			inflight--
			completed++
		}
	}
	return errors
}

// nextPow2 returns the smallest power of 2 >= v.
func nextPow2(v uint32) uint32 {
	if v == 0 {
		return 1
	}
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	return v + 1
}
