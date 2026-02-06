// Package minhashlsh implements Locality Sensitive Hashing using MinHash signatures
package minhashlsh

import (
	"cmp"
	"encoding/binary"
	"math"
	"slices"
	"unsafe"
)

const (
	integrationPrecision = 0.01
)

// Compute the integral of function f, lower limit a, upper limit l, and
// precision defined as the quantize step
func integral(f func(float64) float64, a, b, precision float64) float64 {
	var area float64
	for x := a; x < b; x += precision {
		area += f(x+0.5*precision) * precision
	}
	return area
}

// Probability density function for false positive
func falsePositive(l, k int) func(float64) float64 {
	return func(j float64) float64 {
		return 1.0 - math.Pow(1.0-math.Pow(j, float64(k)), float64(l))
	}
}

// Probability density function for false negative
func falseNegative(l, k int) func(float64) float64 {
	return func(j float64) float64 {
		return 1.0 - (1.0 - math.Pow(1.0-math.Pow(j, float64(k)), float64(l)))
	}
}

// Compute the cumulative probability of false negative given threshold t
func probFalseNegative(l, k int, t, precision float64) float64 {
	return integral(falseNegative(l, k), t, 1.0, precision)
}

// Compute the cumulative probability of false positive given threshold t
func probFalsePositive(l, k int, t, precision float64) float64 {
	return integral(falsePositive(l, k), 0, t, precision)
}

// optimalKL returns the optimal K and L for Jaccard similarity search,
// and the false positive and negative probabilities.
// t is the Jaccard similarity threshold.
func optimalKL(numHash int, t float64) (optK, optL int, fp, fn float64) {
	minError := math.MaxFloat64
	for l := 1; l <= numHash; l++ {
		for k := 1; k <= numHash; k++ {
			if l*k > numHash {
				continue
			}
			currFp := probFalsePositive(l, k, t, integrationPrecision)
			currFn := probFalseNegative(l, k, t, integrationPrecision)
			currErr := currFn + currFp
			if minError > currErr {
				minError = currErr
				optK = k
				optL = l
				fp = currFp
				fn = currFn
			}
		}
	}
	return
}

// entry contains the hash key (from minhash signature) and the indexed key
type entry[T any] struct {
	hashKey string
	key     T
}

// hashTable is a look-up table implemented as a slice sorted by hash keys.
// Look-up operation is implemented using binary search.
type hashTable[T any] []entry[T]

// MinhashLSH represents a MinHash LSH implemented using LSH Forest
// (http://ilpubs.stanford.edu:8090/678/1/2005-14.pdf).
// It supports query-time setting of the MinHash LSH parameters
// L (number of bands) and
// K (number of hash functions per band).
type MinhashLSH[T comparable] struct {
	k              int
	l              int
	hashTables     []hashTable[T]
	hashValueSize  int
	numIndexedKeys int
	keyBuf  []byte          // reusable buffer for batch hash key computation
	keySize int             // hashValueSize * k (bytes per band key)
	results map[T]struct{}  // reusable between queries (cleared, not reallocated)
}

func newMinhashLSH[T comparable](threshold float64, numHash, hashValueSize, initSize int) *MinhashLSH[T] {
	k, l, _, _ := optimalKL(numHash, threshold)
	hashTables := make([]hashTable[T], l)
	for i := range hashTables {
		hashTables[i] = make(hashTable[T], 0, initSize)
	}
	keySize := hashValueSize * k
	return &MinhashLSH[T]{
		k:              k,
		l:              l,
		hashValueSize:  hashValueSize,
		hashTables:     hashTables,
		numIndexedKeys: 0,
		keyBuf:         make([]byte, l*keySize),
		keySize:        keySize,
	}
}

// NewMinhashLSH64 uses 64-bit hash values and pre-allocation of hash tables.
func NewMinhashLSH64[T comparable](numHash int, threshold float64, initSize int) *MinhashLSH[T] {
	return newMinhashLSH[T](threshold, numHash, 8, initSize)
}

