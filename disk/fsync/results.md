# Disk Fsync Performance Benchmark

## Test Environment

| Parameter | Value |
|-----------|-------|
| Date | 2026-03-12 11:20:58 |
| OS | linux |
| Architecture | amd64 |
| Directory | `/home/gary/gary/bench-run` |
| Iterations per Size | 100 |
| Queue Depths | 1, 2, 4, 8, 16, 32, 64 |
| Passes | 5 |
| Sync Method | `fdatasync` |

> Latency values are per individual `fdatasync` call.  
> Throughput = dirty_size x fsync_count / wall_time (includes write overhead).

## Fsync Results (QD=1)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 500.00 | 2.11 MB/s | 539.76 | 1.85 ms | 1.62 ms | 1.66 ms | 5.52 ms | 5.65 ms | 1.53 ms | 5.65 ms |
| 8 KB | 500.00 | 4.76 MB/s | 609.92 | 1.64 ms | 1.63 ms | 1.65 ms | 1.67 ms | 5.56 ms | 1.58 ms | 5.56 ms |
| 16 KB | 500.00 | 9.62 MB/s | 615.72 | 1.62 ms | 1.62 ms | 1.64 ms | 1.68 ms | 2.33 ms | 1.58 ms | 2.33 ms |
| 32 KB | 500.00 | 19.97 MB/s | 639.18 | 1.56 ms | 1.56 ms | 1.58 ms | 1.59 ms | 5.48 ms | 1.52 ms | 5.48 ms |
| 64 KB | 500.00 | 40.36 MB/s | 645.76 | 1.55 ms | 1.54 ms | 1.56 ms | 1.59 ms | 2.01 ms | 1.52 ms | 2.01 ms |
| 128 KB | 500.00 | 78.63 MB/s | 629.08 | 1.59 ms | 1.58 ms | 1.60 ms | 1.63 ms | 5.52 ms | 1.53 ms | 5.52 ms |
| 256 KB | 500.00 | 156.39 MB/s | 625.55 | 1.59 ms | 1.59 ms | 1.61 ms | 1.64 ms | 2.30 ms | 1.56 ms | 2.30 ms |
| 512 KB | 500.00 | 299.94 MB/s | 599.88 | 1.66 ms | 1.65 ms | 1.67 ms | 1.69 ms | 5.59 ms | 1.60 ms | 5.59 ms |
| 1 MB | 500.00 | 559.60 MB/s | 559.60 | 1.76 ms | 1.75 ms | 1.77 ms | 1.80 ms | 6.14 ms | 1.71 ms | 6.14 ms |
| 2 MB | 500.00 | 1006.37 MB/s | 503.18 | 1.95 ms | 1.95 ms | 1.97 ms | 1.99 ms | 5.89 ms | 1.91 ms | 5.89 ms |
| 4 MB | 500.00 | 1650.78 MB/s | 412.69 | 2.36 ms | 2.33 ms | 2.36 ms | 2.40 ms | 6.83 ms | 2.30 ms | 6.83 ms |
| 8 MB | 500.00 | 2407.47 MB/s | 300.93 | 3.19 ms | 3.14 ms | 3.17 ms | 3.26 ms | 7.71 ms | 3.07 ms | 7.71 ms |
| 16 MB | 500.00 | 2966.28 MB/s | 185.39 | 4.82 ms | 4.74 ms | 4.78 ms | 9.09 ms | 9.46 ms | 4.69 ms | 9.46 ms |
| 32 MB | 500.00 | 3197.05 MB/s | 99.91 | 8.06 ms | 7.89 ms | 7.96 ms | 12.31 ms | 12.41 ms | 7.82 ms | 12.41 ms |
| 64 MB | 500.00 | 3431.78 MB/s | 53.62 | 14.61 ms | 14.25 ms | 14.57 ms | 18.88 ms | 19.30 ms | 14.11 ms | 19.30 ms |
| 128 MB | 500.00 | 3466.73 MB/s | 27.08 | 28.34 ms | 27.62 ms | 31.89 ms | 35.33 ms | 41.13 ms | 26.92 ms | 41.13 ms |

