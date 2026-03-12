# io_uring vs pread – Read Performance Comparison

## Test Environment

| Parameter | Value |
|-----------|-------|
| Date | 2026-03-12 14:59:12 |
| OS | linux |
| Architecture | amd64 |
| Kernel | 6.17.0-14-generic |
| Data File | `bench_readcmp.bin` |
| File Size | 5120 MB |
| I/O Mode | O_DIRECT |
| Workload | entire file per pass |
| Passes | 3 |
| io_uring features | SINGLE_MMAP, NODROP, SUBMIT_STABLE |

## Sequential Read – 4 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 330.97 MB/s | 84.73K | 11.2 µs | 11.1 µs | 11.5 µs | 15.0 µs | 24.2 µs | 5.5 µs | 11913116 | 0 |
| 1 | uring | 330.62 MB/s | 84.64K | 11.7 µs | 11.6 µs | 12.0 µs | 15.3 µs | 25.8 µs | 1.7 µs | 4239536 | 0 |
| 4 | pread | 1608.88 MB/s | 411.87K | 9.4 µs | 9.2 µs | 10.6 µs | 11.7 µs | 20.9 µs | 3.4 µs | 6693295 | 0 |
| 4 | uring | 1180.69 MB/s | 302.26K | 13.1 µs | 12.9 µs | 13.6 µs | 16.2 µs | 72.0 µs | 0.8 µs | 1040956 | 0 |
| 16 | pread | 2192.37 MB/s | 561.25K | 28.2 µs | 28.4 µs | 30.9 µs | 34.3 µs | 45.2 µs | 3.5 µs | 5314617 | 0 |
| 16 | uring | 2165.82 MB/s | 554.45K | 28.3 µs | 28.1 µs | 28.8 µs | 31.9 µs | 88.2 µs | 0.6 µs | 275783 | 0 |
| 64 | pread | 2045.19 MB/s | 523.57K | 60.8 µs | 56.7 µs | 122.6 µs | 151.1 µs | 188.2 µs | 2.5 µs | 3950580 | 0 |
| 64 | uring | 2794.50 MB/s | 715.39K | 88.5 µs | 87.0 µs | 106.4 µs | 108.1 µs | 177.1 µs | 0.5 µs | 141930 | 0 |

### Observations

- **QD=1**: pread achieves **1.00x** IOPS advantage (84640 vs 84728). Average latency ratio pread/uring = **0.96**. Context switches: pread=11913116, uring=4239536.
- **QD=4**: pread achieves **1.36x** IOPS advantage (302258 vs 411873). Average latency ratio pread/uring = **0.72**. Context switches: pread=6693295, uring=1040956.
- **QD=16**: pread achieves **1.01x** IOPS advantage (554451 vs 561247). Average latency ratio pread/uring = **1.00**. Context switches: pread=5314617, uring=275783.
- **QD=64**: uring achieves **1.37x** IOPS advantage (715392 vs 523568). Average latency ratio pread/uring = **0.69**. Context switches: pread=3950580, uring=141930.

## Sequential Read – 8 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 660.92 MB/s | 84.60K | 11.3 µs | 11.1 µs | 11.6 µs | 15.2 µs | 24.9 µs | 5.5 µs | 5957816 | 0 |
| 1 | uring | 660.63 MB/s | 84.56K | 11.7 µs | 11.6 µs | 11.9 µs | 15.3 µs | 32.0 µs | 1.8 µs | 2115719 | 0 |
| 4 | pread | 2485.28 MB/s | 318.12K | 12.2 µs | 12.2 µs | 13.0 µs | 13.6 µs | 23.2 µs | 4.0 µs | 4031226 | 0 |
| 4 | uring | 1795.90 MB/s | 229.87K | 17.3 µs | 17.1 µs | 17.7 µs | 20.4 µs | 80.7 µs | 0.9 µs | 534022 | 0 |
| 16 | pread | 2699.51 MB/s | 345.54K | 46.0 µs | 30.8 µs | 108.8 µs | 148.8 µs | 211.0 µs | 4.1 µs | 3119985 | 0 |
| 16 | uring | 2763.09 MB/s | 353.68K | 44.7 µs | 44.4 µs | 45.6 µs | 49.6 µs | 98.0 µs | 0.7 µs | 146725 | 0 |
| 64 | pread | 3535.02 MB/s | 452.48K | 70.4 µs | 68.1 µs | 98.0 µs | 122.0 µs | 160.1 µs | 2.4 µs | 1970291 | 0 |
| 64 | uring | 4723.12 MB/s | 604.56K | 105.3 µs | 104.9 µs | 112.4 µs | 117.2 µs | 149.4 µs | 0.7 µs | 116311 | 0 |

