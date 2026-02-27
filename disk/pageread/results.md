# Disk Read Performance Benchmark

## Test Environment

| Parameter | Value |
|-----------|-------|
| Date | 2026-02-27 13:13:57 |
| OS | linux |
| Architecture | amd64 |
| Data File | `bench_pageread.bin` |
| File Size | 5120 MB |
| Queue Depths | 1, 4, 8, 16, 32 |
| Passes per Page Size | 3 |
| Cache Bypass Method | `O_DIRECT` |

> Latency values are per individual `read(2)` / `pread(2)` syscall.  
> Throughput and IOPS are averaged across all passes.

## Sequential Read Results (QD=1)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 104.79 MB/s | 26.83K | 37.2 µs | 23.4 µs | 72.8 µs | 86.7 µs | 93.7 µs | 6.4 µs | 3.42 ms |
| 8 KB | 1.97M | 444.29 MB/s | 56.87K | 17.5 µs | 11.7 µs | 30.7 µs | 88.8 µs | 154.6 µs | 6.5 µs | 3.44 ms |
| 16 KB | 983.04K | 558.26 MB/s | 35.73K | 27.9 µs | 11.9 µs | 98.1 µs | 116.6 µs | 169.3 µs | 8.3 µs | 961.2 µs |
| 32 KB | 491.52K | 641.42 MB/s | 20.53K | 48.7 µs | 13.0 µs | 129.5 µs | 190.6 µs | 378.0 µs | 11.3 µs | 2.08 ms |
| 64 KB | 245.76K | 627.72 MB/s | 10.04K | 99.5 µs | 31.2 µs | 226.6 µs | 303.4 µs | 519.2 µs | 16.5 µs | 1.34 ms |
| 128 KB | 122.88K | 595.76 MB/s | 4.77K | 209.8 µs | 235.0 µs | 341.7 µs | 455.7 µs | 667.8 µs | 27.0 µs | 2.05 ms |
| 256 KB | 61.44K | 824.71 MB/s | 3.30K | 303.1 µs | 325.2 µs | 445.5 µs | 610.0 µs | 862.0 µs | 45.9 µs | 2.13 ms |
| 512 KB | 30.72K | 1169.22 MB/s | 2.34K | 427.6 µs | 444.8 µs | 599.7 µs | 788.9 µs | 1.14 ms | 84.0 µs | 2.34 ms |
| 1 MB | 15.36K | 1677.74 MB/s | 1.68K | 596.0 µs | 593.4 µs | 745.4 µs | 1.05 ms | 1.39 ms | 221.7 µs | 1.83 ms |

### Observations

- Peak throughput **1677.74 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **16.0x higher** than with 4 KB pages.

## Sequential Read Results (QD=4)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 184.15 MB/s | 47.14K | 83.9 µs | 81.1 µs | 94.8 µs | 101.9 µs | 149.9 µs | 6.9 µs | 1.94 ms |
| 8 KB | 1.97M | 451.89 MB/s | 57.84K | 68.3 µs | 30.1 µs | 162.8 µs | 278.6 µs | 496.6 µs | 6.7 µs | 3.47 ms |
| 16 KB | 983.04K | 995.88 MB/s | 63.74K | 62.2 µs | 22.6 µs | 186.8 µs | 303.1 µs | 521.3 µs | 9.0 µs | 6.79 ms |
| 32 KB | 491.52K | 1213.29 MB/s | 38.83K | 102.4 µs | 57.0 µs | 232.5 µs | 392.8 µs | 647.0 µs | 12.2 µs | 2.12 ms |
| 64 KB | 245.76K | 1321.52 MB/s | 21.14K | 188.6 µs | 187.6 µs | 321.2 µs | 520.4 µs | 790.0 µs | 18.4 µs | 2.12 ms |
| 128 KB | 122.88K | 1570.81 MB/s | 12.57K | 317.6 µs | 304.3 µs | 454.0 µs | 731.5 µs | 1.08 ms | 35.5 µs | 2.53 ms |
| 256 KB | 61.44K | 2419.67 MB/s | 9.68K | 412.6 µs | 404.6 µs | 525.7 µs | 626.0 µs | 759.2 µs | 169.7 µs | 2.40 ms |
| 512 KB | 30.72K | 3600.23 MB/s | 7.20K | 554.7 µs | 539.6 µs | 688.8 µs | 815.5 µs | 931.9 µs | 297.6 µs | 2.25 ms |
| 1 MB | 15.36K | 4756.63 MB/s | 4.76K | 839.8 µs | 817.0 µs | 1.06 ms | 1.26 ms | 1.50 ms | 466.5 µs | 2.47 ms |

