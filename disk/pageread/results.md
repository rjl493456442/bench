# Disk Read Performance Benchmark

## Test Environment

| Parameter | Value |
|-----------|-------|
| Date | 2026-03-13 10:17:33 |
| OS | linux |
| Architecture | amd64 |
| Data File | `bench_pageread.bin` |
| File Size | 5120 MB |
| Queue Depths | 1, 4, 8, 16, 32, 64 |
| Passes per Page Size | 3 |
| Cache Bypass Method | `O_DIRECT` |

> Latency values are per individual `read(2)` / `pread(2)` syscall.  
> Throughput and IOPS are averaged across all passes.

## Sequential Read Results (QD=1)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 330.54 MB/s | 84.62K | 11.7 µs | 11.6 µs | 11.8 µs | 15.2 µs | 25.0 µs | 7.9 µs | 8.59 ms |
| 8 KB | 1.97M | 661.20 MB/s | 84.63K | 11.7 µs | 11.6 µs | 11.7 µs | 15.3 µs | 26.5 µs | 8.7 µs | 1.07 ms |
| 16 KB | 983.04K | 1285.12 MB/s | 82.25K | 12.1 µs | 11.9 µs | 12.3 µs | 14.4 µs | 71.9 µs | 10.9 µs | 297.5 µs |
| 32 KB | 491.52K | 1964.11 MB/s | 62.85K | 15.8 µs | 15.7 µs | 16.1 µs | 18.1 µs | 71.3 µs | 14.6 µs | 488.9 µs |
| 64 KB | 245.76K | 2609.71 MB/s | 41.76K | 23.9 µs | 23.7 µs | 24.4 µs | 26.8 µs | 85.8 µs | 21.8 µs | 701.4 µs |
| 128 KB | 122.88K | 2773.68 MB/s | 22.19K | 45.0 µs | 44.7 µs | 45.5 µs | 47.5 µs | 92.6 µs | 36.1 µs | 738.7 µs |
| 256 KB | 61.44K | 2808.88 MB/s | 11.24K | 89.0 µs | 88.9 µs | 90.4 µs | 91.9 µs | 121.2 µs | 63.0 µs | 351.8 µs |
| 512 KB | 30.72K | 2980.08 MB/s | 5.96K | 167.7 µs | 166.4 µs | 172.4 µs | 194.1 µs | 801.0 µs | 142.6 µs | 1.54 ms |
| 1 MB | 15.36K | 4344.29 MB/s | 4.34K | 230.1 µs | 229.1 µs | 236.4 µs | 249.6 µs | 355.9 µs | 222.2 µs | 363.6 µs |

### Observations

- Peak throughput **4344.29 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **13.1x higher** than with 4 KB pages.

## Sequential Read Results (QD=4)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 272.57 MB/s | 69.78K | 56.6 µs | 56.5 µs | 59.6 µs | 77.8 µs | 97.5 µs | 42.7 µs | 1.72 ms |
| 8 KB | 1.97M | 2480.27 MB/s | 317.48K | 12.1 µs | 12.1 µs | 12.8 µs | 13.6 µs | 23.2 µs | 8.9 µs | 444.8 µs |
| 16 KB | 983.04K | 2730.79 MB/s | 174.77K | 22.4 µs | 22.4 µs | 28.1 µs | 30.4 µs | 39.9 µs | 11.3 µs | 750.5 µs |
| 32 KB | 491.52K | 2463.46 MB/s | 78.83K | 50.2 µs | 42.3 µs | 82.8 µs | 155.7 µs | 168.5 µs | 14.5 µs | 812.4 µs |
| 64 KB | 245.76K | 3058.88 MB/s | 48.94K | 81.2 µs | 82.9 µs | 88.3 µs | 119.8 µs | 166.8 µs | 22.4 µs | 966.1 µs |
| 128 KB | 122.88K | 5025.84 MB/s | 40.21K | 98.9 µs | 95.5 µs | 113.6 µs | 123.5 µs | 149.4 µs | 86.5 µs | 996.8 µs |
| 256 KB | 61.44K | 6391.72 MB/s | 25.57K | 155.9 µs | 153.0 µs | 167.1 µs | 205.5 µs | 299.0 µs | 106.2 µs | 450.7 µs |
| 512 KB | 30.72K | 6531.45 MB/s | 13.06K | 305.5 µs | 299.3 µs | 400.5 µs | 480.8 µs | 681.7 µs | 148.3 µs | 1.67 ms |
| 1 MB | 15.36K | 6659.01 MB/s | 6.66K | 600.0 µs | 595.1 µs | 628.0 µs | 751.1 µs | 1.16 ms | 277.1 µs | 1.46 ms |

