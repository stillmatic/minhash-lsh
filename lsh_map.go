package minhashlsh

import (
	"encoding/binary"
	"unsafe"
)

// MinhashLSHMap is a map-backed Minhash LSH optimized for interleaved
// Add/Query workloads. Entries are immediately searchable after Add() with
// O(1) insert and O(1) lookup per band — no sorting or explicit Index()
// call required.
type MinhashLSHMap[T comparable] struct {
	k             int
	l             int
	hashTables    []map[string][]T
	hashValueSize int
	keyBuf        []byte
	keySize       int
	results       map[T]struct{} // reusable between queries
	entryBufs     [][]T          // pre-allocated entry storage per band
	entryPos      []int          // next free position in each entryBuf
}

func NewMinhashLSHMap[T comparable](numHash int, threshold float64) *MinhashLSHMap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]map[string][]T, l)
	for i := range hashTables {
		hashTables[i] = make(map[string][]T)
	}
	keySize := 4 * k // 32-bit hash values
	return &MinhashLSHMap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4,
		hashTables:    hashTables,
		keyBuf:        make([]byte, l*keySize),
		keySize:       keySize,
	}
}

func NewMinhashLSHMapWithSize[T comparable](numHash int, threshold float64, initSize int) *MinhashLSHMap[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]map[string][]T, l)
	entryBufs := make([][]T, l)
	for i := range hashTables {
		hashTables[i] = make(map[string][]T, initSize)
		entryBufs[i] = make([]T, initSize)
	}
	keySize := 4 * k
	return &MinhashLSHMap[T]{
		k:             k,
		l:             l,
		hashValueSize: 4,
		hashTables:    hashTables,
		keyBuf:        make([]byte, l*keySize),
		keySize:       keySize,
		entryBufs:     entryBufs,
		entryPos:      make([]int, l),
	}
}

// fillHashKeys fills keyBuf with all l band hash keys from sig.
// The map variant always uses 32-bit (4 byte) hash values.
func (f *MinhashLSHMap[T]) fillHashKeys(sig []uint64) {
	for i := 0; i < f.l; i++ {
		band := sig[i*f.k : (i+1)*f.k]
		offset := i * f.keySize
		for j, v := range band {
			binary.LittleEndian.PutUint32(f.keyBuf[offset+j*4:], uint32(v))
		}
	}
}

// allocEntry returns a single-element []T backed by the pre-allocated entry
// buffer when available. Falls back to a heap allocation when the buffer is
// exhausted. The returned slice has cap=1 so append on collision correctly
// allocates a new backing array.
func (f *MinhashLSHMap[T]) allocEntry(band int, key T) []T {
	if f.entryBufs != nil && f.entryPos[band] < len(f.entryBufs[band]) {
		pos := f.entryPos[band]
		f.entryBufs[band][pos] = key
		f.entryPos[band]++
		return f.entryBufs[band][pos : pos+1 : pos+1]
	}
	return []T{key}
}

func (f *MinhashLSHMap[T]) Add(key T, sig []uint64) {
	f.fillHashKeys(sig)
	allKeys := string(f.keyBuf)
	for i := 0; i < f.l; i++ {
		hk := allKeys[i*f.keySize : (i+1)*f.keySize]
		if existing, ok := f.hashTables[i][hk]; ok {
			f.hashTables[i][hk] = append(existing, key)
		} else {
			f.hashTables[i][hk] = f.allocEntry(i, key)
		}
	}
}

// Query returns candidate keys given the query signature.
func (f *MinhashLSHMap[T]) Query(sig []uint64) []T {
	f.fillHashKeys(sig)
	// Zero-copy string view for map lookups (read-only, doesn't escape).
	allKeys := unsafe.String(unsafe.SliceData(f.keyBuf), len(f.keyBuf))
	clear(f.results)
	for i := 0; i < f.l; i++ {
		hk := allKeys[i*f.keySize : (i+1)*f.keySize]
		for _, key := range f.hashTables[i][hk] {
			if f.results == nil {
				f.results = make(map[T]struct{})
			}
			f.results[key] = struct{}{}
		}
	}
	if len(f.results) == 0 {
		return nil
	}
	results := make([]T, 0, len(f.results))
	for key := range f.results {
		results = append(results, key)
	}
	return results
}