// NewMinhashLSH32 uses 32-bit hash values and pre-allocation of hash tables.
// MinHash signatures with 64 bit hash values will have
// their hash values trimmed.
func NewMinhashLSH32[T comparable](numHash int, threshold float64, initSize int) *MinhashLSH[T] {
	return newMinhashLSH[T](threshold, numHash, 4, initSize)
}

// NewMinhashLSH16 uses 16-bit hash values and pre-allocation of hash tables.
// MinHash signatures with 64 or 32 bit hash values will have
// their hash values trimmed.
func NewMinhashLSH16[T comparable](numHash int, threshold float64, initSize int) *MinhashLSH[T] {
	return newMinhashLSH[T](threshold, numHash, 2, initSize)
}

// NewMinhashLSH is the default constructor uses 32 bit hash value
// with pre-allocation of hash tables.
func NewMinhashLSH[T comparable](numHash int, threshold float64, initSize int) *MinhashLSH[T] {
	return NewMinhashLSH32[T](numHash, threshold, initSize)
}

// NewMinhashLSHWithDefaults is the default constructor with default parameters
func NewMinhashLSHWithDefaults[T comparable](initSize int) *MinhashLSH[T] {
	return NewMinhashLSH32[T](128, 0.9, initSize)
}

// Params returns the LSH parameters k and l
func (f *MinhashLSH[T]) Params() (k, l int) {
	return f.k, f.l
}

// fillHashKeys fills keyBuf with all l band hash keys from sig.
func (f *MinhashLSH[T]) fillHashKeys(sig []uint64) {
	switch f.hashValueSize {
	case 2:
		for i := 0; i < f.l; i++ {
			band := sig[i*f.k : (i+1)*f.k]
			offset := i * f.keySize
			for j, v := range band {
				binary.LittleEndian.PutUint16(f.keyBuf[offset+j*2:], uint16(v))
			}
		}
	case 4:
		for i := 0; i < f.l; i++ {
			band := sig[i*f.k : (i+1)*f.k]
			offset := i * f.keySize
			for j, v := range band {
				binary.LittleEndian.PutUint32(f.keyBuf[offset+j*4:], uint32(v))
			}
		}
	case 8:
		for i := 0; i < f.l; i++ {
			band := sig[i*f.k : (i+1)*f.k]
			offset := i * f.keySize
			for j, v := range band {
				binary.LittleEndian.PutUint64(f.keyBuf[offset+j*8:], v)
			}
		}
	}
}

// Add a key with MinHash signature into the index.
// The key won't be searchable until Index() is called.
func (f *MinhashLSH[T]) Add(key T, sig []uint64) {
	f.fillHashKeys(sig)
	// Single string allocation; substrings share the backing data.
	allKeys := string(f.keyBuf)
	for i := 0; i < f.l; i++ {
		hk := allKeys[i*f.keySize : (i+1)*f.keySize]
		f.hashTables[i] = append(f.hashTables[i], entry[T]{hk, key})
	}
}

// Index makes all the keys added searchable.
func (f *MinhashLSH[T]) Index() {
	for i := range f.hashTables {
		slices.SortFunc(f.hashTables[i], func(a, b entry[T]) int {
			return cmp.Compare(a.hashKey, b.hashKey)
		})
	}
	f.numIndexedKeys = len(f.hashTables[0])
}

// Query returns candidate keys given the query signature.
func (f *MinhashLSH[T]) Query(sig []uint64) []T {
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

func (f *MinhashLSH[T]) query(sig []uint64) map[T]struct{} {
	f.fillHashKeys(sig)
	// Zero-copy string view of keyBuf — safe because allKeys is only used
	// for comparison within this function and keyBuf is not modified.
	allKeys := unsafe.String(unsafe.SliceData(f.keyBuf), len(f.keyBuf))
	// Reuse map between queries; clear() keeps allocated buckets.
	clear(f.results)
	for i := 0; i < f.l; i++ {
		hashTable := f.hashTables[i][:f.numIndexedKeys]
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