### Observations

- Peak throughput **6659.01 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **24.4x higher** than with 4 KB pages.

## Sequential Read Results (QD=8)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 491.66 MB/s | 125.87K | 62.7 µs | 59.7 µs | 77.1 µs | 100.0 µs | 123.4 µs | 40.0 µs | 842.4 µs |
| 8 KB | 1.97M | 2746.47 MB/s | 351.55K | 22.1 µs | 22.0 µs | 30.9 µs | 32.7 µs | 43.0 µs | 9.6 µs | 772.3 µs |
| 16 KB | 983.04K | 2306.74 MB/s | 147.63K | 53.5 µs | 28.8 µs | 101.7 µs | 153.8 µs | 165.7 µs | 11.0 µs | 1.02 ms |
| 32 KB | 491.52K | 3333.56 MB/s | 106.67K | 74.3 µs | 79.2 µs | 115.5 µs | 166.2 µs | 179.9 µs | 14.9 µs | 780.6 µs |
| 64 KB | 245.76K | 4628.31 MB/s | 74.05K | 107.2 µs | 118.2 µs | 147.4 µs | 207.8 µs | 258.7 µs | 22.3 µs | 761.8 µs |
| 128 KB | 122.88K | 6467.28 MB/s | 51.74K | 153.9 µs | 150.5 µs | 183.2 µs | 249.1 µs | 357.8 µs | 90.7 µs | 1.03 ms |
| 256 KB | 61.44K | 6579.67 MB/s | 26.32K | 303.3 µs | 225.2 µs | 575.1 µs | 674.8 µs | 827.9 µs | 122.3 µs | 1.57 ms |
| 512 KB | 30.72K | 6658.81 MB/s | 13.32K | 599.9 µs | 628.6 µs | 858.1 µs | 938.2 µs | 1.26 ms | 207.7 µs | 1.55 ms |
| 1 MB | 15.36K | 6659.34 MB/s | 6.66K | 1.20 ms | 1.19 ms | 1.23 ms | 1.42 ms | 1.98 ms | 431.6 µs | 2.15 ms |

### Observations

- Peak throughput **6659.34 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **13.5x higher** than with 4 KB pages.

## Sequential Read Results (QD=16)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 881.72 MB/s | 225.72K | 69.5 µs | 66.2 µs | 91.7 µs | 109.8 µs | 125.2 µs | 38.9 µs | 803.6 µs |
| 8 KB | 1.97M | 2698.96 MB/s | 345.47K | 45.1 µs | 30.7 µs | 86.9 µs | 155.2 µs | 215.1 µs | 9.6 µs | 8.63 ms |
| 16 KB | 983.04K | 3484.20 MB/s | 222.99K | 70.7 µs | 72.2 µs | 95.9 µs | 149.7 µs | 172.3 µs | 12.6 µs | 859.1 µs |
| 32 KB | 491.52K | 5412.53 MB/s | 173.20K | 91.2 µs | 92.2 µs | 97.8 µs | 115.6 µs | 155.2 µs | 68.3 µs | 959.2 µs |
| 64 KB | 245.76K | 6545.18 MB/s | 104.72K | 151.8 µs | 147.5 µs | 179.2 µs | 299.8 µs | 453.0 µs | 81.6 µs | 800.5 µs |
| 128 KB | 122.88K | 6601.35 MB/s | 52.81K | 301.9 µs | 230.8 µs | 550.9 µs | 921.9 µs | 1.08 ms | 90.7 µs | 1.57 ms |
| 256 KB | 61.44K | 6644.86 MB/s | 26.58K | 600.7 µs | 501.8 µs | 930.9 µs | 1.14 ms | 1.41 ms | 119.7 µs | 2.02 ms |
| 512 KB | 30.72K | 6666.15 MB/s | 13.33K | 1.20 ms | 1.23 ms | 1.46 ms | 1.59 ms | 1.98 ms | 255.1 µs | 2.37 ms |
| 1 MB | 15.36K | 6653.64 MB/s | 6.65K | 2.40 ms | 2.39 ms | 2.46 ms | 2.75 ms | 3.32 ms | 486.2 µs | 3.80 ms |