### Observations

- **QD=1**: pread achieves **1.00x** IOPS advantage (84561 vs 84598). Average latency ratio pread/uring = **0.96**. Context switches: pread=5957816, uring=2115719.
- **QD=4**: pread achieves **1.38x** IOPS advantage (229875 vs 318116). Average latency ratio pread/uring = **0.71**. Context switches: pread=4031226, uring=534022.
- **QD=16**: uring achieves **1.02x** IOPS advantage (353676 vs 345537). Average latency ratio pread/uring = **1.03**. Context switches: pread=3119985, uring=146725.
- **QD=64**: uring achieves **1.34x** IOPS advantage (604559 vs 452482). Average latency ratio pread/uring = **0.67**. Context switches: pread=1970291, uring=116311.

## Sequential Read – 16 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 214.85 MB/s | 13.75K | 71.8 µs | 71.9 µs | 72.8 µs | 75.9 µs | 79.3 µs | 7.7 µs | 3206008 | 0 |
| 1 | uring | 214.78 MB/s | 13.75K | 72.7 µs | 72.7 µs | 73.5 µs | 76.7 µs | 83.2 µs | 2.8 µs | 1563923 | 0 |
| 4 | pread | 883.46 MB/s | 56.54K | 70.2 µs | 68.6 µs | 86.1 µs | 92.6 µs | 98.5 µs | 5.5 µs | 2555309 | 0 |
| 4 | uring | 733.18 MB/s | 46.92K | 85.1 µs | 83.9 µs | 91.7 µs | 92.3 µs | 96.4 µs | 2.0 µs | 946740 | 0 |
| 16 | pread | 3588.77 MB/s | 229.68K | 69.3 µs | 72.0 µs | 81.4 µs | 96.2 µs | 103.5 µs | 5.1 µs | 1997432 | 0 |
| 16 | uring | 2730.87 MB/s | 174.78K | 91.3 µs | 91.0 µs | 93.0 µs | 94.4 µs | 110.8 µs | 0.9 µs | 157053 | 0 |
| 64 | pread | 5680.11 MB/s | 363.53K | 87.8 µs | 84.4 µs | 123.1 µs | 161.8 µs | 221.5 µs | 2.4 µs | 983815 | 0 |
| 64 | uring | 6541.65 MB/s | 418.67K | 152.6 µs | 150.8 µs | 189.6 µs | 204.6 µs | 348.8 µs | 0.9 µs | 114837 | 0 |

### Observations

- **QD=1**: pread achieves **1.00x** IOPS advantage (13746 vs 13750). Average latency ratio pread/uring = **0.99**. Context switches: pread=3206008, uring=1563923.
- **QD=4**: pread achieves **1.20x** IOPS advantage (46924 vs 56541). Average latency ratio pread/uring = **0.83**. Context switches: pread=2555309, uring=946740.
- **QD=16**: pread achieves **1.31x** IOPS advantage (174776 vs 229682). Average latency ratio pread/uring = **0.76**. Context switches: pread=1997432, uring=157053.
- **QD=64**: uring achieves **1.15x** IOPS advantage (418665 vs 363527). Average latency ratio pread/uring = **0.58**. Context switches: pread=983815, uring=114837.

