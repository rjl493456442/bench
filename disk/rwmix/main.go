package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	var (
		dir        = flag.String("dir", ".", "directory for the benchmark data file")
		sizeFlag   = flag.String("size", "10g", "dataset size (e.g. 10g, 20480m); keep ≥10g so NVMe SLC cache doesn't skew writes")
		output     = flag.String("output", "results.md", "path for the markdown results file")
		duration   = flag.Duration("duration", 30*time.Second, "run time per mix scenario")
		mixFlag    = flag.String("mix", "8:8,16:0,0:16,16:16", `read:write worker mixes, comma-separated (e.g. "8:8,16:0,0:16")`)
		readMode   = flag.String("read-mode", "rand", "read access pattern: seq or rand")
		writeMode  = flag.String("write-mode", "rand", "write access pattern: seq or rand")
		readBSFlag = flag.String("read-bs", "4k", "read block size (e.g. 4k, 64k, 1m)")
		wrBSFlag   = flag.String("write-bs", "4k", "write block size (e.g. 4k, 64k, 1m)")
		fsyncEvery = flag.Int("fsync-every", 1, "fdatasync after every N writes (1 = every write)")
		directIO   = flag.Bool("direct-io", true, "use direct I/O (O_DIRECT / F_NOCACHE) to bypass the page cache")
		skipInit   = flag.Bool("skip-init", false, "skip dataset creation if a correct-size file exists")
	)
	flag.Parse()

	size, err := parseSize(*sizeFlag)
	if err != nil {
		log.Fatalf("invalid --size: %v", err)
	}
	readBS, err := parseSize(*readBSFlag)
	if err != nil {
		log.Fatalf("invalid --read-bs: %v", err)
	}
	wrBS, err := parseSize(*wrBSFlag)
	if err != nil {
		log.Fatalf("invalid --write-bs: %v", err)
	}
	if *fsyncEvery < 1 {
		log.Fatalf("--fsync-every must be ≥ 1")
	}
	// O_DIRECT requires offsets and lengths aligned to the logical block size
	// (≤512 B everywhere); unaligned sizes fail with a cryptic EINVAL.
	if *directIO && (readBS%512 != 0 || wrBS%512 != 0) {
		log.Fatalf("--read-bs and --write-bs must be multiples of 512 with --direct-io (got %s, %s)",
			fmtSize(readBS), fmtSize(wrBS))
	}
	rMode := parseMode("read-mode", *readMode)
	wMode := parseMode("write-mode", *writeMode)
	mixes := parseMixes(*mixFlag)

	dataPath := filepath.Join(*dir, dataFileName)
	cfg := Config{
		FilePath:   dataPath,
		FileSize:   int64(size),
		Duration:   *duration,
		ReadBlock:  readBS,
		WriteBlock: wrBS,
		ReadMode:   rMode,
		WriteMode:  wMode,
		DirectIO:   *directIO,
		FsyncEvery: *fsyncEvery,
	}

	// ── Dataset ───────────────────────────────────────────────────────────
	if *skipInit {
		if info, err := os.Stat(dataPath); err != nil || info.Size() != cfg.FileSize {
			log.Fatalf("dataset missing or wrong size (--skip-init): %s", dataPath)
		}
	} else if err := initDataset(dataPath, cfg.FileSize); err != nil {
		log.Fatalf("dataset init: %v", err)
	}

	// ── Storage detection (also yields the /proc/diskstats device name) ─────
	absDir, _ := filepath.Abs(*dir)
	storage, storageErr := detectStorage(absDir)
	if storageErr != nil {
		log.Printf("Warning: storage detection failed: %v", storageErr)
	}
	device := ""
	if storage != nil {
		device = storage.Device
	}

	// ── Plan ────────────────────────────────────────────────────────────────
	mixStrs := make([]string, len(mixes))
	for i, m := range mixes {
		mixStrs[i] = m.String()
	}
	log.Printf("Starting concurrent read/write benchmark")
	log.Printf("  File     : %s (%.1f GB)", dataPath, float64(cfg.FileSize)/(1<<30))
	log.Printf("  Mixes    : %s", strings.Join(mixStrs, ", "))
	log.Printf("  Duration : %s each", cfg.Duration)
	log.Printf("  Read     : %s, %s", rMode, fmtSize(readBS))
	log.Printf("  Write    : %s, %s, %s every %d", wMode, fmtSize(wrBS), syncMethod(), *fsyncEvery)
	log.Printf("  I/O Mode : %s", ioModeLabel(cfg.DirectIO))
	if device != "" {
		log.Printf("  Device   : %s", device)
	} else {
		log.Printf("  Device   : (device-level counters unavailable)")
	}

	// ── Run ─────────────────────────────────────────────────────────────────
	var results []*MixResult
	for _, m := range mixes {
		// Drop the page cache between mixes so a prior run's data can't be served
		// from RAM (matters most for buffered I/O).
		if err := dropPageCache(dataPath); err != nil {
			log.Printf("  Warning: could not drop page cache: %v", err)
		}
		log.Printf("=== mix %s (%s) ===", m, cfg.Duration)
		r, err := runMix(cfg, m, device)
		if err != nil {
			log.Fatalf("mix %s: %v", m, err)
		}
		logMixResult(r)
		results = append(results, r)
	}

	if err := exportMarkdown(*output, cfg, results, storage); err != nil {
		log.Fatalf("export: %v", err)
	}
	log.Printf("Results written to: %s", *output)
}