### Observations

- Peak throughput **6666.15 MB/s** achieved with **512 KB** page size.
- Throughput with 1 MB pages is **7.5x higher** than with 4 KB pages.

## Sequential Read Results (QD=32)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 1644.50 MB/s | 420.99K | 73.0 µs | 69.8 µs | 93.2 µs | 112.4 µs | 127.9 µs | 39.3 µs | 876.1 µs |
| 8 KB | 1.97M | 3446.05 MB/s | 441.09K | 69.0 µs | 53.7 µs | 125.1 µs | 186.4 µs | 230.8 µs | 10.0 µs | 8.68 ms |
| 16 KB | 983.04K | 5616.05 MB/s | 359.43K | 86.6 µs | 84.9 µs | 104.7 µs | 147.8 µs | 194.4 µs | 13.8 µs | 1.21 ms |
| 32 KB | 491.52K | 6582.12 MB/s | 210.63K | 150.3 µs | 143.9 µs | 181.4 µs | 310.2 µs | 535.4 µs | 72.3 µs | 985.5 µs |
| 64 KB | 245.76K | 6617.16 MB/s | 105.87K | 300.9 µs | 255.8 µs | 520.0 µs | 796.4 µs | 1.02 ms | 99.0 µs | 1.56 ms |
| 128 KB | 122.88K | 6659.59 MB/s | 53.28K | 599.2 µs | 539.2 µs | 845.9 µs | 1.12 ms | 1.31 ms | 117.7 µs | 1.76 ms |
| 256 KB | 61.44K | 6618.86 MB/s | 26.48K | 1.21 ms | 1.11 ms | 1.53 ms | 1.77 ms | 2.21 ms | 164.7 µs | 2.76 ms |
| 512 KB | 30.72K | 6655.15 MB/s | 13.31K | 2.40 ms | 2.43 ms | 2.70 ms | 2.86 ms | 3.41 ms | 229.6 µs | 3.83 ms |
| 1 MB | 15.36K | 6648.73 MB/s | 6.65K | 4.80 ms | 4.78 ms | 4.91 ms | 5.48 ms | 6.25 ms | 566.9 µs | 7.04 ms |

### Observations

- Peak throughput **6659.59 MB/s** achieved with **128 KB** page size.
- Throughput with 1 MB pages is **4.0x higher** than with 4 KB pages.

## Sequential Read Results (QD=64)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 1917.63 MB/s | 490.91K | 76.4 µs | 73.8 µs | 113.1 µs | 145.5 µs | 170.2 µs | 18.4 µs | 1.39 ms |
| 8 KB | 1.97M | 2138.70 MB/s | 273.75K | 116.0 µs | 112.8 µs | 166.1 µs | 206.3 µs | 243.6 µs | 9.9 µs | 8.76 ms |
| 16 KB | 983.04K | 3756.99 MB/s | 240.45K | 132.2 µs | 128.7 µs | 167.2 µs | 222.7 µs | 286.5 µs | 60.1 µs | 1.11 ms |
| 32 KB | 491.52K | 5601.48 MB/s | 179.25K | 227.7 µs | 214.1 µs | 306.0 µs | 565.4 µs | 881.2 µs | 85.6 µs | 1.21 ms |
| 64 KB | 245.76K | 5344.03 MB/s | 85.50K | 373.2 µs | 262.0 µs | 756.1 µs | 1.54 ms | 2.11 ms | 98.2 µs | 2.90 ms |
| 128 KB | 122.88K | 5493.47 MB/s | 43.95K | 727.0 µs | 379.9 µs | 1.98 ms | 3.38 ms | 3.87 ms | 110.4 µs | 4.35 ms |
| 256 KB | 61.44K | 6442.16 MB/s | 25.77K | 2.24 ms | 2.31 ms | 2.71 ms | 4.65 ms | 5.28 ms | 137.7 µs | 11.15 ms |
| 512 KB | 30.72K | 6626.10 MB/s | 13.25K | 4.60 ms | 4.69 ms | 5.09 ms | 5.32 ms | 6.21 ms | 248.2 µs | 6.45 ms |
| 1 MB | 15.36K | 5968.30 MB/s | 5.97K | 10.67 ms | 9.43 ms | 16.71 ms | 30.99 ms | 37.45 ms | 396.5 µs | 43.35 ms |