## Sequential Read – 32 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 388.49 MB/s | 12.43K | 79.7 µs | 79.8 µs | 80.7 µs | 86.4 µs | 88.1 µs | 7.1 µs | 1594511 | 0 |
| 1 | uring | 388.22 MB/s | 12.42K | 80.4 µs | 80.5 µs | 81.5 µs | 87.2 µs | 93.8 µs | 5.6 µs | 1567749 | 0 |
| 4 | pread | 1647.92 MB/s | 52.73K | 75.3 µs | 74.4 µs | 80.5 µs | 82.5 µs | 89.1 µs | 6.0 µs | 1368729 | 0 |
| 4 | uring | 1341.29 MB/s | 42.92K | 93.1 µs | 93.1 µs | 94.1 µs | 95.0 µs | 109.1 µs | 2.3 µs | 475194 | 0 |
| 16 | pread | 5543.90 MB/s | 177.40K | 89.8 µs | 91.5 µs | 98.5 µs | 102.1 µs | 147.2 µs | 5.8 µs | 1121642 | 0 |
| 16 | uring | 5136.33 MB/s | 164.36K | 97.2 µs | 95.8 µs | 106.6 µs | 113.9 µs | 148.1 µs | 1.4 µs | 140321 | 0 |
| 64 | pread | 6525.62 MB/s | 208.82K | 197.4 µs | 159.8 µs | 320.0 µs | 572.5 µs | 1.01 ms | 3.9 µs | 581779 | 0 |
| 64 | uring | 6671.18 MB/s | 213.48K | 299.6 µs | 239.8 µs | 654.7 µs | 777.2 µs | 923.5 µs | 1.4 µs | 117735 | 0 |

### Observations

- **QD=1**: pread achieves **1.00x** IOPS advantage (12423 vs 12432). Average latency ratio pread/uring = **0.99**. Context switches: pread=1594511, uring=1567749.
- **QD=4**: pread achieves **1.23x** IOPS advantage (42921 vs 52733). Average latency ratio pread/uring = **0.81**. Context switches: pread=1368729, uring=475194.
- **QD=16**: pread achieves **1.08x** IOPS advantage (164362 vs 177405). Average latency ratio pread/uring = **0.92**. Context switches: pread=1121642, uring=140321.
- **QD=64**: uring achieves **1.02x** IOPS advantage (213478 vs 208820). Average latency ratio pread/uring = **0.66**. Context switches: pread=581779, uring=117735.

## Random Read – 4 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 77.04 MB/s | 19.72K | 50.0 µs | 49.9 µs | 52.0 µs | 52.5 µs | 60.3 µs | 6.4 µs | 12387689 | 0 |
| 1 | uring | 77.04 MB/s | 19.72K | 50.6 µs | 50.6 µs | 52.6 µs | 53.1 µs | 56.3 µs | 2.7 µs | 6541642 | 0 |
| 4 | pread | 341.47 MB/s | 87.42K | 45.3 µs | 44.2 µs | 53.7 µs | 67.7 µs | 80.3 µs | 4.2 µs | 8064398 | 0 |
| 4 | uring | 331.45 MB/s | 84.85K | 47.0 µs | 45.3 µs | 55.0 µs | 69.6 µs | 82.7 µs | 1.3 µs | 2217513 | 0 |
| 16 | pread | 1282.16 MB/s | 328.23K | 48.4 µs | 45.8 µs | 67.5 µs | 81.6 µs | 104.0 µs | 4.2 µs | 6331403 | 0 |
| 16 | uring | 1231.12 MB/s | 315.17K | 50.6 µs | 48.3 µs | 70.2 µs | 85.0 µs | 107.0 µs | 1.2 µs | 1178148 | 0 |
| 64 | pread | 2088.09 MB/s | 534.55K | 59.5 µs | 55.8 µs | 87.5 µs | 110.7 µs | 145.1 µs | 2.5 µs | 3952516 | 0 |
| 64 | uring | 3201.18 MB/s | 819.50K | 77.7 µs | 67.1 µs | 139.1 µs | 190.3 µs | 267.5 µs | 1.2 µs | 108752 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (19723 vs 19721). Average latency ratio pread/uring = **0.99**. Context switches: pread=12387689, uring=6541642.
- **QD=4**: pread achieves **1.03x** IOPS advantage (84852 vs 87416). Average latency ratio pread/uring = **0.96**. Context switches: pread=8064398, uring=2217513.
- **QD=16**: pread achieves **1.04x** IOPS advantage (315167 vs 328234). Average latency ratio pread/uring = **0.96**. Context switches: pread=6331403, uring=1178148.
- **QD=64**: uring achieves **1.53x** IOPS advantage (819501 vs 534550). Average latency ratio pread/uring = **0.77**. Context switches: pread=3952516, uring=108752.