### Observations

- Peak throughput **4756.63 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **25.8x higher** than with 4 KB pages.

## Sequential Read Results (QD=8)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 246.12 MB/s | 63.01K | 126.0 µs | 138.3 µs | 165.7 µs | 203.4 µs | 266.7 µs | 6.9 µs | 2.77 ms |
| 8 KB | 1.97M | 557.77 MB/s | 71.39K | 111.1 µs | 92.7 µs | 238.5 µs | 388.7 µs | 632.1 µs | 6.7 µs | 3.62 ms |
| 16 KB | 983.04K | 1176.95 MB/s | 75.32K | 105.6 µs | 53.0 µs | 252.8 µs | 412.4 µs | 657.2 µs | 9.4 µs | 2.31 ms |
| 32 KB | 491.52K | 1476.13 MB/s | 47.24K | 168.8 µs | 151.1 µs | 307.1 µs | 517.1 µs | 797.3 µs | 19.2 µs | 2.45 ms |
| 64 KB | 245.76K | 1809.66 MB/s | 28.95K | 275.7 µs | 260.0 µs | 433.5 µs | 726.1 µs | 1.05 ms | 52.3 µs | 3.01 ms |
| 128 KB | 122.88K | 2717.82 MB/s | 21.74K | 367.3 µs | 362.0 µs | 477.0 µs | 592.2 µs | 670.1 µs | 136.0 µs | 2.26 ms |
| 256 KB | 61.44K | 3953.53 MB/s | 15.81K | 504.7 µs | 487.1 µs | 650.8 µs | 784.8 µs | 895.5 µs | 200.9 µs | 2.15 ms |
| 512 KB | 30.72K | 5141.32 MB/s | 10.28K | 776.6 µs | 749.4 µs | 1.00 ms | 1.20 ms | 1.37 ms | 352.2 µs | 2.75 ms |
| 1 MB | 15.36K | 5671.31 MB/s | 5.67K | 1.41 ms | 1.40 ms | 1.66 ms | 1.88 ms | 2.14 ms | 777.4 µs | 2.92 ms |

### Observations

- Peak throughput **5671.31 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **23.0x higher** than with 4 KB pages.

## Sequential Read Results (QD=16)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 356.38 MB/s | 91.23K | 174.3 µs | 164.1 µs | 286.2 µs | 339.0 µs | 395.9 µs | 7.7 µs | 2.13 ms |
| 8 KB | 1.97M | 703.45 MB/s | 90.04K | 176.8 µs | 165.7 µs | 322.3 µs | 514.7 µs | 756.6 µs | 7.2 µs | 4.05 ms |
| 16 KB | 983.04K | 1359.27 MB/s | 86.99K | 183.2 µs | 146.8 µs | 358.7 µs | 592.3 µs | 857.2 µs | 12.0 µs | 4.05 ms |
| 32 KB | 491.52K | 1857.02 MB/s | 59.42K | 268.6 µs | 248.8 µs | 428.8 µs | 718.8 µs | 1.04 ms | 38.6 µs | 2.31 ms |
| 64 KB | 245.76K | 3023.06 MB/s | 48.37K | 329.7 µs | 323.4 µs | 446.9 µs | 584.4 µs | 683.0 µs | 109.0 µs | 2.53 ms |
| 128 KB | 122.88K | 4261.89 MB/s | 34.10K | 468.0 µs | 455.2 µs | 612.3 µs | 763.1 µs | 881.1 µs | 168.4 µs | 2.22 ms |
| 256 KB | 61.44K | 5445.24 MB/s | 21.78K | 732.9 µs | 695.2 µs | 971.4 µs | 1.18 ms | 1.36 ms | 322.8 µs | 2.41 ms |
| 512 KB | 30.72K | 5668.45 MB/s | 11.34K | 1.41 ms | 1.39 ms | 1.67 ms | 1.90 ms | 2.11 ms | 521.1 µs | 2.89 ms |
| 1 MB | 15.36K | 5670.62 MB/s | 5.67K | 2.82 ms | 2.81 ms | 3.11 ms | 3.37 ms | 3.65 ms | 1.08 ms | 4.26 ms |

### Observations

- Peak throughput **5670.62 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **15.9x higher** than with 4 KB pages.

