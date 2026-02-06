# Why not a B-tree?

A B-tree (`github.com/google/btree`, degree 32) was evaluated as a potential
backend for interleaved add/query workloads. The idea: O(log n) insert
maintains sorted order without re-sorting, and O(log n) lookup avoids hash
table overhead.

It loses to both existing backends on every metric.

## Interleaved add+query (add 1, query 1, repeat)

| Backend | 1k base | 10k base | 100k base |
|---------|---------|----------|-----------|
| **map** | **5,300 ns** | **6,200 ns** | **6,000 ns** |
| btree | 29,700 ns | 27,600 ns | 36,400 ns |
| reindex | 483,600 ns | 914,000 ns | 29,057,000 ns |

The B-tree is 5-6x slower than the map. It does beat re-indexing sorted slices
(16x at 1k, 800x at 100k), but that's a low bar.

## Bulk insert memory

| Backend | 10k B/op | 10k allocs | 100k B/op | 100k allocs | 100k wall clock |
|---------|---------|-----------|----------|------------|----------------|
| **sorted** | **6.9 MB** | **19,917** | **67.8 MB** | **199,917** | **265 ms** |
| map | 14.9 MB | 19,961 | 135.5 MB | 223,209 | 266 ms |
| btree | 17.8 MB | 170,248 | 174.2 MB | 1,697,796 | 1,446 ms |

The B-tree uses **2.6x more memory** than sorted slices and **1.3x more** than
the map at 100k entries. It also allocates **8.5x more objects** (1.7M vs 200k)
due to tree node allocations, putting more pressure on the GC.

## Why it's slow

1. **Pointer chasing.** B-tree nodes are heap-allocated and linked by pointers.
   Traversing the tree on each insert/query causes cache misses. Hash maps and
   sorted slices store data contiguously.

2. **Indirect comparison function.** `google/btree` takes a `less func(T, T) bool`
   at construction time. Every comparison goes through an indirect function call
   that the compiler cannot inline.

3. **Double traversal on insert.** The B-tree API requires `Get` + `ReplaceOrInsert`
   (2 tree walks) when appending to an existing key's value slice. Maps do this
   in a single operation.

4. **Per-node allocation overhead.** Each B-tree node is a separate heap object.
   With degree 32, nodes hold 31-63 entries. At 100k entries per band × 14 bands,
   that's tens of thousands of node allocations.

## When a B-tree would make sense

- **Range queries** ("find all entries with hash keys between X and Y") — B-trees
  support this in O(log n + k). Maps cannot.
- **Ordered iteration** — B-trees yield entries in sorted order. Not needed for LSH.
- **Memory-constrained environments** where you need a single data structure that
  handles both batch and streaming — though in practice the map still uses less
  memory than the B-tree.

None of these apply to the LSH use case.

## Conclusion

The two-backend design (sorted slices for batch, hash maps for streaming) is the
right call. The B-tree is a "compromise" data structure that ends up worse than
both specialized alternatives.