## Random Read – 8 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 100.20 MB/s | 12.83K | 77.2 µs | 77.2 µs | 79.1 µs | 80.3 µs | 88.2 µs | 6.9 µs | 6374240 | 0 |
| 1 | uring | 100.32 MB/s | 12.84K | 77.8 µs | 77.9 µs | 79.8 µs | 81.2 µs | 88.5 µs | 5.6 µs | 7206052 | 0 |
| 4 | pread | 415.67 MB/s | 53.21K | 74.7 µs | 71.4 µs | 95.3 µs | 118.4 µs | 145.4 µs | 5.1 µs | 5013959 | 0 |
| 4 | uring | 419.90 MB/s | 53.75K | 74.3 µs | 71.0 µs | 95.5 µs | 118.7 µs | 147.2 µs | 1.8 µs | 1793140 | 0 |
| 16 | pread | 1332.32 MB/s | 170.54K | 93.4 µs | 83.7 µs | 144.5 µs | 190.3 µs | 257.5 µs | 5.0 µs | 3989772 | 0 |
| 16 | uring | 1329.66 MB/s | 170.20K | 93.9 µs | 84.4 µs | 145.5 µs | 191.3 µs | 258.1 µs | 1.5 µs | 1085638 | 0 |
| 64 | pread | 2117.82 MB/s | 271.08K | 151.9 µs | 122.0 µs | 328.5 µs | 546.7 µs | 905.4 µs | 3.4 µs | 2208761 | 0 |
| 64 | uring | 2432.98 MB/s | 311.42K | 205.4 µs | 164.4 µs | 465.8 µs | 710.3 µs | 1.07 ms | 1.5 µs | 848591 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (12840 vs 12825). Average latency ratio pread/uring = **0.99**. Context switches: pread=6374240, uring=7206052.
- **QD=4**: uring achieves **1.01x** IOPS advantage (53747 vs 53206). Average latency ratio pread/uring = **1.00**. Context switches: pread=5013959, uring=1793140.
- **QD=16**: pread achieves **1.00x** IOPS advantage (170196 vs 170537). Average latency ratio pread/uring = **0.99**. Context switches: pread=3989772, uring=1085638.
- **QD=64**: uring achieves **1.15x** IOPS advantage (311422 vs 271081). Average latency ratio pread/uring = **0.74**. Context switches: pread=2208761, uring=848591.

## Random Read – 16 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 173.41 MB/s | 11.10K | 89.3 µs | 89.4 µs | 91.5 µs | 91.9 µs | 99.5 µs | 7.2 µs | 3201483 | 0 |
| 1 | uring | 173.83 MB/s | 11.12K | 89.8 µs | 90.1 µs | 92.1 µs | 92.5 µs | 97.3 µs | 6.3 µs | 3904370 | 0 |
| 4 | pread | 739.48 MB/s | 47.33K | 84.0 µs | 81.2 µs | 107.3 µs | 132.0 µs | 163.1 µs | 5.7 µs | 2683444 | 0 |
| 4 | uring | 748.05 MB/s | 47.88K | 83.5 µs | 80.3 µs | 107.3 µs | 132.3 µs | 163.4 µs | 2.2 µs | 1000544 | 0 |
| 16 | pread | 2333.01 MB/s | 149.31K | 106.7 µs | 95.9 µs | 167.6 µs | 223.4 µs | 308.4 µs | 5.9 µs | 2209319 | 0 |
| 16 | uring | 2355.32 MB/s | 150.74K | 106.1 µs | 95.1 µs | 168.2 µs | 225.0 µs | 309.9 µs | 1.8 µs | 714707 | 0 |
| 64 | pread | 3520.74 MB/s | 225.33K | 172.9 µs | 137.3 µs | 376.4 µs | 628.6 µs | 1.07 ms | 3.5 µs | 1098637 | 0 |
| 64 | uring | 4047.38 MB/s | 259.03K | 246.9 µs | 192.8 µs | 580.4 µs | 889.0 µs | 1.34 ms | 1.7 µs | 538141 | 0 |

