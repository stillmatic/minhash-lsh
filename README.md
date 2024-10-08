# Minhash LSH in Golang

[![GoDoc](https://godoc.org/github.com/stillmatic/minhash-lsh?status.svg)](https://godoc.org/github.com/stillmatic/minhash-lsh)

This is a fork of [ekzhu/minhash-lsh](github.com/ekzhu/minhash-lsh) modified to support generics and with a new heap backend. Additionally, I've added defaults from the Python package.

Install: `go get github.com/stillmatic/minhash-lsh`

The heap backend is up to 7x faster (inserts) than the original implementation. I've additionally added memory pooling which reduces memory usage by 25% and reduces allocations by 50%.

Example of using this to deduplicate some text:

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

func ExampleMinhashLSHHeap() {
	newsItems := []newsItem{
		{URL: "https://example.com/1", Description: "This is a test"},
		{URL: "https://example.com/2", Description: "This is another test"},
		{URL: "https://example.com/3", Description: "This is a test"},
	}

	// key on the URL, so instantiate with `string` generic
	lsh := minhashlsh.NewMinhashLSHHeapWithSize[string](88, 0.7, len(newsItems))
	for _, item := range newsItems {
		mh := minhashlsh.NewMinhashWithDefaults()
		mh.Push([]byte(item.Description))
		lsh.Add(item.URL, mh.Signature())
	}

	// no need to build index with heap backend

	// find duplicate entries
	dupeKeys := make(map[string]struct{})
	for _, item := range newsItems {
		if _, ok := dupeKeys[item.URL]; ok {
			//already a duplicate
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