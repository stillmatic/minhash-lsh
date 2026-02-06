package minhashlsh

import (
	"cmp"
	"encoding/binary"
	"slices"
	"unsafe"
)

// MinhashLSHHeap represents a Minhash LSH that does not require an explicit
// Index() call. All entries are immediately searchable after Add().
// Internally it uses sorted slices with lazy re-sorting on query.
type MinhashLSHHeap[T comparable] struct {
	k             int
	l             int
	hashTables    []hashTable[T]
	hashValueSize int
	sorted        bool
	keyBuf        []byte
	keySize       int
	results       map[T]struct{} // reusable between queries
}

func NewMinhashLSHHeap[T comparable](numHash int, threshold float64) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]hashTable[T], l)
	for i := range hashTables {
		hashTables[i] = make(hashTable[T], 0)
	}
	keySize := 4 * k // 32-bit hash values
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4,
		hashTables:    hashTables,
		sorted:        true,
		keyBuf:        make([]byte, l*keySize),
		keySize:       keySize,
	}
}

func NewMinhashLSHHeapWithSize[T comparable](numHash int, threshold float64, initSize int) *MinhashLSHHeap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]hashTable[T], l)
	for i := range hashTables {
		hashTables[i] = make(hashTable[T], 0, initSize)
	}
	keySize := 4 * k
	return &MinhashLSHHeap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4,
		hashTables:    hashTables,
		sorted:        true,
		keyBuf:        make([]byte, l*keySize),
		keySize:       keySize,
	}
}

// fillHashKeys fills keyBuf with all l band hash keys from sig.
// The heap variant always uses 32-bit (4 byte) hash values.
func (f *MinhashLSHHeap[T]) fillHashKeys(sig []uint64) {
	for i := 0; i < f.l; i++ {
		band := sig[i*f.k : (i+1)*f.k]
		offset := i * f.keySize
		for j, v := range band {
			binary.LittleEndian.PutUint32(f.keyBuf[offset+j*4:], uint32(v))
		}
	}
}

func (f *MinhashLSHHeap[T]) Add(key T, sig []uint64) {
	f.fillHashKeys(sig)
	allKeys := string(f.keyBuf)
	for i := 0; i < f.l; i++ {
		hk := allKeys[i*f.keySize : (i+1)*f.keySize]
		f.hashTables[i] = append(f.hashTables[i], entry[T]{hk, key})
	}
	f.sorted = false
}

func (f *MinhashLSHHeap[T]) ensureSorted() {
	if !f.sorted {
		for i := range f.hashTables {
			slices.SortFunc(f.hashTables[i], func(a, b entry[T]) int {
				return cmp.Compare(a.hashKey, b.hashKey)
			})
		}
		f.sorted = true
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
	f.ensureSorted()
	f.fillHashKeys(sig)
	// Zero-copy string view — safe because only used for comparison here.
	allKeys := unsafe.String(unsafe.SliceData(f.keyBuf), len(f.keyBuf))
	n := len(f.hashTables[0])
	clear(f.results)
	for i := 0; i < f.l; i++ {
		hashTable := f.hashTables[i][:n]
		hashKey := allKeys[i*f.keySize : (i+1)*f.keySize]
		k, found := slices.BinarySearchFunc(hashTable, hashKey, func(e entry[T], target string) int {
			return cmp.Compare(e.hashKey, target)
		})
		if found {
			if f.results == nil {
				f.results = make(map[T]struct{})
			}
			for j := k; j < len(hashTable) && hashTable[j].hashKey == hashKey; j++ {
				f.results[hashTable[j].key] = struct{}{}
			}
		}
	}
	return f.results
}