### Observations

- Peak effective throughput **3466.73 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **1.55 ms** with **64 KB** dirty size.
- Fsync latency with 128 MB dirty data is **15.3x** the latency with 4 KB.

## Fsync Results (QD=2)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 500.00 | 2.77 MB/s | 709.99 | 2.81 ms | 3.07 ms | 3.19 ms | 3.22 ms | 3.23 ms | 1.61 ms | 3.23 ms |
| 8 KB | 500.00 | 5.34 MB/s | 684.02 | 2.92 ms | 3.11 ms | 3.19 ms | 3.29 ms | 3.59 ms | 1.64 ms | 3.59 ms |
| 16 KB | 500.00 | 10.18 MB/s | 651.74 | 3.06 ms | 3.05 ms | 3.28 ms | 3.36 ms | 7.11 ms | 1.55 ms | 7.11 ms |
| 32 KB | 500.00 | 20.47 MB/s | 655.07 | 3.05 ms | 3.08 ms | 3.20 ms | 3.34 ms | 3.55 ms | 1.58 ms | 3.55 ms |
| 64 KB | 500.00 | 41.63 MB/s | 666.02 | 3.00 ms | 3.04 ms | 3.08 ms | 3.21 ms | 6.98 ms | 1.57 ms | 6.98 ms |
| 128 KB | 500.00 | 80.84 MB/s | 646.70 | 3.08 ms | 3.09 ms | 3.17 ms | 3.20 ms | 3.25 ms | 1.63 ms | 3.25 ms |
| 256 KB | 500.00 | 162.61 MB/s | 650.46 | 3.06 ms | 3.07 ms | 3.11 ms | 3.19 ms | 7.03 ms | 1.65 ms | 7.03 ms |
| 512 KB | 500.00 | 317.16 MB/s | 634.32 | 3.14 ms | 3.16 ms | 3.21 ms | 3.32 ms | 3.47 ms | 1.72 ms | 3.47 ms |
| 1 MB | 500.00 | 618.08 MB/s | 618.08 | 3.21 ms | 3.28 ms | 3.36 ms | 4.31 ms | 10.41 ms | 1.81 ms | 10.41 ms |
| 2 MB | 500.00 | 1225.17 MB/s | 612.59 | 3.22 ms | 3.36 ms | 3.51 ms | 3.84 ms | 8.06 ms | 2.30 ms | 8.06 ms |
| 4 MB | 500.00 | 2311.46 MB/s | 577.86 | 3.37 ms | 3.16 ms | 3.75 ms | 4.77 ms | 7.59 ms | 3.02 ms | 7.59 ms |
| 8 MB | 500.00 | 3232.52 MB/s | 404.07 | 4.80 ms | 4.71 ms | 4.78 ms | 9.08 ms | 9.22 ms | 4.55 ms | 9.22 ms |
| 16 MB | 500.00 | 3503.76 MB/s | 218.98 | 8.13 ms | 7.97 ms | 8.16 ms | 12.40 ms | 12.63 ms | 7.52 ms | 12.63 ms |
| 32 MB | 500.00 | 3531.76 MB/s | 110.37 | 14.79 ms | 14.26 ms | 18.38 ms | 20.12 ms | 22.20 ms | 12.22 ms | 22.20 ms |
| 64 MB | 500.00 | 4219.02 MB/s | 65.92 | 24.57 ms | 23.98 ms | 28.22 ms | 32.50 ms | 36.23 ms | 14.20 ms | 36.23 ms |
| 128 MB | 500.00 | 4320.41 MB/s | 33.75 | 47.60 ms | 46.41 ms | 56.92 ms | 61.23 ms | 65.29 ms | 28.00 ms | 65.29 ms |

### Observations

