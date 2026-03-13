# io_uring vs pread – Read Performance Comparison

## Test Environment

| Parameter | Value |
|-----------|-------|
| Date | 2026-03-12 22:47:58 |
| OS | linux |
| Architecture | amd64 |
| Kernel | 6.17.0-14-generic |
| Data File | `bench_readcmp.bin` |
| File Size | 5120 MB |
| I/O Mode | buffered |
| Workload | entire file per pass |
| Passes | 3 |
| io_uring features | SINGLE_MMAP, NODROP, SUBMIT_STABLE |

## Sequential Read – 4 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 2785.54 MB/s | 713.10K | 1.3 µs | 0.3 µs | 0.6 µs | 44.9 µs | 51.8 µs | 0.7 µs | 422517 | 0 |
| 1 | uring | 2790.66 MB/s | 714.41K | 1.4 µs | 0.3 µs | 0.6 µs | 46.0 µs | 51.9 µs | 0.6 µs | 138746 | 0 |
| 4 | pread | 2790.15 MB/s | 714.28K | 3.5 µs | 0.4 µs | 30.2 µs | 51.7 µs | 61.4 µs | 1.1 µs | 812671 | 0 |
| 4 | uring | 2785.63 MB/s | 713.12K | 5.5 µs | 1.2 µs | 37.8 µs | 51.5 µs | 54.6 µs | 0.5 µs | 140676 | 0 |
| 16 | pread | 2795.57 MB/s | 715.67K | 14.6 µs | 1.2 µs | 52.9 µs | 61.6 µs | 72.8 µs | 3.1 µs | 3189766 | 0 |
| 16 | uring | 2795.35 MB/s | 715.61K | 21.8 µs | 9.2 µs | 57.0 µs | 58.6 µs | 78.5 µs | 0.6 µs | 141447 | 0 |
| 64 | pread | 2680.85 MB/s | 686.30K | 39.4 µs | 37.7 µs | 85.0 µs | 110.8 µs | 165.5 µs | 4.5 µs | 3985973 | 0 |
| 64 | uring | 3465.35 MB/s | 887.13K | 71.1 µs | 71.2 µs | 103.4 µs | 143.1 µs | 183.0 µs | 0.6 µs | 97783 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (714409 vs 713097). Average latency ratio pread/uring = **0.95**. Context switches: pread=422517, uring=138746.
- **QD=4**: pread achieves **1.00x** IOPS advantage (713121 vs 714277). Average latency ratio pread/uring = **0.64**. Context switches: pread=812671, uring=140676.
- **QD=16**: pread achieves **1.00x** IOPS advantage (715611 vs 715667). Average latency ratio pread/uring = **0.67**. Context switches: pread=3189766, uring=141447.
- **QD=64**: uring achieves **1.29x** IOPS advantage (887129 vs 686298). Average latency ratio pread/uring = **0.55**. Context switches: pread=3985973, uring=97783.

## Sequential Read – 8 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 2790.42 MB/s | 357.17K | 2.7 µs | 0.4 µs | 22.1 µs | 49.7 µs | 61.5 µs | 1.2 µs | 424100 | 0 |
| 1 | uring | 2798.35 MB/s | 358.19K | 2.7 µs | 0.5 µs | 22.3 µs | 53.2 µs | 60.9 µs | 0.9 µs | 143721 | 0 |
| 4 | pread | 2644.83 MB/s | 338.54K | 9.2 µs | 0.7 µs | 60.4 µs | 79.1 µs | 85.0 µs | 2.0 µs | 835642 | 0 |
| 4 | uring | 2544.11 MB/s | 325.65K | 12.1 µs | 1.7 µs | 79.4 µs | 82.2 µs | 90.4 µs | 0.8 µs | 120006 | 0 |
| 16 | pread | 2599.79 MB/s | 332.77K | 46.9 µs | 46.0 µs | 82.8 µs | 89.8 µs | 120.3 µs | 5.2 µs | 2578067 | 0 |
| 16 | uring | 2790.06 MB/s | 357.13K | 44.2 µs | 46.9 µs | 58.0 µs | 66.0 µs | 90.8 µs | 0.9 µs | 147313 | 0 |
| 64 | pread | 3318.10 MB/s | 424.72K | 94.2 µs | 90.1 µs | 160.9 µs | 215.2 µs | 282.6 µs | 7.8 µs | 2070021 | 0 |
| 64 | uring | 5060.78 MB/s | 647.78K | 98.0 µs | 91.9 µs | 165.2 µs | 200.7 µs | 234.4 µs | 0.9 µs | 90032 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (358189 vs 357174). Average latency ratio pread/uring = **0.97**. Context switches: pread=424100, uring=143721.
- **QD=4**: pread achieves **1.04x** IOPS advantage (325646 vs 338538). Average latency ratio pread/uring = **0.75**. Context switches: pread=835642, uring=120006.
- **QD=16**: uring achieves **1.07x** IOPS advantage (357127 vs 332773). Average latency ratio pread/uring = **1.06**. Context switches: pread=2578067, uring=147313.
- **QD=64**: uring achieves **1.53x** IOPS advantage (647780 vs 424717). Average latency ratio pread/uring = **0.96**. Context switches: pread=2070021, uring=90032.

