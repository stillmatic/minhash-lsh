package minhashlsh

import (
	"encoding/binary"
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

// hashKeyFuncer stores the hash key function and the buffer for encoding hash values.
// this allows us to reuse without creating new buffers.
type hashKeyFuncer struct {
	hashValueSize int
	s             []byte
}

func newHashKeyFuncer(hashValueSize int, k int) *hashKeyFuncer {
	s := make([]byte, hashValueSize*k)
	return &hashKeyFuncer{hashValueSize: hashValueSize, s: s}
}

func (h *hashKeyFuncer) hashKeyFunc(sig []uint64) string {
	switch h.hashValueSize {
	case 8:
		for i, v := range sig {
			binary.LittleEndian.PutUint64(h.s[i*8:], v)
		}
	case 4:
		for i, v := range sig {
			binary.LittleEndian.PutUint32(h.s[i*4:], uint32(v))
		}
	case 2:
		for i, v := range sig {
			binary.LittleEndian.PutUint16(h.s[i*2:], uint16(v))
		}
	default:
		panic("unsupported hash value size")
	}
	return string(h.s)
}

// MinhashLSHHeap represents a Minhash LSH object with heap implementation
// It does not require knowing the size of the indexed keys in advance.
// It also 2-3x faster at the cost of increased memory usage.
type MinhashLSHHeap[T comparable] struct {
	k           int
	l           int
	hashTables  []*similarityHeap[T]
	tableDirty  []bool
	hashKeyFunc hashKeyFunc
	hs          []string
}

func NewMinhashLSHHeap[T comparable](numHash int, threshold float64) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKLCached(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := &similarityHeap[T]{}
		hashTables[i] = h
	}
	funcer := newHashKeyFuncer(4, k)
	return &MinhashLSHHeap[T]{
		k:           k,
		l:           l,
		hashTables:  hashTables,
		tableDirty:  make([]bool, l),
		hashKeyFunc: funcer.hashKeyFunc,
		hs:          make([]string, l),
	}
}

func NewMinhashLSHHeapWithSize[T comparable](numHash int, threshold float64, initSize int) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKLCached(numHash, threshold)
	hashTables := make([]*similarityHeap[T], l)
	for i := range hashTables {
		h := make(similarityHeap[T], 0, initSize)
		hashTables[i] = &h
	}
	funcer := newHashKeyFuncer(4, k)
	return &MinhashLSHHeap[T]{
		k:           k,
		l:           l,
		hashKeyFunc: funcer.hashKeyFunc,
		hashTables:  hashTables,
		tableDirty:  make([]bool, l),
		hs:          make([]string, l),
	}
}

func (f *MinhashLSHHeap[T]) Add(key T, sig []uint64) {
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	hashKeys := f.hashKeys(sig)
	for i, hashKey := range hashKeys {
		*f.hashTables[i] = append(*f.hashTables[i], nodeSimilarity[T]{Key: key, Value: hashKey})
		f.tableDirty[i] = true
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

func (f *MinhashLSHHeap[T]) query(sig []uint64) map[T]struct{} {
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	hashKeys := f.hashKeys(sig)
	results := make(map[T]struct{})
	// Query hash tables using binary search.
	for i := 0; i < f.l; i++ {
		if f.tableDirty[i] {
			sort.Sort(*f.hashTables[i])
			f.tableDirty[i] = false
		}
		hashTable := *f.hashTables[i]
		hashKey := hashKeys[i]
		k := sort.Search(len(hashTable), func(x int) bool {
			return hashTable[x].Value >= hashKey
		})
		if k < len(hashTable) && hashTable[k].Value == hashKey {
			for j := k; j < len(hashTable) && hashTable[j].Value == hashKey; j++ {
				key := hashTable[j].Key
				results[key] = struct{}{}
			}
		}
	}
	return results
}

func (f *MinhashLSHHeap[T]) hashKeys(sig []uint64) []string {
	for i := 0; i < f.l; i++ {
		f.hs[i] = f.hashKeyFunc(sig[i*f.k : (i+1)*f.k])
	}
	return f.hs
}
