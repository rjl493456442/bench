package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	var (
		dir        = flag.String("dir", ".", "directory for benchmark data file")
		sizeMB     = flag.Int("size", 512, "dataset size in MB")
		output     = flag.String("output", "results.md", "path for the markdown results file")
		skipInit   = flag.Bool("skip-init", false, "skip dataset creation if correct-size file exists")
		passes     = flag.Int("passes", 1, "number of passes; latencies are pooled across passes")
		blockFlag  = flag.String("block-sizes", "4k,64k,1m", "block sizes, comma-separated (e.g. 4k,64k,1m)")
		modeFlag   = flag.String("modes", "rand", `access patterns: "seq", "rand", or "both"`)
		qdFlag     = flag.String("qds", "1,4,16,64", "queue depths, comma-separated")
		engineFlag = flag.String("engines", "pread,uring", `engines: "pread", "uring", or "pread,uring"`)
		directIO   = flag.Bool("direct-io", true, "use direct I/O (O_DIRECT / F_NOCACHE)")
		pinCPUFlag = flag.Int("pin-cpu", -1, "pin benchmark to a specific CPU core (-1 = no pinning)")
		sqpoll     = flag.Bool("uring-sqpoll", false, "enable SQPOLL mode for io_uring")
		regFiles   = flag.Bool("uring-registered", false, "register files with io_uring")
	)
	flag.Parse()

	// ── Parse block sizes ─────────────────────────────────────────────────
	var blockSizes []int
	for _, s := range strings.Split(*blockFlag, ",") {
		bs, err := parseSize(strings.TrimSpace(s))
		if err != nil {
			log.Fatalf("invalid block size %q: %v", s, err)
		}
		blockSizes = append(blockSizes, bs)
	}

	// ── Parse modes ───────────────────────────────────────────────────────
	var modes []AccessMode
	switch *modeFlag {
	case "seq", "sequential":
		modes = []AccessMode{ModeSeq}
	case "rand", "random":
		modes = []AccessMode{ModeRand}
	case "both", "all":
		modes = []AccessMode{ModeSeq, ModeRand}
	default:
		log.Fatalf("unknown --modes %q: use seq, rand, or both", *modeFlag)
	}

	// ── Parse queue depths ────────────────────────────────────────────────
	var qds []int
	for _, s := range strings.Split(*qdFlag, ",") {
		qd, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil || qd <= 0 {
			log.Fatalf("invalid queue depth %q", s)
		}
		qds = append(qds, qd)
	}

	// ── Parse engines ─────────────────────────────────────────────────────
	engineNames := strings.Split(*engineFlag, ",")
	engines := make(map[string]Engine)
	for _, name := range engineNames {
		name = strings.TrimSpace(name)
		switch name {
		case "pread":
			engines[name] = &PreadEngine{}
		case "uring":
			eng, err := newUringEngine()
			if err != nil {
				log.Printf("Warning: %v – skipping uring engine", err)
				continue
			}
			engines[name] = eng
		default:
			log.Fatalf("unknown engine %q: use pread or uring", name)
		}
	}
	if len(engines) == 0 {
		log.Fatalf("no engines available")
	}

	// ── CPU pinning ───────────────────────────────────────────────────────
	if *pinCPUFlag >= 0 {
		if err := pinCPU(*pinCPUFlag); err != nil {
			log.Printf("Warning: CPU pinning failed: %v", err)
		} else {
			log.Printf("Pinned to CPU %d", *pinCPUFlag)
		}
	}

	// ── Dataset ───────────────────────────────────────────────────────────
	fileSize := int64(*sizeMB) * 1024 * 1024
	dataPath := filepath.Join(*dir, dataFileName)

	if !*skipInit {
		if err := initDataset(dataPath, fileSize); err != nil {
			log.Fatalf("dataset init: %v", err)
		}
	} else {
		if _, err := os.Stat(dataPath); err != nil {
			log.Fatalf("dataset not found (--skip-init): %v", err)
		}
	}

	// ── Print plan ────────────────────────────────────────────────────────
	engineList := make([]string, 0, len(engines))
	for name := range engines {
		engineList = append(engineList, name)
	}
	sort.Strings(engineList)

	bsStrs := make([]string, len(blockSizes))
	for i, bs := range blockSizes {
		bsStrs[i] = fmtSize(bs)
	}
	qdStrs := make([]string, len(qds))
	for i, qd := range qds {
		qdStrs[i] = strconv.Itoa(qd)
	}
	modeStrs := make([]string, len(modes))
	for i, m := range modes {
		modeStrs[i] = m.Short()
	}

	totalRuns := len(engines) * len(blockSizes) * len(modes) * len(qds) * *passes
	log.Printf("Starting benchmark (%d total runs)", totalRuns)
	log.Printf("  File     : %s (%.0f MB)", dataPath, float64(fileSize)/(1024*1024))
	log.Printf("  Engines  : %s", strings.Join(engineList, ", "))
	log.Printf("  Blocks   : %s", strings.Join(bsStrs, ", "))
	log.Printf("  Modes    : %s", strings.Join(modeStrs, ", "))
	log.Printf("  QDs      : %s", strings.Join(qdStrs, ", "))
	log.Printf("  Passes   : %d", *passes)
	log.Printf("  Direct IO: %v", *directIO)
	log.Printf("  Kernel   : %s", kernelVersion())
	uringOK, uringFeats := probeIOUring()
	if uringOK {
		log.Printf("  io_uring : features=%s", ioUringFeaturesString(uringFeats))
	}

	// ── Run benchmarks ────────────────────────────────────────────────────
	type resultKey struct {
		mode   AccessMode
		bs     int
		qd     int
		engine string
	}
	results := make(map[resultKey]*RunResult)

	for _, mode := range modes {
		for _, bs := range blockSizes {
			// Each run reads the entire file once: iters = fileSize / blockSize.
			nReads := int(fileSize / int64(bs))
			if nReads == 0 {
				log.Printf("Skipping %s/%s: file too small for block size", fmtSize(bs), mode)
				continue
			}
			offsets := generateOffsets(fileSize, bs, mode, nReads)

			log.Printf("=== %s / %s (%d reads per pass) ===", mode, fmtSize(bs), nReads)

			for _, qd := range qds {
				for _, engineName := range engineList {
					eng := engines[engineName]
					cfg := RunConfig{
						FilePath:        dataPath,
						FileSize:        fileSize,
						BlockSize:       bs,
						QueueDepth:      qd,
						Mode:            mode,
						DirectIO:        *directIO,
						SQPoll:          *sqpoll && engineName == "uring",
						RegisteredFiles: *regFiles && engineName == "uring",
					}
					var combined *RunResult
					for pass := 1; pass <= *passes; pass++ {
						// Drop page cache before each run for fairness.
						if err := dropPageCache(dataPath); err != nil {
							log.Printf("  Warning: could not drop page cache: %v", err)
						}
						log.Printf("  [%s/%s/%s/QD%d] pass %d/%d...",
							engineName, fmtSize(bs), mode.Short(), qd, pass, *passes)

						r, err := eng.Run(cfg, offsets)
						if err != nil {
							log.Fatalf("  %s failed: %v", engineName, err)
						}

						log.Printf("    => %.2f MB/s, %s IOPS, avg %s, p99 %s, ctx_sw %d",
							r.ThroughputMBps(), fmtFloat(r.IOPS()),
							fmtDuration(r.AvgLat()), fmtDuration(r.Pct(99)),
							r.VolCtxSw+r.InvolCtxSw)

						if combined == nil {
							combined = r
						} else {
							combined.Reads += r.Reads
							combined.Bytes += r.Bytes
							combined.Duration += r.Duration
							combined.Latencies = append(combined.Latencies, r.Latencies...)
							combined.CPUUser += r.CPUUser
							combined.CPUSys += r.CPUSys
							combined.VolCtxSw += r.VolCtxSw
							combined.InvolCtxSw += r.InvolCtxSw
							combined.Errors += r.Errors
						}
					}
					combined.SortLatencies()
					key := resultKey{mode: mode, bs: bs, qd: qd, engine: engineName}
					results[key] = combined
				}
			}
		}
	}

	// ── Build cases for export ────────────────────────────────────────────
	var cases []benchCase
	for _, mode := range modes {
		for _, bs := range blockSizes {
			for _, qd := range qds {
				c := benchCase{Mode: mode, BlockSize: bs, QD: qd}
				pKey := resultKey{mode: mode, bs: bs, qd: qd, engine: "pread"}
				uKey := resultKey{mode: mode, bs: bs, qd: qd, engine: "uring"}
				c.Pread = results[pKey]
				c.Uring = results[uKey]
				if c.Pread != nil || c.Uring != nil {
					cases = append(cases, c)
				}
			}
		}
	}

	// ── Storage detection ─────────────────────────────────────────────────
	absDir, _ := filepath.Abs(*dir)
	storage, storageErr := detectStorage(absDir)
	if storageErr != nil {
		log.Printf("Warning: storage detection failed: %v", storageErr)
		storage = nil
	}

	// ── Export ─────────────────────────────────────────────────────────────
	if err := exportMarkdown(*output, fileSize, dataPath, *directIO,
		*passes, cases, storage); err != nil {
		log.Fatalf("export: %v", err)
	}
	log.Printf("Results written to: %s", *output)
}
