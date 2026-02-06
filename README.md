# Minhash LSH in Golang

[![GoDoc](https://godoc.org/github.com/stillmatic/minhash-lsh?status.svg)](https://godoc.org/github.com/stillmatic/minhash-lsh)

This is a fork of [ekzhu/minhash-lsh](github.com/ekzhu/minhash-lsh) with generics, two backend implementations, and significant performance optimizations.

Install: `go get github.com/stillmatic/minhash-lsh`

## Two backends

### `MinhashLSH` — batch mode (sorted slices)

Best when you add all entries first, then query many times. Requires an explicit
`Index()` call after adding entries.

- **Insert**: O(1) append per band
- **Index**: O(n log n) sort per band (called once)
- **Query**: O(log n) binary search per band, zero allocations

### `MinhashLSHMap` — online/streaming mode (hash maps)

Best for interleaved add/query workloads where new entries must be immediately
searchable. No `Index()` call needed.

- **Insert**: O(1) map insert per band
- **Query**: O(1) map lookup per band, zero allocations

At 100k entries, the map backend is **5,000x faster** than re-indexing sorted
slices for interleaved add+query workloads.

## Example

Deduplicating text using the map backend:

```go
package minhashlsh_test

import (
	"fmt"
	minhashlsh "github.com/stillmatic/minhash-lsh"
)

type newsItem struct {
	URL         string
	Description string
}

func ExampleMinhashLSHMap() {
	newsItems := []newsItem{
		{URL: "https://example.com/1", Description: "This is a test"},
		{URL: "https://example.com/2", Description: "This is another test"},
		{URL: "https://example.com/3", Description: "This is a test"},
	}

	// key on the URL, so instantiate with `string` generic
	lsh := minhashlsh.NewMinhashLSHMapWithSize[string](88, 0.7, len(newsItems))
	for _, item := range newsItems {
		mh := minhashlsh.NewMinhashWithDefaults()
		mh.Push([]byte(item.Description))
		lsh.Add(item.URL, mh.Signature())
	}

	// no need to build index with map backend

	// find duplicate entries
	dupeKeys := make(map[string]struct{})
	for _, item := range newsItems {
		if _, ok := dupeKeys[item.URL]; ok {
			continue
		}
		mh := minhashlsh.NewMinhashWithDefaults()
		mh.Push([]byte(item.Description))
		queryRes := lsh.Query(mh.Signature())
		if len(queryRes) == 0 {
			continue
		}

		for _, res := range queryRes {
			if res != item.URL {
				dupeKeys[res] = struct{}{}
			}
		}
	}
	// should be 1 duplicate to remove
	fmt.Println(dupeKeys)
}
```

## Performance

See [docs/performance-improvements.md](docs/performance-improvements.md) for
full benchmark tables. Highlights vs the original implementation:

| Metric | Before | After |
|--------|--------|-------|
| Insert allocs (100k) | 1,599,931 | 199,917 (**8x fewer**) |
| Query allocs | 30 | **0** |
| Query latency (100k) | 1,375 ns | 556 ns (**1.9x**) |
| Interleaved add+query (100k) | 37,400,000 ns | 6,600 ns (**5,700x**) |