### Observations

- Peak throughput **6626.10 MB/s** achieved with **512 KB** page size.
- Throughput with 1 MB pages is **3.1x higher** than with 4 KB pages.

## Random Read Results (QD=1)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 77.05 MB/s | 19.72K | 50.6 µs | 50.6 µs | 52.4 µs | 53.2 µs | 61.2 µs | 21.9 µs | 4.08 ms |
| 8 KB | 1.97M | 100.33 MB/s | 12.84K | 77.8 µs | 77.9 µs | 79.6 µs | 81.3 µs | 87.7 µs | 50.7 µs | 8.64 ms |
| 16 KB | 983.04K | 173.90 MB/s | 11.13K | 89.8 µs | 90.0 µs | 91.9 µs | 92.5 µs | 98.6 µs | 57.7 µs | 880.2 µs |
| 32 KB | 491.52K | 302.52 MB/s | 9.68K | 103.3 µs | 105.2 µs | 106.0 µs | 109.5 µs | 117.2 µs | 71.6 µs | 920.1 µs |
| 64 KB | 245.76K | 596.68 MB/s | 9.55K | 104.7 µs | 105.7 µs | 108.8 µs | 111.0 µs | 119.5 µs | 76.5 µs | 895.6 µs |
| 128 KB | 122.88K | 1094.00 MB/s | 8.75K | 114.2 µs | 116.0 µs | 116.9 µs | 119.7 µs | 130.2 µs | 86.8 µs | 817.4 µs |
| 256 KB | 61.44K | 1964.07 MB/s | 7.86K | 127.2 µs | 127.5 µs | 132.2 µs | 137.0 µs | 166.7 µs | 110.2 µs | 825.5 µs |
| 512 KB | 30.72K | 2823.90 MB/s | 5.65K | 177.0 µs | 184.7 µs | 190.7 µs | 195.6 µs | 235.1 µs | 141.1 µs | 910.1 µs |
| 1 MB | 15.36K | 3748.06 MB/s | 3.75K | 266.8 µs | 270.4 µs | 278.3 µs | 302.4 µs | 416.7 µs | 223.0 µs | 446.4 µs |

### Observations

- Peak throughput **3748.06 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **48.6x higher** than with 4 KB pages.

## Random Read Results (QD=4)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 338.12 MB/s | 86.56K | 45.2 µs | 44.0 µs | 51.9 µs | 67.0 µs | 79.7 µs | 18.5 µs | 3.78 ms |
| 8 KB | 1.97M | 414.59 MB/s | 53.07K | 74.6 µs | 71.3 µs | 82.1 µs | 118.2 µs | 145.7 µs | 41.6 µs | 8.63 ms |
| 16 KB | 983.04K | 739.38 MB/s | 47.32K | 83.9 µs | 81.2 µs | 93.3 µs | 131.9 µs | 161.6 µs | 51.0 µs | 864.7 µs |
| 32 KB | 491.52K | 1198.38 MB/s | 38.35K | 103.7 µs | 99.1 µs | 122.9 µs | 166.3 µs | 205.9 µs | 71.9 µs | 828.2 µs |
| 64 KB | 245.76K | 2143.42 MB/s | 34.29K | 115.9 µs | 108.7 µs | 143.4 µs | 195.9 µs | 245.9 µs | 77.0 µs | 919.1 µs |
| 128 KB | 122.88K | 3508.58 MB/s | 28.07K | 141.8 µs | 134.6 µs | 181.6 µs | 248.6 µs | 307.6 µs | 88.3 µs | 869.0 µs |
| 256 KB | 61.44K | 4936.79 MB/s | 19.75K | 201.7 µs | 190.7 µs | 262.6 µs | 344.3 µs | 408.3 µs | 115.8 µs | 1.14 ms |
| 512 KB | 30.72K | 5547.53 MB/s | 11.10K | 359.8 µs | 343.4 µs | 508.0 µs | 625.6 µs | 708.2 µs | 161.7 µs | 797.6 µs |
| 1 MB | 15.36K | 6262.76 MB/s | 6.26K | 637.9 µs | 610.9 µs | 798.6 µs | 1.03 ms | 1.21 ms | 272.4 µs | 1.93 ms |

### Observations