- Peak effective throughput **4320.41 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **2.81 ms** with **4 KB** dirty size.
- Fsync latency with 128 MB dirty data is **16.9x** the latency with 4 KB.

## Fsync Results (QD=4)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 500.00 | 3.71 MB/s | 948.99 | 4.15 ms | 3.55 ms | 6.19 ms | 6.58 ms | 8.87 ms | 1.63 ms | 8.87 ms |
| 8 KB | 500.00 | 8.02 MB/s | 1.03K | 3.79 ms | 3.59 ms | 4.95 ms | 6.67 ms | 8.07 ms | 1.67 ms | 8.07 ms |
| 16 KB | 500.00 | 13.67 MB/s | 875.10 | 4.51 ms | 3.58 ms | 6.28 ms | 6.65 ms | 10.20 ms | 1.66 ms | 10.20 ms |
| 32 KB | 500.00 | 34.69 MB/s | 1.11K | 3.50 ms | 3.49 ms | 4.82 ms | 6.20 ms | 7.91 ms | 1.68 ms | 7.91 ms |
| 64 KB | 500.00 | 55.00 MB/s | 880.05 | 4.48 ms | 4.60 ms | 6.28 ms | 10.42 ms | 17.29 ms | 1.56 ms | 17.29 ms |
| 128 KB | 500.00 | 109.62 MB/s | 876.98 | 4.44 ms | 3.51 ms | 6.61 ms | 7.26 ms | 10.50 ms | 1.57 ms | 10.50 ms |
| 256 KB | 500.00 | 217.04 MB/s | 868.17 | 4.45 ms | 3.36 ms | 6.57 ms | 6.68 ms | 9.26 ms | 1.60 ms | 9.26 ms |
| 512 KB | 500.00 | 445.03 MB/s | 890.07 | 4.37 ms | 3.42 ms | 6.78 ms | 6.92 ms | 6.97 ms | 1.64 ms | 6.97 ms |
| 1 MB | 500.00 | 901.95 MB/s | 901.95 | 4.27 ms | 3.43 ms | 6.82 ms | 7.01 ms | 10.67 ms | 1.75 ms | 10.67 ms |
| 2 MB | 500.00 | 1867.20 MB/s | 933.60 | 4.11 ms | 3.59 ms | 7.11 ms | 7.73 ms | 7.97 ms | 1.91 ms | 7.97 ms |
| 4 MB | 500.00 | 2891.91 MB/s | 722.98 | 5.45 ms | 5.12 ms | 6.52 ms | 9.45 ms | 10.69 ms | 4.23 ms | 10.69 ms |
| 8 MB | 500.00 | 3423.18 MB/s | 427.90 | 8.79 ms | 8.34 ms | 10.89 ms | 12.81 ms | 14.03 ms | 7.02 ms | 14.03 ms |
| 16 MB | 500.00 | 3804.31 MB/s | 237.77 | 15.21 ms | 14.86 ms | 18.43 ms | 20.09 ms | 25.71 ms | 7.97 ms | 25.71 ms |
| 32 MB | 500.00 | 4411.62 MB/s | 137.86 | 25.81 ms | 25.34 ms | 30.05 ms | 35.95 ms | 43.02 ms | 9.98 ms | 43.02 ms |
| 64 MB | 500.00 | 4309.31 MB/s | 67.33 | 49.78 ms | 52.79 ms | 59.91 ms | 72.19 ms | 84.51 ms | 14.20 ms | 84.51 ms |
| 128 MB | 500.00 | 4272.16 MB/s | 33.38 | 95.82 ms | 106.37 ms | 124.65 ms | 137.04 ms | 144.06 ms | 26.98 ms | 144.06 ms |

### Observations

- Peak effective throughput **4411.62 MB/s** achieved with **32 MB** dirty size.
- Lowest average fsync latency **3.50 ms** with **32 KB** dirty size.
- Fsync latency with 128 MB dirty data is **23.1x** the latency with 4 KB.

