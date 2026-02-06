package minhashlsh

import (
	"sort"
)

type nodeSimilarity[T comparable] struct {
	Key   T
	Value string
}

type similarityHeap[T comparable] []nodeSimilarity[T]

func (h similarityHeap[T]) Len() int           { return len(h) }
func (h similarityHeap[T]) Less(i, j int) bool { return h[i].Value < h[j].Value }
func (h similarityHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// MinhashLSHHeap represents a Minhash LSH object with heap implementation
// It does not require knowing the size of the indexed keys in advance.
// It also 2-3x faster at the cost of increased memory usage.
type MinhashLSHHeap[T comparable] struct {
	k             int
	l             int
	hashTables    []*similarityHeap[T]
	tableDirty    []bool
	hashValueSize int
	keySize       int
	tmpKeyBuf     []byte
}

func NewMinhashLSHHeap[T comparable](numHash int, threshold float64) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKLCached(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := &similarityHeap[T]{}
		hashTables[i] = h
	}
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashTables:    hashTables,
		tableDirty:    make([]bool, l),
		hashValueSize: 4,
		keySize:       4 * k,
		tmpKeyBuf:     make([]byte, 4*k*l),
	}
}

func NewMinhashLSHHeapWithSize[T comparable](numHash int, threshold float64, initSize int) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKLCached(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := make(similarityHeap[T], 0, initSize)
		hashTables[i] = &h
	}
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashTables:    hashTables,
		tableDirty:    make([]bool, l),
		hashValueSize: 4,
		keySize:       4 * k,
		tmpKeyBuf:     make([]byte, 4*k*l),
	}
}

func (f *MinhashLSHHeap[T]) Add(key T, sig []uint64) {
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	fillBandedHashKeys(f.tmpKeyBuf, sig, f.hashValueSize, f.keySize, f.k, f.l)
	allKeys := string(f.tmpKeyBuf)
	for i := 0; i < f.l; i++ {
		start := i * f.keySize
		hashKey := allKeys[start : start+f.keySize]
		*f.hashTables[i] = append(*f.hashTables[i], nodeSimilarity[T]{Key: key, Value: hashKey})
		f.tableDirty[i] = true
	}
}

// Query returns candidate keys given the query signature.
func (f *MinhashLSHHeap[T]) Query(sig []uint64) []T {
	set := f.query(sig)
	if len(set) == 0 {
		return nil
	}
	results := make([]T, 0, len(set))
	for key := range set {
		results = append(results, key)
	}
	return results
}

func (f *MinhashLSHHeap[T]) query(sig []uint64) map[T]struct{} {
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	fillBandedHashKeys(f.tmpKeyBuf, sig, f.hashValueSize, f.keySize, f.k, f.l)
	allKeys := string(f.tmpKeyBuf)
	var results map[T]struct{}
	// Query hash tables using binary search.
	for i := 0; i < f.l; i++ {
		if f.tableDirty[i] {
			sort.Sort(*f.hashTables[i])
			f.tableDirty[i] = false
		}
		hashTable := *f.hashTables[i]
		start := i * f.keySize
		hashKey := allKeys[start : start+f.keySize]
		k := sort.Search(len(hashTable), func(x int) bool {
			return hashTable[x].Value >= hashKey
		})
		if k < len(hashTable) && hashTable[k].Value == hashKey {
			if results == nil {
				results = make(map[T]struct{}, 4)
			}
			for j := k; j < len(hashTable) && hashTable[j].Value == hashKey; j++ {
				key := hashTable[j].Key
				results[key] = struct{}{}
			}
		}
	}
	return results
}