func ioModeLabel(direct bool) string {
	if direct {
		return directIOMethod()
	}
	return "buffered"
}

func logMixResult(r *MixResult) {
	if r.Mix.Readers > 0 {
		log.Printf("  read : %.1f MB/s, %s IOPS, avg %s, p99 %s",
			r.Read.MBps(r.Elapsed), fmtFloat(r.Read.IOPS(r.Elapsed)),
			fmtDuration(r.Read.Avg()), fmtDuration(r.Read.Pct(99)))
	}
	if r.Mix.Writers > 0 {
		log.Printf("  write: %.1f MB/s, %s IOPS, avg %s, p99 %s",
			r.Write.MBps(r.Elapsed), fmtFloat(r.Write.IOPS(r.Elapsed)),
			fmtDuration(r.Write.Avg()), fmtDuration(r.Write.Pct(99)))
	}
	if d := r.Disk; d != nil {
		log.Printf("  disk : r_await %s, w_await %s, aqu-sz %.2f, %%util %.1f",
			fmtMs(d.AvgReadWaitMs), fmtMs(d.AvgWriteWaitMs), d.AvgQueueDepth, d.UtilPct)
	}
}

// ── Flag parsing helpers ────────────────────────────────────────────────────

func parseMode(flagName, s string) AccessMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "seq", "sequential":
		return ModeSeq
	case "rand", "random":
		return ModeRand
	default:
		log.Fatalf("--%s: want seq or rand, got %q", flagName, s)
		return ModeRand
	}
}

func parseMixes(s string) []Mix {
	var mixes []Mix
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rw := strings.SplitN(part, ":", 2)
		if len(rw) != 2 {
			log.Fatalf("invalid mix %q: want readers:writers", part)
		}
		r, err1 := strconv.Atoi(strings.TrimSpace(rw[0]))
		w, err2 := strconv.Atoi(strings.TrimSpace(rw[1]))
		if err1 != nil || err2 != nil || r < 0 || w < 0 || r+w == 0 {
			log.Fatalf("invalid mix %q: readers/writers must be non-negative and not both zero", part)
		}
		mixes = append(mixes, Mix{Readers: r, Writers: w})
	}
	if len(mixes) == 0 {
		log.Fatalf("no valid mixes parsed from %q", s)
	}
	return mixes
}