## Fsync Results (QD=8)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 480.00 | 4.80 MB/s | 1.23K | 5.89 ms | 5.68 ms | 9.54 ms | 14.06 ms | 17.11 ms | 1.72 ms | 17.11 ms |
| 8 KB | 480.00 | 10.19 MB/s | 1.30K | 5.60 ms | 5.65 ms | 8.13 ms | 12.27 ms | 13.75 ms | 1.72 ms | 13.75 ms |
| 16 KB | 480.00 | 21.99 MB/s | 1.41K | 5.28 ms | 5.18 ms | 8.02 ms | 11.59 ms | 13.96 ms | 1.73 ms | 13.96 ms |
| 32 KB | 480.00 | 43.17 MB/s | 1.38K | 5.39 ms | 5.27 ms | 8.46 ms | 11.78 ms | 13.55 ms | 1.57 ms | 13.55 ms |
| 64 KB | 480.00 | 70.28 MB/s | 1.12K | 6.56 ms | 6.48 ms | 10.03 ms | 13.30 ms | 16.61 ms | 1.60 ms | 16.61 ms |
| 128 KB | 480.00 | 123.90 MB/s | 991.18 | 7.32 ms | 6.86 ms | 11.17 ms | 12.75 ms | 13.20 ms | 1.71 ms | 13.20 ms |
| 256 KB | 480.00 | 315.32 MB/s | 1.26K | 5.96 ms | 4.72 ms | 9.72 ms | 12.60 ms | 14.91 ms | 1.60 ms | 14.91 ms |
| 512 KB | 480.00 | 487.59 MB/s | 975.17 | 7.74 ms | 7.78 ms | 11.31 ms | 14.21 ms | 17.71 ms | 1.81 ms | 17.71 ms |
| 1 MB | 480.00 | 805.62 MB/s | 805.62 | 9.56 ms | 9.96 ms | 13.02 ms | 14.90 ms | 18.08 ms | 1.83 ms | 18.08 ms |
| 2 MB | 480.00 | 1479.69 MB/s | 739.84 | 10.46 ms | 10.20 ms | 13.25 ms | 20.98 ms | 26.59 ms | 1.96 ms | 26.59 ms |
| 4 MB | 480.00 | 2415.04 MB/s | 603.76 | 12.79 ms | 12.95 ms | 16.09 ms | 19.49 ms | 22.55 ms | 2.37 ms | 22.55 ms |
| 8 MB | 480.00 | 3109.18 MB/s | 388.65 | 19.48 ms | 18.38 ms | 24.96 ms | 30.27 ms | 34.51 ms | 4.73 ms | 34.51 ms |
| 16 MB | 480.00 | 3889.67 MB/s | 243.10 | 30.87 ms | 30.85 ms | 37.57 ms | 44.37 ms | 56.00 ms | 4.90 ms | 56.00 ms |
| 32 MB | 480.00 | 4053.66 MB/s | 126.68 | 56.44 ms | 57.93 ms | 72.48 ms | 93.35 ms | 102.23 ms | 7.87 ms | 102.23 ms |
| 64 MB | 480.00 | 4469.52 MB/s | 69.84 | 99.10 ms | 101.30 ms | 122.51 ms | 143.37 ms | 149.38 ms | 14.15 ms | 149.38 ms |
| 128 MB | 480.00 | 4582.18 MB/s | 35.80 | 197.52 ms | 202.97 ms | 230.91 ms | 253.81 ms | 268.27 ms | 36.88 ms | 268.27 ms |

### Observations

- Peak effective throughput **4582.18 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **5.28 ms** with **16 KB** dirty size.
- Fsync latency with 128 MB dirty data is **33.6x** the latency with 4 KB.