## Sequential Read Results (QD=32)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 556.54 MB/s | 142.47K | 222.8 µs | 205.8 µs | 391.9 µs | 540.5 µs | 625.6 µs | 15.0 µs | 1.68 ms |
| 8 KB | 1.97M | 855.83 MB/s | 109.55K | 290.9 µs | 271.7 µs | 488.9 µs | 731.4 µs | 979.3 µs | 8.4 µs | 3.98 ms |
| 16 KB | 983.04K | 1610.25 MB/s | 103.06K | 309.6 µs | 292.5 µs | 505.2 µs | 764.4 µs | 1.06 ms | 30.1 µs | 4.32 ms |
| 32 KB | 491.52K | 3252.59 MB/s | 104.08K | 306.5 µs | 297.6 µs | 457.6 µs | 634.0 µs | 751.2 µs | 93.5 µs | 2.36 ms |
| 64 KB | 245.76K | 4573.49 MB/s | 73.18K | 436.4 µs | 428.2 µs | 592.6 µs | 778.4 µs | 945.8 µs | 128.6 µs | 2.44 ms |
| 128 KB | 122.88K | 5651.59 MB/s | 45.21K | 706.5 µs | 669.6 µs | 952.9 µs | 1.21 ms | 1.39 ms | 197.9 µs | 1.89 ms |
| 256 KB | 61.44K | 5669.06 MB/s | 22.68K | 1.41 ms | 1.37 ms | 1.67 ms | 1.92 ms | 2.16 ms | 385.6 µs | 3.38 ms |
| 512 KB | 30.72K | 5666.54 MB/s | 11.33K | 2.82 ms | 2.80 ms | 3.11 ms | 3.39 ms | 3.67 ms | 665.3 µs | 4.74 ms |
| 1 MB | 15.36K | 5624.58 MB/s | 5.62K | 5.67 ms | 5.66 ms | 6.10 ms | 6.53 ms | 7.26 ms | 1.11 ms | 8.06 ms |

### Observations

- Peak throughput **5669.06 MB/s** achieved with **256 KB** page size.
- Throughput with 1 MB pages is **10.1x higher** than with 4 KB pages.

## Random Read Results (QD=1)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 46.58 MB/s | 11.92K | 83.8 µs | 80.3 µs | 93.8 µs | 95.9 µs | 98.8 µs | 16.9 µs | 2.20 ms |
| 8 KB | 1.97M | 89.79 MB/s | 11.49K | 87.0 µs | 82.0 µs | 96.3 µs | 155.7 µs | 159.3 µs | 22.6 µs | 3.44 ms |
| 16 KB | 983.04K | 150.19 MB/s | 9.61K | 104.0 µs | 99.0 µs | 113.1 µs | 166.2 µs | 169.1 µs | 75.2 µs | 2.19 ms |
| 32 KB | 491.52K | 231.07 MB/s | 7.39K | 135.2 µs | 121.4 µs | 186.3 µs | 191.9 µs | 238.1 µs | 82.2 µs | 2.19 ms |
| 64 KB | 245.76K | 318.90 MB/s | 5.10K | 195.9 µs | 193.4 µs | 264.0 µs | 291.5 µs | 337.1 µs | 98.4 µs | 2.16 ms |
| 128 KB | 122.88K | 466.65 MB/s | 3.73K | 267.8 µs | 269.6 µs | 345.4 µs | 401.6 µs | 490.1 µs | 121.8 µs | 2.41 ms |
| 256 KB | 61.44K | 699.52 MB/s | 2.80K | 357.3 µs | 351.4 µs | 444.2 µs | 567.6 µs | 630.3 µs | 164.9 µs | 2.04 ms |
| 512 KB | 30.72K | 1071.29 MB/s | 2.14K | 466.7 µs | 455.5 µs | 586.3 µs | 645.5 µs | 682.0 µs | 254.1 µs | 2.38 ms |
| 1 MB | 15.36K | 1626.74 MB/s | 1.63K | 614.7 µs | 604.8 µs | 714.8 µs | 953.3 µs | 1.29 ms | 314.1 µs | 1.95 ms |

### Observations

- Peak throughput **1626.74 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **34.9x higher** than with 4 KB pages.

