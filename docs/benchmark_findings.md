# Benchmark Findings (2026-02-06)

Environment:
- CPU: AMD Ryzen 9 7950X
- OS/Arch: linux/amd64
- Go: `/usr/local/go/bin/go`

Method:
- Added `Benchmark_QueryN` to `lsh_benchmark_test.go`.
- Ran benchmarks 3 times on baseline (`git archive HEAD`) and 3 times on optimized code.
- Used best (lowest) `ns/op` from each side for each benchmark.
- Command: `/usr/local/go/bin/go test -run '^$' -bench 'Benchmark_(InsertN|QueryN)$' -benchmem`

### Insert (best of 3 runs)

| Benchmark        | Before ns/op | After ns/op | Speedup | Before allocs | After allocs | Alloc reduction |
| ---------------- | ------------ | ----------- | ------- | ------------- | ------------ | --------------- |
| insert-1k        | 2,662,324    | 1,389,669   | 1.92x   | 30,491        | 28,932       | 1.05x fewer     |
| insert-10k       | 23,504,500   | 18,208,116  | 1.29x   | 300,491       | 289,932      | 1.04x fewer     |
| insert-100k      | 338,616,145  | 284,958,704 | 1.19x   | 3,000,491     | 2,899,934    | 1.03x fewer     |
| heap-insert-1k   | 1,632,746    | 511,247     | 3.19x   | 15,649        | 15,089       | 1.04x fewer     |
| heap-insert-10k  | 9,384,399    | 7,984,156   | 1.18x   | 150,747       | 150,187      | 1.00x fewer     |
| heap-insert-100k | 93,470,445   | 87,426,456  | 1.07x   | 1,500,887     | 1,500,327    | 1.00x fewer     |
| fixed-heap-1k    | 1,472,776    | 325,003     | 4.53x   | 15,495        | 14,935       | 1.04x fewer     |
| fixed-heap-10k   | 4,746,807    | 3,106,277   | 1.53x   | 150,495       | 149,935      | 1.00x fewer     |
| fixed-heap-100k  | 47,477,403   | 39,933,444  | 1.19x   | 1,500,495     | 1,499,935    | 1.00x fewer     |

### Query (best of 3 runs)

| Benchmark       | Before ns/op | After ns/op | Speedup | Before allocs | After allocs |
| --------------- | ------------ | ----------- | ------- | ------------- | ------------ |
| query-1k        | 1,006        | 835         | 1.20x   | 30            | 29           |
| query-10k       | 1,167        | 987         | 1.18x   | 30            | 29           |
| query-100k      | 1,321        | 1,161       | 1.14x   | 30            | 29           |
| heap-query-1k   | 796          | 797         | 1.00x   | 15            | 15           |
| heap-query-10k  | 964          | 932         | 1.03x   | 15            | 15           |
| heap-query-100k | 1,110        | 1,096       | 1.01x   | 15            | 15           |