- Peak throughput **6262.76 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **18.5x higher** than with 4 KB pages.

## Random Read Results (QD=8)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 673.13 MB/s | 172.32K | 45.3 µs | 43.7 µs | 52.1 µs | 71.3 µs | 86.5 µs | 18.6 µs | 798.5 µs |
| 8 KB | 1.97M | 770.67 MB/s | 98.65K | 80.2 µs | 75.1 µs | 102.3 µs | 138.4 µs | 181.8 µs | 41.0 µs | 8.63 ms |
| 16 KB | 983.04K | 1375.11 MB/s | 88.01K | 90.2 µs | 84.1 µs | 114.8 µs | 157.0 µs | 207.9 µs | 50.4 µs | 1.04 ms |
| 32 KB | 491.52K | 2146.38 MB/s | 68.68K | 115.8 µs | 106.0 µs | 150.6 µs | 220.5 µs | 301.4 µs | 73.2 µs | 950.3 µs |
| 64 KB | 245.76K | 3433.62 MB/s | 54.94K | 144.9 µs | 131.9 µs | 201.9 µs | 317.2 µs | 429.5 µs | 80.6 µs | 940.7 µs |
| 128 KB | 122.88K | 4713.17 MB/s | 37.71K | 211.4 µs | 184.6 µs | 330.0 µs | 523.7 µs | 651.5 µs | 93.0 µs | 1.05 ms |
| 256 KB | 61.44K | 5455.37 MB/s | 21.82K | 365.7 µs | 313.8 µs | 626.2 µs | 827.6 µs | 961.5 µs | 124.1 µs | 1.22 ms |
| 512 KB | 30.72K | 5588.58 MB/s | 11.18K | 714.8 µs | 661.7 µs | 1.22 ms | 1.54 ms | 1.75 ms | 166.1 µs | 1.95 ms |
| 1 MB | 15.36K | 6390.75 MB/s | 6.39K | 1.25 ms | 1.21 ms | 1.48 ms | 2.09 ms | 2.55 ms | 240.3 µs | 3.38 ms |

### Observations

- Peak throughput **6390.75 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **9.5x higher** than with 4 KB pages.

## Random Read Results (QD=16)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 1259.89 MB/s | 322.53K | 47.9 µs | 45.4 µs | 58.0 µs | 80.7 µs | 103.5 µs | 18.6 µs | 870.4 µs |
| 8 KB | 1.97M | 1316.41 MB/s | 168.50K | 93.3 µs | 83.6 µs | 126.0 µs | 189.2 µs | 254.3 µs | 46.7 µs | 8.66 ms |
| 16 KB | 983.04K | 2318.47 MB/s | 148.38K | 106.6 µs | 95.9 µs | 145.9 µs | 221.5 µs | 300.5 µs | 53.8 µs | 954.8 µs |
| 32 KB | 491.52K | 3363.15 MB/s | 107.62K | 147.4 µs | 129.5 µs | 215.2 µs | 383.0 µs | 585.9 µs | 75.4 µs | 1.13 ms |
| 64 KB | 245.76K | 4480.19 MB/s | 71.68K | 222.0 µs | 178.5 µs | 380.2 µs | 736.4 µs | 1.01 ms | 82.0 µs | 1.43 ms |
| 128 KB | 122.88K | 5224.80 MB/s | 41.80K | 381.6 µs | 274.8 µs | 784.4 µs | 1.34 ms | 1.68 ms | 92.6 µs | 2.19 ms |
| 256 KB | 61.44K | 5601.43 MB/s | 22.41K | 712.7 µs | 507.9 µs | 1.56 ms | 2.09 ms | 2.40 ms | 121.5 µs | 3.10 ms |
| 512 KB | 30.72K | 5752.31 MB/s | 11.50K | 1.39 ms | 1.22 ms | 2.75 ms | 3.37 ms | 3.83 ms | 169.3 µs | 5.78 ms |
| 1 MB | 15.36K | 6362.58 MB/s | 6.36K | 2.51 ms | 2.42 ms | 3.16 ms | 5.28 ms | 7.65 ms | 275.1 µs | 9.22 ms |

### Observations

- Peak throughput **6362.58 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **5.1x higher** than with 4 KB pages.

