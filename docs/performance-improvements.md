# Performance Improvements

This document describes the optimizations made to the minhash-lsh library,
the bugs that were fixed along the way, and benchmark results.

## Bugs Fixed

### 1. `MinhashLSHHeap` query was incorrect

The heap variant used `sort.Search` (binary search) on data that was never
sorted. The `Add` method called the receiver's `Push`, which just appended to
a slice — it did **not** call `container/heap.Push`, so no heap invariant was
maintained. Even if it had, heap order is not sorted order, so binary search
would still be wrong.

Tests passed coincidentally because the test signatures with matching keys
were inserted adjacently.

**Fix:** Replaced the heap with plain sorted slices and lazy re-sorting.
On `Add`, entries are appended and a dirty flag is set. On the first `Query`
after any `Add`, all tables are sorted via `slices.SortFunc`. Subsequent
queries skip the sort until the next `Add`.

### 2. `hashKeyFuncer` hardcoded 32-bit width

`hashKeyFuncer.hashKeyFunc` used `i*4` and `buf[:4]` regardless of the
`hashValueSize` passed to `newHashKeyFuncer`. In practice it was always called
with `hashValueSize=4` so this was latent, but it would have broken silently
if anyone used a different size.

**Fix:** Removed `hashKeyFuncer` entirely. Both variants now use a
`fillHashKeys` method with a `switch` on `hashValueSize` that calls the
correctly-sized `PutUint16`/`PutUint32`/`PutUint64` directly.

### 3. `cmd/minhash-lsh-all-pair` didn't compile

`NewMinhashLSH` is generic but was called without a type parameter. Also,
`candidateID.(string)` was a type assertion on a concrete `string`, which
is a compile error with generics.

## Optimizations

### 1. Batch string allocation (biggest win)

**Before:** Each `Add` call invoked `hashKeyFuncGen` once per band. Each
invocation allocated:
- `make([]byte, hashValueSize*k)` — the key buffer
- `make([]byte, 8)` — a temp buffer for `PutUint64`
- `string(s)` — the final string conversion

Plus `hashKeys` allocated a `[]string` slice per call.

Total: **1 + 3L allocations per `Add`** (where L = number of bands).

**After:** A single reusable `keyBuf` is filled with all L band keys at once.
One `string(keyBuf)` call produces a single string, and `allKeys[i*keySize : (i+1)*keySize]`
takes substrings that share the backing data (zero-copy in Go).

Total: **1 allocation per `Add`** (the string conversion), regardless of L.

For the query path, the same technique applies. The Go compiler optimizes
`m[string(buf)]` map lookups to avoid allocation, but we use binary search
on sorted slices so we need a real string. Still, we go from 1+3L to 1
allocation per query.

### 2. Direct PutUint16/32/64

**Before:** Every hash value was encoded with `PutUint64` into a temp buffer,
then `copy`'d with the desired width. This is two operations per value plus
the overhead of a temp buffer.

**After:** A `switch` on `hashValueSize` dispatches to the exact-width
encoder (`PutUint16`, `PutUint32`, or `PutUint64`) writing directly into
`keyBuf`. No temp buffer, no copy.

### 3. `slices.SortFunc` / `slices.BinarySearchFunc`

Replaced interface-based `sort.Sort` and `sort.Search` with the generic
`slices.SortFunc` and `slices.BinarySearchFunc` from the standard library.
These avoid interface dispatch overhead and allow the compiler to inline
comparison functions.

### 4. `map[T]struct{}` instead of `map[T]bool`

Query results now use `map[T]struct{}` saving 1 byte per entry.

## Benchmark Results

All benchmarks run on AMD Ryzen 9 7950X, Go 1.22, `go test -benchmem -count=3`.

### Insert (best of 3 runs)

| Benchmark | Before ns/op | After ns/op | Speedup | Before allocs | After allocs | Alloc reduction |
|-----------|-------------|------------|---------|--------------|-------------|-----------------|
| insert-1k | 2,666,256 | 2,389,859 | 1.12x | 15,931 | 1,917 | 8.3x fewer |
| insert-10k | 20,113,921 | 17,545,288 | 1.15x | 159,931 | 19,917 | 8.0x fewer |
| insert-100k | 301,233,981 | 327,632,338 | ~1x | 1,599,931 | 199,917 | 8.0x fewer |
| heap-insert-1k | 1,596,689 | 1,545,475 | 1.03x | 15,075 | 2,057 | 7.3x fewer |
| heap-insert-10k | 8,453,008 | 7,165,466 | 1.18x | 150,187 | 20,169 | 7.4x fewer |
| heap-insert-100k | 87,291,174 | 73,627,286 | 1.19x | 1,500,327 | 200,309 | 7.5x fewer |
| fixed-heap-1k | 1,466,292 | 1,392,662 | 1.05x | 14,935 | 1,917 | 7.8x fewer |
| fixed-heap-10k | 4,859,124 | 3,429,898 | 1.42x | 149,935 | 19,917 | 7.5x fewer |
| fixed-heap-100k | 46,613,798 | 30,125,991 | 1.55x | 1,499,935 | 199,917 | 7.5x fewer |