## Random Read Results (QD=4)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 184.62 MB/s | 47.26K | 84.0 µs | 80.7 µs | 94.1 µs | 138.0 µs | 164.0 µs | 16.9 µs | 2.20 ms |
| 8 KB | 1.97M | 351.76 MB/s | 45.03K | 88.2 µs | 84.2 µs | 101.0 µs | 155.7 µs | 192.7 µs | 21.8 µs | 3.45 ms |
| 16 KB | 983.04K | 619.29 MB/s | 39.63K | 100.3 µs | 95.8 µs | 116.2 µs | 173.6 µs | 221.9 µs | 71.5 µs | 2.06 ms |
| 32 KB | 491.52K | 889.56 MB/s | 28.47K | 139.9 µs | 123.9 µs | 192.1 µs | 257.9 µs | 327.7 µs | 81.1 µs | 2.30 ms |
| 64 KB | 245.76K | 1186.43 MB/s | 18.98K | 210.1 µs | 210.8 µs | 283.3 µs | 389.1 µs | 489.5 µs | 99.9 µs | 2.06 ms |
| 128 KB | 122.88K | 1633.31 MB/s | 13.07K | 305.4 µs | 295.5 µs | 409.8 µs | 557.4 µs | 699.5 µs | 128.3 µs | 2.26 ms |
| 256 KB | 61.44K | 2247.14 MB/s | 8.99K | 444.3 µs | 426.6 µs | 593.6 µs | 802.2 µs | 1.02 ms | 172.4 µs | 2.48 ms |
| 512 KB | 30.72K | 3080.63 MB/s | 6.16K | 648.5 µs | 622.4 µs | 860.5 µs | 1.17 ms | 1.47 ms | 259.7 µs | 2.45 ms |
| 1 MB | 15.36K | 4098.90 MB/s | 4.10K | 975.0 µs | 933.1 µs | 1.31 ms | 1.71 ms | 2.09 ms | 456.2 µs | 2.84 ms |

### Observations

- Peak throughput **4098.90 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **22.2x higher** than with 4 KB pages.

## Random Read Results (QD=8)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 364.09 MB/s | 93.21K | 85.0 µs | 80.9 µs | 105.4 µs | 153.6 µs | 199.1 µs | 16.3 µs | 2.19 ms |
| 8 KB | 1.97M | 676.93 MB/s | 86.65K | 91.6 µs | 86.0 µs | 119.7 µs | 172.6 µs | 224.2 µs | 22.0 µs | 3.50 ms |
| 16 KB | 983.04K | 1176.04 MB/s | 75.27K | 105.7 µs | 97.5 µs | 145.7 µs | 205.9 µs | 263.4 µs | 74.0 µs | 2.18 ms |
| 32 KB | 491.52K | 1629.84 MB/s | 52.16K | 152.8 µs | 133.3 µs | 218.6 µs | 312.4 µs | 395.8 µs | 80.9 µs | 2.37 ms |
| 64 KB | 245.76K | 2083.45 MB/s | 33.34K | 239.4 µs | 230.9 µs | 338.2 µs | 481.1 µs | 610.9 µs | 101.9 µs | 2.27 ms |
| 128 KB | 122.88K | 2709.03 MB/s | 21.67K | 368.4 µs | 349.6 µs | 518.0 µs | 719.6 µs | 931.4 µs | 131.2 µs | 2.14 ms |
| 256 KB | 61.44K | 3453.73 MB/s | 13.81K | 578.0 µs | 546.2 µs | 805.3 µs | 1.16 ms | 1.54 ms | 185.4 µs | 2.27 ms |
| 512 KB | 30.72K | 4191.08 MB/s | 8.38K | 952.9 µs | 886.8 µs | 1.39 ms | 1.91 ms | 2.29 ms | 330.3 µs | 2.87 ms |
| 1 MB | 15.36K | 4753.79 MB/s | 4.75K | 1.68 ms | 1.62 ms | 2.32 ms | 3.01 ms | 3.51 ms | 516.8 µs | 4.02 ms |

### Observations

- Peak throughput **4753.79 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **13.1x higher** than with 4 KB pages.