## Fsync Results (QD=16)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 480.00 | 6.62 MB/s | 1.69K | 8.77 ms | 8.57 ms | 14.04 ms | 19.65 ms | 22.74 ms | 1.78 ms | 22.74 ms |
| 8 KB | 480.00 | 12.11 MB/s | 1.55K | 9.56 ms | 8.63 ms | 14.72 ms | 18.85 ms | 19.13 ms | 2.09 ms | 19.13 ms |
| 16 KB | 480.00 | 23.16 MB/s | 1.48K | 10.02 ms | 9.33 ms | 16.50 ms | 20.12 ms | 21.70 ms | 2.08 ms | 21.70 ms |
| 32 KB | 480.00 | 43.91 MB/s | 1.41K | 10.64 ms | 10.70 ms | 16.22 ms | 20.14 ms | 21.60 ms | 1.71 ms | 21.60 ms |
| 64 KB | 480.00 | 92.68 MB/s | 1.48K | 9.93 ms | 9.97 ms | 14.47 ms | 17.84 ms | 19.55 ms | 1.84 ms | 19.55 ms |
| 128 KB | 480.00 | 149.32 MB/s | 1.19K | 12.52 ms | 12.99 ms | 17.52 ms | 20.44 ms | 23.21 ms | 1.85 ms | 23.21 ms |
| 256 KB | 480.00 | 268.37 MB/s | 1.07K | 14.26 ms | 14.94 ms | 19.74 ms | 23.32 ms | 28.33 ms | 2.10 ms | 28.33 ms |
| 512 KB | 480.00 | 405.52 MB/s | 811.04 | 18.72 ms | 19.31 ms | 27.25 ms | 33.27 ms | 36.78 ms | 3.63 ms | 36.78 ms |
| 1 MB | 480.00 | 944.38 MB/s | 944.38 | 15.69 ms | 16.17 ms | 21.32 ms | 25.57 ms | 29.04 ms | 3.10 ms | 29.04 ms |
| 2 MB | 480.00 | 1534.79 MB/s | 767.40 | 19.45 ms | 20.40 ms | 25.65 ms | 30.02 ms | 37.71 ms | 3.93 ms | 37.71 ms |
| 4 MB | 480.00 | 2422.46 MB/s | 605.61 | 24.75 ms | 25.09 ms | 30.78 ms | 41.56 ms | 44.82 ms | 2.82 ms | 44.82 ms |
| 8 MB | 480.00 | 2900.57 MB/s | 362.57 | 40.11 ms | 38.09 ms | 60.01 ms | 72.38 ms | 74.65 ms | 4.78 ms | 74.65 ms |
| 16 MB | 480.00 | 3445.87 MB/s | 215.37 | 69.07 ms | 74.40 ms | 86.11 ms | 95.73 ms | 104.36 ms | 12.90 ms | 104.36 ms |
| 32 MB | 480.00 | 3801.42 MB/s | 118.79 | 123.20 ms | 131.47 ms | 144.48 ms | 160.99 ms | 177.72 ms | 8.29 ms | 177.72 ms |
| 64 MB | 480.00 | 3948.01 MB/s | 61.69 | 239.49 ms | 251.16 ms | 267.73 ms | 283.64 ms | 290.02 ms | 26.07 ms | 290.02 ms |
| 128 MB | 480.00 | 4100.82 MB/s | 32.04 | 455.96 ms | 479.22 ms | 503.66 ms | 533.72 ms | 551.69 ms | 34.70 ms | 551.69 ms |

### Observations

- Peak effective throughput **4100.82 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **8.77 ms** with **4 KB** dirty size.
- Fsync latency with 128 MB dirty data is **52.0x** the latency with 4 KB.

