package minhashlsh

import (
	"container/heap"
	"sort"
)

type nodeSimilarity[T comparable] struct {
	Key        T
	Similarity string
}

type similarityHeap[T comparable] []nodeSimilarity[T]

func (h similarityHeap[T]) Len() int           { return len(h) }
func (h similarityHeap[T]) Less(i, j int) bool { return h[i].Similarity > h[j].Similarity }
func (h similarityHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *similarityHeap[T]) Push(x any) {
	*h = append(*h, x.(nodeSimilarity[T]))
}

func (h *similarityHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MinhashLSHHeap represents a Minhash LSH object with heap implementation
// It does not require knowing the size of the indexed keys in advance.
// It also 2-3x faster at the cost of increased memory usage.
type MinhashLSHHeap[T comparable] struct {
	k             int
	l             int
	hashTables    []*similarityHeap[T]
	hashKeyFunc   hashKeyFunc
	hashValueSize int
}

func NewMinhashLSHHeap[T comparable](numHash int, threshold float64) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := &similarityHeap[T]{}
		heap.Init(h)
		hashTables[i] = h
	}
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4, // Using 32-bit hash values
		hashTables:    hashTables,
		hashKeyFunc:   hashKeyFuncGen(4),
	}
}

func NewMinhashLSHHeapWithSize[T comparable](numHash int, threshold float64, initSize int) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := make(similarityHeap[T], 0, initSize)
		heap.Init(&h)
		hashTables[i] = &h
	}
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4, // Using 32-bit hash values
		hashTables:    hashTables,
		hashKeyFunc:   hashKeyFuncGen(4),
	}
}

func (f *MinhashLSHHeap[T]) Add(key T, sig []uint64) {
	hashKeys := f.hashKeys(sig)
	for i, hashKey := range hashKeys {
		f.hashTables[i].Push(nodeSimilarity[T]{Key: key, Similarity: hashKey})
	}
}

// Query returns candidate keys given the query signature.
func (f *MinhashLSHHeap[T]) Query(sig []uint64) []T {
	set := f.query(sig)
	results := make([]T, 0, len(set))
	for key := range set {
		results = append(results, key)
	}
	return results
}

func (f *MinhashLSHHeap[T]) query(sig []uint64) map[T]bool {
	hashKeys := f.hashKeys(sig)
	results := make(map[T]bool)
	// Query hash tables using binary search.
	for i := 0; i < f.l; i++ {
		// Only search over the indexed keys.
		hashTable := *f.hashTables[i]
		hashKey := hashKeys[i]
		k := sort.Search(len(hashTable), func(x int) bool {
			return hashTable[x].Similarity >= hashKey
		})
		if k < len(hashTable) && hashTable[k].Similarity == hashKey {
			for j := k; j < len(hashTable) && hashTable[j].Similarity == hashKey; j++ {
				key := hashTable[j].Key
				if _, exist := results[key]; !exist {
					results[key] = true
				}
			}
		}
	}
	return results
}

func (f *MinhashLSHHeap[T]) hashKeys(sig []uint64) []string {
	hs := make([]string, f.l)
	for i := 0; i < f.l; i++ {
		hs[i] = f.hashKeyFunc(sig[i*f.k : (i+1)*f.k])
	}
	return hs
}