## Sequential Read – 16 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 2800.25 MB/s | 179.22K | 5.4 µs | 0.6 µs | 50.5 µs | 59.0 µs | 63.0 µs | 2.2 µs | 448674 | 0 |
| 1 | uring | 2802.63 MB/s | 179.37K | 5.5 µs | 0.6 µs | 53.2 µs | 57.9 µs | 61.7 µs | 1.4 µs | 148992 | 0 |
| 4 | pread | 2738.09 MB/s | 175.24K | 22.1 µs | 1.9 µs | 77.7 µs | 80.5 µs | 93.9 µs | 3.1 µs | 759088 | 0 |
| 4 | uring | 2616.42 MB/s | 167.45K | 23.7 µs | 7.1 µs | 80.4 µs | 83.4 µs | 90.2 µs | 1.4 µs | 150402 | 0 |
| 16 | pread | 3049.75 MB/s | 195.18K | 81.5 µs | 81.5 µs | 109.1 µs | 120.9 µs | 166.4 µs | 9.5 µs | 1663062 | 0 |
| 16 | uring | 3834.57 MB/s | 245.41K | 64.9 µs | 58.2 µs | 84.8 µs | 87.6 µs | 122.8 µs | 1.5 µs | 138378 | 0 |
| 64 | pread | 5521.36 MB/s | 353.37K | 146.1 µs | 136.7 µs | 248.9 µs | 318.1 µs | 476.4 µs | 12.8 µs | 1197794 | 0 |
| 64 | uring | 6533.11 MB/s | 418.12K | 152.6 µs | 149.5 µs | 193.2 µs | 233.3 µs | 564.4 µs | 1.6 µs | 87210 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (179368 vs 179216). Average latency ratio pread/uring = **0.98**. Context switches: pread=448674, uring=148992.
- **QD=4**: pread achieves **1.05x** IOPS advantage (167451 vs 175238). Average latency ratio pread/uring = **0.93**. Context switches: pread=759088, uring=150402.
- **QD=16**: uring achieves **1.26x** IOPS advantage (245412 vs 195184). Average latency ratio pread/uring = **1.26**. Context switches: pread=1663062, uring=138378.
- **QD=64**: uring achieves **1.18x** IOPS advantage (418119 vs 353367). Average latency ratio pread/uring = **0.96**. Context switches: pread=1197794, uring=87210.