## Fsync Results (QD=32)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 480.00 | 8.39 MB/s | 2.15K | 13.53 ms | 11.31 ms | 22.96 ms | 32.05 ms | 35.76 ms | 3.14 ms | 35.76 ms |
| 8 KB | 480.00 | 14.09 MB/s | 1.80K | 16.40 ms | 15.99 ms | 24.50 ms | 33.57 ms | 35.41 ms | 1.92 ms | 35.41 ms |
| 16 KB | 480.00 | 33.50 MB/s | 2.14K | 13.96 ms | 14.50 ms | 19.91 ms | 26.26 ms | 27.75 ms | 1.83 ms | 27.75 ms |
| 32 KB | 480.00 | 47.92 MB/s | 1.53K | 19.59 ms | 20.22 ms | 26.81 ms | 31.33 ms | 33.80 ms | 1.73 ms | 33.80 ms |
| 64 KB | 480.00 | 104.69 MB/s | 1.68K | 17.96 ms | 18.24 ms | 24.49 ms | 28.92 ms | 31.16 ms | 1.81 ms | 31.16 ms |
| 128 KB | 480.00 | 187.14 MB/s | 1.50K | 19.72 ms | 20.64 ms | 27.99 ms | 33.11 ms | 35.35 ms | 1.79 ms | 35.35 ms |
| 256 KB | 480.00 | 332.74 MB/s | 1.33K | 22.25 ms | 23.79 ms | 31.94 ms | 39.18 ms | 42.97 ms | 2.20 ms | 42.97 ms |
| 512 KB | 480.00 | 525.88 MB/s | 1.05K | 27.45 ms | 29.19 ms | 41.74 ms | 52.01 ms | 56.65 ms | 2.94 ms | 56.65 ms |
| 1 MB | 480.00 | 1005.71 MB/s | 1.01K | 29.24 ms | 30.30 ms | 42.40 ms | 53.91 ms | 55.80 ms | 4.82 ms | 55.80 ms |
| 2 MB | 480.00 | 1482.04 MB/s | 741.02 | 38.74 ms | 39.55 ms | 51.68 ms | 64.33 ms | 71.09 ms | 4.05 ms | 71.09 ms |
| 4 MB | 480.00 | 2486.70 MB/s | 621.68 | 44.61 ms | 47.28 ms | 65.96 ms | 76.11 ms | 84.72 ms | 3.27 ms | 84.72 ms |
| 8 MB | 480.00 | 2947.61 MB/s | 368.45 | 74.95 ms | 83.59 ms | 97.00 ms | 106.48 ms | 108.58 ms | 3.46 ms | 108.58 ms |
| 16 MB | 480.00 | 3125.12 MB/s | 195.32 | 145.52 ms | 145.47 ms | 196.75 ms | 219.27 ms | 231.35 ms | 22.24 ms | 231.35 ms |
| 32 MB | 480.00 | 3667.21 MB/s | 114.60 | 249.22 ms | 254.83 ms | 277.18 ms | 290.41 ms | 301.27 ms | 76.35 ms | 301.27 ms |
| 64 MB | 480.00 | 3953.76 MB/s | 61.78 | 443.41 ms | 465.86 ms | 499.58 ms | 523.79 ms | 538.74 ms | 138.27 ms | 538.74 ms |
| 128 MB | 480.00 | 4357.16 MB/s | 34.04 | 741.79 ms | 825.50 ms | 877.11 ms | 983.34 ms | 1.00 s | 15.36 ms | 1.00 s |

### Observations

- Peak effective throughput **4357.16 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **13.53 ms** with **4 KB** dirty size.
- Fsync latency with 128 MB dirty data is **54.8x** the latency with 4 KB.

## Fsync Results (QD=64)