## Random Read Results (QD=16)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 660.05 MB/s | 168.97K | 93.6 µs | 84.3 µs | 132.6 µs | 197.2 µs | 260.9 µs | 16.5 µs | 2.25 ms |
| 8 KB | 1.97M | 1202.18 MB/s | 153.88K | 103.1 µs | 91.8 µs | 149.0 µs | 220.6 µs | 292.2 µs | 22.1 µs | 2.20 ms |
| 16 KB | 983.04K | 2019.40 MB/s | 129.24K | 123.1 µs | 108.0 µs | 179.9 µs | 266.2 µs | 346.8 µs | 68.2 µs | 4.78 ms |
| 32 KB | 491.52K | 2653.77 MB/s | 84.92K | 187.7 µs | 171.5 µs | 283.6 µs | 409.6 µs | 529.6 µs | 82.9 µs | 2.18 ms |
| 64 KB | 245.76K | 3210.49 MB/s | 51.37K | 310.4 µs | 291.8 µs | 459.3 µs | 659.3 µs | 847.2 µs | 104.4 µs | 1.84 ms |
| 128 KB | 122.88K | 3900.59 MB/s | 31.20K | 511.6 µs | 479.6 µs | 744.9 µs | 1.11 ms | 1.48 ms | 140.2 µs | 2.31 ms |
| 256 KB | 61.44K | 4460.18 MB/s | 17.84K | 895.2 µs | 799.3 µs | 1.42 ms | 2.15 ms | 2.65 ms | 230.9 µs | 3.96 ms |
| 512 KB | 30.72K | 4626.65 MB/s | 9.25K | 1.73 ms | 1.61 ms | 2.47 ms | 3.28 ms | 4.03 ms | 548.3 µs | 4.98 ms |
| 1 MB | 15.36K | 4772.16 MB/s | 4.77K | 3.35 ms | 3.28 ms | 4.04 ms | 4.79 ms | 5.23 ms | 916.8 µs | 5.78 ms |

### Observations

- Peak throughput **4772.16 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **7.2x higher** than with 4 KB pages.

## Random Read Results (QD=32)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 1065.70 MB/s | 272.82K | 115.1 µs | 96.0 µs | 179.7 µs | 288.3 µs | 394.4 µs | 16.8 µs | 3.58 ms |
| 8 KB | 1.97M | 1892.32 MB/s | 242.22K | 130.7 µs | 110.9 µs | 204.3 µs | 320.3 µs | 433.2 µs | 23.3 µs | 3.58 ms |
| 16 KB | 983.04K | 3031.14 MB/s | 193.99K | 163.9 µs | 145.9 µs | 254.5 µs | 379.7 µs | 494.3 µs | 76.7 µs | 1.99 ms |
| 32 KB | 491.52K | 3760.37 MB/s | 120.33K | 265.1 µs | 246.0 µs | 409.7 µs | 599.6 µs | 801.1 µs | 85.2 µs | 2.28 ms |
| 64 KB | 245.76K | 4289.10 MB/s | 68.63K | 465.1 µs | 431.1 µs | 706.2 µs | 1.12 ms | 1.60 ms | 118.3 µs | 2.70 ms |
| 128 KB | 122.88K | 4718.08 MB/s | 37.74K | 846.1 µs | 718.4 µs | 1.45 ms | 2.53 ms | 3.35 ms | 160.0 µs | 4.37 ms |
| 256 KB | 61.44K | 4662.36 MB/s | 18.65K | 1.71 ms | 1.57 ms | 2.44 ms | 3.36 ms | 4.22 ms | 440.9 µs | 5.05 ms |
| 512 KB | 30.72K | 4630.81 MB/s | 9.26K | 3.45 ms | 3.34 ms | 4.23 ms | 5.15 ms | 5.77 ms | 541.1 µs | 6.56 ms |
| 1 MB | 15.36K | 4737.31 MB/s | 4.74K | 6.73 ms | 6.68 ms | 7.52 ms | 8.29 ms | 8.98 ms | 1.11 ms | 9.79 ms |

### Observations

- Peak throughput **4737.31 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **4.4x higher** than with 4 KB pages.

## Methodology

1. **Dataset**: a `5120 MB` file filled with pseudo-random bytes (PCG-64 RNG,
   deterministic seed) to defeat filesystem-level compression.
2. **Cache bypass**: `O_DIRECT` is applied before each pass to ensure reads come
   from the storage device, not the OS page cache.  Additionally, `echo 3 > /proc/sys/vm/drop_caches` is
   invoked prior to each pass (best-effort; a warning is printed if it fails).
3. **Sequential mode**: reads the file from start to finish with `read(2)`.
4. **Random mode**: issues the same number of reads (`fileSize / pageSize`)
   at uniformly random page-aligned offsets via `pread(2)` (`ReadAt`).
   Offsets are pre-generated before the timed loop to exclude RNG overhead.
5. **Aggregation**: when multiple passes are requested, latency samples are
   pooled and throughput / IOPS are averaged across all passes.
6. **Queue depth**: when QD > 1, a pool of QD goroutines issues concurrent
   `pread(2)` calls on the same file descriptor, keeping the device's NCQ /
   NVMe submission queue busy. QD = 1 uses the legacy single-read path.

---

*Generated by [bench/disk/pageread](https://github.com/rjl493456442/bench)*
