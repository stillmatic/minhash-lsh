# Minhash LSH in Golang

[![GoDoc](https://godoc.org/github.com/stillmatic/minhash-lsh?status.svg)](https://godoc.org/github.com/stillmatic/minhash-lsh)

This is a fork of [ekzhu/minhash-lsh](github.com/ekzhu/minhash-lsh) modified to support generics.

Install: `go get github.com/stillmatic/minhash-lsh`

Example of using this to deduplicate some text:

```go
type newsItem struct {
   URL string
   Description string
}

var newsItems []newsItem

// key on the URL, so instantiate with `string` generic
lsh := minhashlsh.NewMinhashLSH[string](88, 0.7, newsItems)
for _, item := range newItems {
   mh := minhashlsh.NewMinhashWithDefaults()
   mh.Push([]byte(item.Description))
   lsh.Add(item.URL, mh.Signature())
}
// build index
lsh.Index()

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
```