### Observations

- **QD=1**: uring achieves **1.00x** IOPS advantage (11125 vs 11098). Average latency ratio pread/uring = **0.99**. Context switches: pread=3201483, uring=3904370.
- **QD=4**: uring achieves **1.01x** IOPS advantage (47875 vs 47327). Average latency ratio pread/uring = **1.01**. Context switches: pread=2683444, uring=1000544.
- **QD=16**: uring achieves **1.01x** IOPS advantage (150741 vs 149313). Average latency ratio pread/uring = **1.01**. Context switches: pread=2209319, uring=714707.
- **QD=64**: uring achieves **1.15x** IOPS advantage (259032 vs 225328). Average latency ratio pread/uring = **0.70**. Context switches: pread=1098637, uring=538141.

## Random Read – 32 KB blocks

| QD | Engine | Throughput | IOPS | Avg Lat | P50 | P95 | P99 | P99.9 | CPU/op | CtxSw | Errors |
|---:|-------:|-----------:|-----:|--------:|----:|----:|----:|------:|-------:|------:|-------:|
| 1 | pread | 296.93 MB/s | 9.50K | 104.5 µs | 104.5 µs | 105.9 µs | 108.8 µs | 117.4 µs | 7.3 µs | 1614434 | 0 |
| 1 | uring | 296.79 MB/s | 9.50K | 105.2 µs | 105.2 µs | 106.6 µs | 109.5 µs | 117.3 µs | 7.2 µs | 1953934 | 0 |
| 4 | pread | 1174.58 MB/s | 37.59K | 105.8 µs | 100.2 µs | 137.2 µs | 168.1 µs | 213.2 µs | 6.4 µs | 1431315 | 0 |
| 4 | uring | 1187.91 MB/s | 38.01K | 105.1 µs | 99.0 µs | 137.0 µs | 168.0 µs | 212.7 µs | 2.6 µs | 575290 | 0 |
| 16 | pread | 3344.09 MB/s | 107.01K | 148.9 µs | 129.5 µs | 269.5 µs | 402.5 µs | 569.3 µs | 7.1 µs | 1193188 | 0 |
| 16 | uring | 3358.20 MB/s | 107.46K | 148.8 µs | 128.9 µs | 272.4 µs | 408.3 µs | 586.8 µs | 2.2 µs | 392331 | 0 |
| 64 | pread | 4795.44 MB/s | 153.45K | 412.1 µs | 258.9 µs | 1.21 ms | 2.13 ms | 3.61 ms | 6.8 µs | 859431 | 0 |
| 64 | uring | 4805.83 MB/s | 153.79K | 415.9 µs | 261.3 µs | 1.23 ms | 2.14 ms | 3.57 ms | 2.1 µs | 330475 | 0 |

### Observations

- **QD=1**: pread achieves **1.00x** IOPS advantage (9497 vs 9502). Average latency ratio pread/uring = **0.99**. Context switches: pread=1614434, uring=1953934.
- **QD=4**: uring achieves **1.01x** IOPS advantage (38013 vs 37587). Average latency ratio pread/uring = **1.01**. Context switches: pread=1431315, uring=575290.
- **QD=16**: uring achieves **1.00x** IOPS advantage (107462 vs 107011). Average latency ratio pread/uring = **1.00**. Context switches: pread=1193188, uring=392331.
- **QD=64**: uring achieves **1.00x** IOPS advantage (153787 vs 153454). Average latency ratio pread/uring = **0.99**. Context switches: pread=859431, uring=330475.

## Summary

- **io_uring shines at QD=64** with **1.53x** IOPS over pread. High queue depths expose the batched submission advantage.
- **pread competitive at QD=4** (ratio 0.72x). At low concurrency the syscall overhead difference is minimal.

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