## Sequential Read – 32 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 2620.95 MB/s | 83.87K | 11.7 µs | 1.1 µs | 77.7 µs | 80.8 µs | 89.1 µs | 3.8 µs | 405323 | 0 |
| 1 | uring | 2655.44 MB/s | 84.97K | 11.7 µs | 1.1 µs | 76.9 µs | 79.0 µs | 82.6 µs | 2.6 µs | 162075 | 0 |
| 4 | pread | 2693.48 MB/s | 86.19K | 46.1 µs | 46.1 µs | 80.9 µs | 82.8 µs | 87.4 µs | 5.8 µs | 756636 | 0 |
| 4 | uring | 2643.40 MB/s | 84.59K | 47.1 µs | 46.3 µs | 82.5 µs | 85.2 µs | 102.1 µs | 2.7 µs | 177798 | 0 |
| 16 | pread | 4923.67 MB/s | 157.56K | 101.0 µs | 98.8 µs | 120.0 µs | 145.2 µs | 203.1 µs | 13.9 µs | 979825 | 0 |
| 16 | uring | 5404.95 MB/s | 172.96K | 92.3 µs | 87.1 µs | 147.8 µs | 185.3 µs | 325.7 µs | 2.7 µs | 116675 | 0 |
| 64 | pread | 6572.92 MB/s | 210.33K | 303.4 µs | 267.3 µs | 556.6 µs | 737.2 µs | 1.02 ms | 18.1 µs | 779943 | 0 |
| 64 | uring | 6613.14 MB/s | 211.62K | 302.1 µs | 233.9 µs | 666.9 µs | 929.9 µs | 1.11 ms | 2.8 µs | 93429 | 0 |

### Observations

- **QD=1**: uring achieves **1.01x** IOPS advantage (84974 vs 83870). Average latency ratio pread/uring = **1.00**. Context switches: pread=405323, uring=162075.
- **QD=4**: pread achieves **1.02x** IOPS advantage (84589 vs 86191). Average latency ratio pread/uring = **0.98**. Context switches: pread=756636, uring=177798.
- **QD=16**: uring achieves **1.10x** IOPS advantage (172958 vs 157557). Average latency ratio pread/uring = **1.09**. Context switches: pread=979825, uring=116675.
- **QD=64**: uring achieves **1.01x** IOPS advantage (211620 vs 210333). Average latency ratio pread/uring = **1.00**. Context switches: pread=779943, uring=93429.

## Random Read – 4 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 113.09 MB/s | 28.95K | 34.0 µs | 49.0 µs | 59.7 µs | 80.8 µs | 95.0 µs | 5.2 µs | 8244439 | 0 |
| 1 | uring | 113.38 MB/s | 29.03K | 34.4 µs | 49.6 µs | 59.3 µs | 81.2 µs | 95.4 µs | 3.0 µs | 4246583 | 0 |
| 4 | pread | 478.31 MB/s | 122.45K | 32.3 µs | 42.7 µs | 69.4 µs | 86.1 µs | 118.5 µs | 4.1 µs | 5794185 | 0 |
| 4 | uring | 459.46 MB/s | 117.62K | 33.9 µs | 44.0 µs | 70.8 µs | 87.5 µs | 122.5 µs | 2.0 µs | 1583358 | 0 |
| 16 | pread | 1632.52 MB/s | 417.93K | 37.3 µs | 44.4 µs | 87.4 µs | 130.3 µs | 206.6 µs | 4.3 µs | 4281210 | 0 |
| 16 | uring | 1417.85 MB/s | 362.97K | 43.8 µs | 51.2 µs | 92.4 µs | 130.0 µs | 201.2 µs | 1.9 µs | 517925 | 0 |
| 64 | pread | 2484.79 MB/s | 636.11K | 48.2 µs | 49.8 µs | 126.5 µs | 215.2 µs | 378.4 µs | 3.4 µs | 2793301 | 0 |
| 64 | uring | 2205.94 MB/s | 564.72K | 111.6 µs | 110.9 µs | 209.7 µs | 274.9 µs | 368.8 µs | 1.8 µs | 28107 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (29025 vs 28952). Average latency ratio pread/uring = **0.99**. Context switches: pread=8244439, uring=4246583.
- **QD=4**: pread achieves **1.04x** IOPS advantage (117621 vs 122448). Average latency ratio pread/uring = **0.95**. Context switches: pread=5794185, uring=1583358.
- **QD=16**: pread achieves **1.15x** IOPS advantage (362970 vs 417926). Average latency ratio pread/uring = **0.85**. Context switches: pread=4281210, uring=517925.
- **QD=64**: pread achieves **1.13x** IOPS advantage (564721 vs 636106). Average latency ratio pread/uring = **0.43**. Context switches: pread=2793301, uring=28107.

