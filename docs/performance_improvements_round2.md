# Performance Improvements Round 2 (2026-02-06)

This document measures one additional optimization pass on top of the prior optimized code.

## What changed

1. Batched band-key encoding with one string conversion per call:
- `MinhashLSH.Add` and `MinhashLSHHeap.Add` now fill one reusable `tmpKeyBuf` for all bands, then do `string(tmpKeyBuf)` once and slice per-band keys from that string.
- `MinhashLSH.query` and `MinhashLSHHeap.query` use the same batched path.
- Files: `lsh.go`, `lsh_heap.go`.

2. Shared direct-width key encoder:
- Added `fillHashKeyBytes` and `fillBandedHashKeys` helpers.
- Uses `PutUint16`/`PutUint32`/`PutUint64` directly, no extra per-value temp buffer/copy path.
- File: `lsh.go`.

3. Query-path allocation tightening:
- Result maps are now lazily allocated only if a match is found.
- `Query` returns `nil` early when there are no candidates.
- Files: `lsh.go`, `lsh_heap.go`.

## Benchmark method

- CPU: AMD Ryzen 9 7950X
- OS/Arch: linux/amd64
- Go: `/usr/local/go/bin/go` (`go1.20.2`)
- Bench command:
  - `/usr/local/go/bin/go test -run '^$' -bench 'Benchmark_(InsertN|QueryN)$' -benchmem`
- Before:
  - `git archive HEAD` (state before this pass), plus same benchmark file.
- After:
  - Working tree with this pass.
- For both before/after:
  - 3 runs each, best (lowest) `ns/op` reported.

### Insert (best of 3 runs)

| Benchmark        | Before ns/op | After ns/op | Speedup | Before allocs | After allocs | Alloc reduction |
| ---------------- | ------------ | ----------- | ------- | ------------- | ------------ | --------------- |
| insert-1k        | 1,388,066    | 1,189,757   | 1.17x   | 28,932        | 1,931        | 14.98x fewer    |
| insert-10k       | 18,890,425   | 16,818,936  | 1.12x   | 289,932       | 19,931       | 14.55x fewer    |
| insert-100k      | 315,641,151  | 237,528,590 | 1.33x   | 2,899,932     | 199,932      | 14.50x fewer    |
| heap-insert-1k   | 516,120      | 378,318     | 1.36x   | 15,089        | 2,086        | 7.23x fewer     |
| heap-insert-10k  | 7,960,776    | 6,243,001   | 1.28x   | 150,187       | 20,184       | 7.44x fewer     |
| heap-insert-100k | 82,926,516   | 66,218,668  | 1.25x   | 1,500,328     | 200,324      | 7.49x fewer     |
| fixed-heap-1k    | 324,434      | 198,374     | 1.64x   | 14,935        | 1,932        | 7.73x fewer     |
| fixed-heap-10k   | 3,317,691    | 1,755,884   | 1.89x   | 149,935       | 19,932       | 7.52x fewer     |
| fixed-heap-100k  | 36,046,011   | 26,149,222  | 1.38x   | 1,499,935     | 199,932      | 7.50x fewer     |

### Query (best of 3 runs)

| Benchmark       | Before ns/op | After ns/op | Speedup | Before allocs | After allocs |
| --------------- | ------------ | ----------- | ------- | ------------- | ------------ |
| query-1k        | 844          | 571         | 1.48x   | 29            | 1            |
| query-10k       | 970          | 699         | 1.39x   | 29            | 1            |
| query-100k      | 1,144        | 888         | 1.29x   | 29            | 1            |
| heap-query-1k   | 785          | 594         | 1.32x   | 15            | 1            |
| heap-query-10k  | 935          | 746         | 1.25x   | 15            | 1            |
| heap-query-100k | 1,083        | 904         | 1.20x   | 15            | 1            |

## Notes

- Biggest win in this pass is allocation collapse:
  - Standard insert allocs/op dropped by ~14.5x.
  - Heap/fixed-heap insert allocs/op dropped by ~7.2x to ~7.7x.
  - Query allocs/op dropped to `1` for both backends.
- Standard backend memory per insert (`B/op`) improved by about `1.20x` (~17% less).
- Query `B/op` improved from `272 -> 112` (standard) and `272 -> 224` (heap).