## Random Read Results (QD=32)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 2038.07 MB/s | 521.75K | 58.0 µs | 54.4 µs | 75.9 µs | 107.4 µs | 139.3 µs | 20.0 µs | 860.1 µs |
| 8 KB | 1.97M | 1963.30 MB/s | 251.30K | 124.3 µs | 108.8 µs | 188.5 µs | 315.5 µs | 457.2 µs | 46.3 µs | 8.71 ms |
| 16 KB | 983.04K | 3351.43 MB/s | 214.49K | 147.1 µs | 126.3 µs | 230.1 µs | 398.1 µs | 578.6 µs | 57.1 µs | 1.22 ms |
| 32 KB | 491.52K | 4379.31 MB/s | 140.14K | 226.6 µs | 171.1 µs | 407.6 µs | 893.3 µs | 1.38 ms | 78.4 µs | 2.36 ms |
| 64 KB | 245.76K | 4998.90 MB/s | 79.98K | 398.7 µs | 254.4 µs | 856.4 µs | 1.92 ms | 2.63 ms | 85.1 µs | 4.16 ms |
| 128 KB | 122.88K | 5407.49 MB/s | 43.26K | 738.2 µs | 445.4 µs | 1.77 ms | 3.42 ms | 4.11 ms | 104.5 µs | 4.82 ms |
| 256 KB | 61.44K | 5643.23 MB/s | 22.57K | 1.42 ms | 705.8 µs | 3.92 ms | 4.86 ms | 5.44 ms | 127.8 µs | 5.92 ms |
| 512 KB | 30.72K | 5700.09 MB/s | 11.40K | 2.80 ms | 1.80 ms | 6.12 ms | 7.10 ms | 7.92 ms | 176.9 µs | 10.65 ms |
| 1 MB | 15.36K | 6333.53 MB/s | 6.33K | 5.04 ms | 4.83 ms | 6.83 ms | 13.08 ms | 17.17 ms | 264.5 µs | 18.50 ms |

### Observations

- Peak throughput **6333.53 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **3.1x higher** than with 4 KB pages.

## Random Read Results (QD=64)

| Page Size | Reads | Throughput | IOPS | Avg Lat | P50 | P90 | P99 | P99.9 | Min | Max |
|----------:|------:|-----------:|-----:|--------:|----:|----:|----:|------:|----:|----:|
| 4 KB | 3.93M | 2093.84 MB/s | 536.02K | 58.7 µs | 54.9 µs | 77.2 µs | 109.9 µs | 144.7 µs | 20.1 µs | 2.06 ms |
| 8 KB | 1.97M | 2110.76 MB/s | 270.18K | 150.6 µs | 121.5 µs | 249.7 µs | 534.4 µs | 877.6 µs | 52.6 µs | 8.86 ms |
| 16 KB | 983.04K | 3366.56 MB/s | 215.46K | 148.4 µs | 126.5 µs | 232.4 µs | 403.8 µs | 593.2 µs | 59.5 µs | 8.97 ms |
| 32 KB | 491.52K | 4529.79 MB/s | 144.95K | 288.5 µs | 210.4 µs | 540.1 µs | 1.36 ms | 2.49 ms | 76.0 µs | 4.21 ms |
| 64 KB | 245.76K | 4996.10 MB/s | 79.94K | 399.3 µs | 256.2 µs | 857.7 µs | 1.91 ms | 2.64 ms | 83.8 µs | 3.65 ms |
| 128 KB | 122.88K | 5409.40 MB/s | 43.28K | 738.3 µs | 452.6 µs | 1.75 ms | 3.39 ms | 4.08 ms | 97.1 µs | 5.41 ms |
| 256 KB | 61.44K | 5645.00 MB/s | 22.58K | 2.07 ms | 1.15 ms | 5.31 ms | 9.24 ms | 10.20 ms | 131.0 µs | 11.33 ms |
| 512 KB | 30.72K | 5695.68 MB/s | 11.39K | 5.61 ms | 2.29 ms | 12.38 ms | 13.94 ms | 16.13 ms | 180.4 µs | 28.72 ms |
| 1 MB | 15.36K | 5816.75 MB/s | 5.82K | 10.96 ms | 10.67 ms | 15.32 ms | 24.78 ms | 30.64 ms | 279.3 µs | 31.55 ms |

### Observations

- Peak throughput **5816.75 MB/s** achieved with **1 MB** page size.
- Throughput with 1 MB pages is **2.8x higher** than with 4 KB pages.

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