## Random Read – 8 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 141.52 MB/s | 18.12K | 54.6 µs | 76.3 µs | 84.8 µs | 94.1 µs | 128.7 µs | 6.5 µs | 4556780 | 0 |
| 1 | uring | 141.63 MB/s | 18.13K | 55.1 µs | 77.0 µs | 85.2 µs | 94.6 µs | 129.3 µs | 4.7 µs | 3552410 | 0 |
| 4 | pread | 564.42 MB/s | 72.25K | 54.9 µs | 71.5 µs | 98.1 µs | 124.0 µs | 160.7 µs | 5.7 µs | 3666939 | 0 |
| 4 | uring | 563.22 MB/s | 72.09K | 55.4 µs | 71.5 µs | 99.2 µs | 125.8 µs | 166.3 µs | 2.9 µs | 1213102 | 0 |
| 16 | pread | 1711.54 MB/s | 219.08K | 72.7 µs | 81.4 µs | 158.4 µs | 220.8 µs | 322.1 µs | 5.6 µs | 2835964 | 0 |
| 16 | uring | 1646.45 MB/s | 210.75K | 75.8 µs | 84.7 µs | 159.8 µs | 219.9 µs | 311.2 µs | 2.7 µs | 545050 | 0 |
| 64 | pread | 2780.17 MB/s | 355.86K | 177.6 µs | 146.7 µs | 516.0 µs | 816.2 µs | 1.32 ms | 6.0 µs | 1978770 | 0 |
| 64 | uring | 2664.40 MB/s | 341.04K | 187.0 µs | 156.9 µs | 482.1 µs | 767.6 µs | 1.25 ms | 2.6 µs | 152189 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (18129 vs 18115). Average latency ratio pread/uring = **0.99**. Context switches: pread=4556780, uring=3552410.
- **QD=4**: pread achieves **1.00x** IOPS advantage (72092 vs 72246). Average latency ratio pread/uring = **0.99**. Context switches: pread=3666939, uring=1213102.
- **QD=16**: pread achieves **1.04x** IOPS advantage (210745 vs 219077). Average latency ratio pread/uring = **0.96**. Context switches: pread=2835964, uring=545050.
- **QD=64**: pread achieves **1.04x** IOPS advantage (341043 vs 355861). Average latency ratio pread/uring = **0.95**. Context switches: pread=1978770, uring=152189.

## Random Read – 16 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 229.60 MB/s | 14.69K | 67.4 µs | 88.9 µs | 94.8 µs | 110.1 µs | 144.0 µs | 8.4 µs | 2480951 | 0 |
| 1 | uring | 230.61 MB/s | 14.76K | 67.7 µs | 89.4 µs | 95.1 µs | 110.6 µs | 144.8 µs | 6.4 µs | 1988420 | 0 |
| 4 | pread | 926.18 MB/s | 59.28K | 67.0 µs | 82.5 µs | 114.5 µs | 140.3 µs | 183.4 µs | 7.4 µs | 2083207 | 0 |
| 4 | uring | 926.16 MB/s | 59.27K | 67.4 µs | 82.5 µs | 115.4 µs | 142.5 µs | 201.9 µs | 4.3 µs | 703032 | 0 |
| 16 | pread | 2738.21 MB/s | 175.25K | 90.9 µs | 96.3 µs | 190.6 µs | 264.9 µs | 383.1 µs | 7.7 µs | 1669191 | 0 |
| 16 | uring | 2627.31 MB/s | 168.15K | 95.0 µs | 100.5 µs | 193.7 µs | 266.6 µs | 370.9 µs | 4.0 µs | 280233 | 0 |
| 64 | pread | 4239.07 MB/s | 271.30K | 235.2 µs | 202.6 µs | 639.7 µs | 993.0 µs | 1.56 ms | 8.3 µs | 1130832 | 0 |
| 64 | uring | 3933.21 MB/s | 251.73K | 253.3 µs | 225.4 µs | 571.3 µs | 871.9 µs | 1.38 ms | 3.9 µs | 29781 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (14759 vs 14695). Average latency ratio pread/uring = **1.00**. Context switches: pread=2480951, uring=1988420.
- **QD=4**: pread achieves **1.00x** IOPS advantage (59275 vs 59276). Average latency ratio pread/uring = **0.99**. Context switches: pread=2083207, uring=703032.
- **QD=16**: pread achieves **1.04x** IOPS advantage (168148 vs 175245). Average latency ratio pread/uring = **0.96**. Context switches: pread=1669191, uring=280233.
- **QD=64**: pread achieves **1.08x** IOPS advantage (251726 vs 271300). Average latency ratio pread/uring = **0.93**. Context switches: pread=1130832, uring=29781.