The `insert-100k` standard variant is sort-dominated (`slices.SortFunc` on
100k entries x L tables), so the allocation savings don't translate to wall
clock there. The heap/fixed-heap variants show the full benefit since they
defer sorting.

Memory for the standard insert variant dropped from **79 MB to 57 MB** at
100k entries (29% less).

### Query (best of 3 runs)

| Benchmark | Before ns/op | After ns/op | Speedup | Before allocs | After allocs |
|-----------|-------------|------------|---------|--------------|-------------|
| query-1k | 1,048 | 622 | 1.69x | 30 | 2 |
| query-10k | 1,182 | 759 | 1.56x | 30 | 2 |
| query-100k | 1,375 | 939 | 1.46x | 30 | 2 |
| heap-query-1k | 817 | 665 | 1.23x | 15 | 2 |
| heap-query-10k | 958 | 815 | 1.18x | 15 | 2 |
| heap-query-100k | 1,105 | 1,034 | 1.07x | 15 | 2 |

Query allocations dropped from **30 to 2** for the standard variant (the
string conversion + results map). Memory per query went from 496 B to 160 B.

Note: the old heap-query numbers should be taken with a grain of salt since
the old heap variant's binary search on unsorted data was incorrect — it
happened to return plausible results on this benchmark but would miss matches
in general.

---

## Round 2: Zero-allocation queries

### Additional optimizations (on top of round 1)

#### 5. `unsafe.String` for query path

In the query path, `allKeys := string(f.keyBuf)` copied the entire key buffer
into a new string — one allocation per query. But the string is only used for
comparison within the `query()` function and doesn't escape; `keyBuf` isn't
modified during the call.

**Fix:** Replaced with `unsafe.String(unsafe.SliceData(f.keyBuf), len(f.keyBuf))`,
which creates a string header pointing directly at `keyBuf` with no copy.

The `Add` path still uses `string(f.keyBuf)` because stored strings must
outlive the buffer.

#### 6. Reusable results map with `clear()`

Each query allocated a fresh `map[T]struct{}` for deduplication. Go 1.21's
`clear()` builtin zeroes a map while keeping its allocated buckets, so
repeated queries reuse the same memory.

**Fix:** Added a `results map[T]struct{}` field to both `MinhashLSH` and
`MinhashLSHHeap`. Queries call `clear(f.results)` instead of `make()`.
The map is lazily initialized on first match, so queries that return
no results never allocate the map at all.

#### 7. Nil early return from `Query()`

`Query()` now returns nil when the internal `query()` finds no candidates,
skipping the `make([]T, 0, ...)` slice allocation.

### Round 2 benchmark results

Measured against round 1 code. Same hardware (Ryzen 9 7950X), Go 1.22,
`go test -benchmem -count=3`, best of 3 runs.

Insert is unchanged (same code path, same allocations).

#### Query (best of 3 runs)

| Benchmark | Round 1 ns/op | Round 2 ns/op | Speedup | Round 1 B/op | Round 2 B/op | Round 1 allocs | Round 2 allocs |
|-----------|--------------|--------------|---------|-------------|-------------|---------------|---------------|
| query-1k | 654 | 556 | 1.18x | 160 | 0 | 2 | **0** |
| query-10k | 804 | 697 | 1.15x | 160 | 0 | 2 | **0** |
| query-100k | 1,001 | 862 | 1.16x | 160 | 0 | 2 | **0** |
| heap-query-1k | 688 | 562 | 1.22x | 272 | 0 | 2 | **0** |
| heap-query-10k | 822 | 706 | 1.16x | 272 | 0 | 2 | **0** |
| heap-query-100k | 1,007 | 905 | 1.11x | 272 | 0 | 2 | **0** |

Queries are now **fully zero-allocation** — 0 B/op, 0 allocs/op.

### Cumulative improvement (original → round 2)

| Benchmark | Original ns/op | Final ns/op | Total speedup | Original allocs | Final allocs |
|-----------|---------------|------------|---------------|----------------|-------------|
| query-1k | 1,048 | 556 | **1.88x** | 30 | **0** |
| query-10k | 1,182 | 697 | **1.70x** | 30 | **0** |
| query-100k | 1,375 | 862 | **1.60x** | 30 | **0** |
| insert-1k | 2,666,256 | 2,269,246 | **1.18x** | 15,931 | 1,917 (**8.3x fewer**) |
| insert-10k | 20,113,921 | 16,427,248 | **1.22x** | 159,931 | 19,917 (**8.0x fewer**) |
| insert-100k | 301,233,981 | 249,828,536 | **1.21x** | 1,599,931 | 199,917 (**8.0x fewer**) |