| Dirty Size | Fsyncs | Throughput | Fsync/s | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|-----------:|-------:|-----------:|--------:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 320.00 | 14.09 MB/s | 3.61K | 14.21 ms | 13.40 ms | 20.12 ms | 23.74 ms | 24.03 ms | 1.96 ms | 24.03 ms |
| 8 KB | 320.00 | 23.30 MB/s | 2.98K | 15.90 ms | 15.08 ms | 26.04 ms | 28.41 ms | 29.04 ms | 1.88 ms | 29.04 ms |
| 16 KB | 320.00 | 56.39 MB/s | 3.61K | 12.40 ms | 13.46 ms | 16.51 ms | 18.64 ms | 20.41 ms | 1.88 ms | 20.41 ms |
| 32 KB | 320.00 | 113.20 MB/s | 3.62K | 14.12 ms | 15.18 ms | 18.41 ms | 19.33 ms | 19.51 ms | 1.70 ms | 19.51 ms |
| 64 KB | 320.00 | 189.40 MB/s | 3.03K | 15.38 ms | 15.19 ms | 21.57 ms | 22.67 ms | 24.35 ms | 1.71 ms | 24.35 ms |
| 128 KB | 320.00 | 323.85 MB/s | 2.59K | 19.98 ms | 19.67 ms | 26.79 ms | 31.65 ms | 31.79 ms | 1.88 ms | 31.79 ms |
| 256 KB | 320.00 | 761.35 MB/s | 3.05K | 16.93 ms | 17.92 ms | 20.00 ms | 21.43 ms | 21.57 ms | 2.29 ms | 21.57 ms |
| 512 KB | 320.00 | 1341.02 MB/s | 2.68K | 19.90 ms | 19.42 ms | 28.84 ms | 30.77 ms | 30.89 ms | 3.24 ms | 30.89 ms |
| 1 MB | 320.00 | 1438.25 MB/s | 1.44K | 35.69 ms | 38.80 ms | 44.49 ms | 45.85 ms | 46.38 ms | 4.50 ms | 46.38 ms |
| 2 MB | 320.00 | 2356.89 MB/s | 1.18K | 39.15 ms | 40.95 ms | 47.39 ms | 54.02 ms | 54.14 ms | 3.99 ms | 54.14 ms |
| 4 MB | 320.00 | 2744.26 MB/s | 686.06 | 60.03 ms | 71.72 ms | 93.84 ms | 99.02 ms | 99.80 ms | 4.24 ms | 99.80 ms |
| 8 MB | 320.00 | 3345.81 MB/s | 418.23 | 108.86 ms | 125.91 ms | 139.57 ms | 147.29 ms | 152.02 ms | 22.54 ms | 152.02 ms |
| 16 MB | 320.00 | 3556.86 MB/s | 222.30 | 218.10 ms | 234.90 ms | 252.39 ms | 261.10 ms | 268.45 ms | 23.91 ms | 268.45 ms |
| 32 MB | 320.00 | 3747.08 MB/s | 117.10 | 405.73 ms | 440.30 ms | 464.82 ms | 482.06 ms | 500.15 ms | 39.19 ms | 500.15 ms |
| 64 MB | 320.00 | 3916.79 MB/s | 61.20 | 763.86 ms | 830.98 ms | 867.01 ms | 910.47 ms | 943.18 ms | 69.22 ms | 943.18 ms |
| 128 MB | 320.00 | 4327.71 MB/s | 33.81 | 584.24 ms | 628.25 ms | 934.59 ms | 1.07 s | 1.07 s | 18.59 ms | 1.07 s |

### Observations

- Peak effective throughput **4327.71 MB/s** achieved with **128 MB** dirty size.
- Lowest average fsync latency **12.40 ms** with **16 KB** dirty size.
- Fsync latency with 128 MB dirty data is **41.1x** the latency with 4 KB.

## Methodology

1. **Write pattern**: for each dirty size, a temporary file is pre-created
   and pre-extended to the target size. Each iteration overwrites the file
   from offset 0 with pseudo-random bytes (PCG-64 RNG, deterministic seed)
   to defeat filesystem-level compression, then calls `fdatasync`.
2. **Latency measurement**: only the `fdatasync` call is timed; the preceding
   `write(2)` is excluded from the measured latency.
3. **Concurrent mode**: when QD > 1, each of QD goroutines operates on its
   own temporary file, performing write+fsync iterations concurrently.
   This measures how fsync latency scales under device contention.
4. **Aggregation**: when multiple passes are requested, latency samples are
   pooled and throughput / ops-per-sec are averaged across all passes.

---

*Generated by [bench/disk/fsync](https://github.com/rjl493456442/bench)*