## Random Read – 32 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 354.32 MB/s | 11.34K | 87.5 µs | 106.4 µs | 112.4 µs | 125.8 µs | 181.9 µs | 11.1 µs | 1376794 | 0 |
| 1 | uring | 354.44 MB/s | 11.34K | 88.1 µs | 107.0 µs | 113.4 µs | 126.6 µs | 181.5 µs | 9.5 µs | 1274990 | 0 |
| 4 | pread | 1347.26 MB/s | 43.11K | 92.2 µs | 104.1 µs | 144.8 µs | 178.3 µs | 238.8 µs | 10.8 µs | 1244505 | 0 |
| 4 | uring | 1348.20 MB/s | 43.14K | 92.6 µs | 104.2 µs | 146.7 µs | 182.9 µs | 277.1 µs | 6.8 µs | 443322 | 0 |
| 16 | pread | 3620.92 MB/s | 115.87K | 137.6 µs | 132.3 µs | 296.2 µs | 457.4 µs | 688.2 µs | 11.6 µs | 1050208 | 0 |
| 16 | uring | 3461.34 MB/s | 110.76K | 144.3 µs | 140.9 µs | 297.2 µs | 441.2 µs | 661.9 µs | 6.6 µs | 146775 | 0 |
| 64 | pread | 4845.70 MB/s | 155.06K | 411.8 µs | 330.2 µs | 1.17 ms | 2.11 ms | 3.89 ms | 13.1 µs | 748236 | 0 |
| 64 | uring | 4574.87 MB/s | 146.40K | 436.3 µs | 368.4 µs | 1.03 ms | 1.81 ms | 3.54 ms | 6.6 µs | 28909 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (11342 vs 11338). Average latency ratio pread/uring = **0.99**. Context switches: pread=1376794, uring=1274990.
- **QD=4**: uring achieves **1.00x** IOPS advantage (43142 vs 43112). Average latency ratio pread/uring = **1.00**. Context switches: pread=1244505, uring=443322.
- **QD=16**: pread achieves **1.05x** IOPS advantage (110763 vs 115870). Average latency ratio pread/uring = **0.95**. Context switches: pread=1050208, uring=146775.
- **QD=64**: pread achieves **1.06x** IOPS advantage (146396 vs 155062). Average latency ratio pread/uring = **0.94**. Context switches: pread=748236, uring=28909.

## Summary

- **io_uring shines at QD=64** with **1.53x** IOPS over pread. High queue depths expose the batched submission advantage.
- **pread competitive at QD=16** (ratio 0.87x). At low concurrency the syscall overhead difference is minimal.

## Methodology

1. **Identical workload**: both engines receive the same pre-generated offset
   sequence (deterministic RNG, same seed), so each read touches the same
   blocks in the same order.  Each pass reads the entire file once
   (`fileSize / blockSize` I/O operations).
2. **Page cache drop**: the page cache is dropped before each run to
   ensure reads hit the storage device.
3. **pread engine**: a pool of QD goroutines issues concurrent `pread(2)`
   calls against a shared file descriptor.
4. **io_uring engine**: a single goroutine manages the submission/completion
   rings, keeping QD requests in-flight.  `IORING_OP_READ` (kernel 5.6+)
   is used instead of the older vectored `IORING_OP_READV`.
5. **Latency**: per-I/O latency is measured from submission to completion.
   For pread this is the `ReadAt` wall time; for io_uring it is from
   SQE preparation to CQE reaping.
6. **CPU accounting**: `getrusage(RUSAGE_SELF)` captures process-wide
   user + system CPU time and voluntary/involuntary context switches.
7. **Aggregation**: when multiple passes are requested, latency samples
   are pooled.

---

*Generated by [bench/disk/readcmp](https://github.com/rjl493456442/bench)*
