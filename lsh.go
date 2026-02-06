// Package minhashlsh implements Locality Sensitive Hashing using MinHash signatures
package minhashlsh

import (
	"encoding/binary"
	"math"
	"sort"
	"sync"
)

const (
	integrationPrecision = 0.01
)

type hashKeyFunc func([]uint64) string

func fillHashKeyBytes(dst []byte, sig []uint64, hashValueSize int) {
	switch hashValueSize {
	case 8:
		for i, v := range sig {
			binary.LittleEndian.PutUint64(dst[i*8:], v)
		}
	case 4:
		for i, v := range sig {
			binary.LittleEndian.PutUint32(dst[i*4:], uint32(v))
		}
	case 2:
		for i, v := range sig {
			binary.LittleEndian.PutUint16(dst[i*2:], uint16(v))
		}
	default:
		panic("unsupported hash value size")
	}
}

func hashKeyFuncGen(hashValueSize int) hashKeyFunc {
	return func(sig []uint64) string {
		s := make([]byte, hashValueSize*len(sig))
		fillHashKeyBytes(s, sig, hashValueSize)
		return string(s)
	}
}

func fillBandedHashKeys(dst []byte, sig []uint64, hashValueSize, keySize, k, l int) {
	for i := 0; i < l; i++ {
		fillHashKeyBytes(dst[i*keySize:(i+1)*keySize], sig[i*k:(i+1)*k], hashValueSize)
	}
}

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

type optimalKLCacheKey struct {
	numHash       int
	thresholdBits uint64
}

type optimalKLCacheValue struct {
	k  int
	l  int
	fp float64
	fn float64
}

var cachedOptimalKL sync.Map

func optimalKLCached(numHash int, t float64) (k, l int, fp, fn float64) {
	cacheKey := optimalKLCacheKey{
		numHash:       numHash,
		thresholdBits: math.Float64bits(t),
	}
	if v, ok := cachedOptimalKL.Load(cacheKey); ok {
		cached := v.(optimalKLCacheValue)
		return cached.k, cached.l, cached.fp, cached.fn
	}
	k, l, fp, fn = optimalKL(numHash, t)
	cachedOptimalKL.Store(cacheKey, optimalKLCacheValue{k: k, l: l, fp: fp, fn: fn})
	return k, l, fp, fn
}

// entry contains the hash key (from minhash signature) and the indexed key
type entry[T any] struct {
	hashKey string
	key     T
}

// hashTable is a look-up table implemented as a slice sorted by hash keys.
// Look-up operation is implemented using binary search.
type hashTable[T any] []entry[T]

func (h hashTable[T]) Len() int           { return len(h) }
func (h hashTable[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h hashTable[T]) Less(i, j int) bool { return h[i].hashKey < h[j].hashKey }

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
	keySize        int
	tmpKeyBuf      []byte
	numIndexedKeys int
}

func newMinhashLSH[T comparable](threshold float64, numHash, hashValueSize, initSize int) *MinhashLSH[T] {
	k, l, _, _ := optimalKLCached(numHash, threshold)
	hashTables := make([]hashTable[T], l)
	for i := range hashTables {
		hashTables[i] = make(hashTable[T], 0, initSize)
	}
	return &MinhashLSH[T]{
		k:              k,
		l:              l,
		hashValueSize:  hashValueSize,
		keySize:        hashValueSize * k,
		tmpKeyBuf:      make([]byte, hashValueSize*k*l),
		hashTables:     hashTables,
		numIndexedKeys: 0,
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

// Add a key with MinHash signature into the index.
// The key won't be searchable until Index() is called.
func (f *MinhashLSH[T]) Add(key T, sig []uint64) {
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	fillBandedHashKeys(f.tmpKeyBuf, sig, f.hashValueSize, f.keySize, f.k, f.l)
	allKeys := string(f.tmpKeyBuf)
	// Insert keys into the hash tables by appending.
	for i := range f.hashTables {
		start := i * f.keySize
		hashKey := allKeys[start : start+f.keySize]
		f.hashTables[i] = append(f.hashTables[i], entry[T]{hashKey: hashKey, key: key})
	}
}

// Index makes all the keys added searchable.
func (f *MinhashLSH[T]) Index() {
	for i := range f.hashTables {
		sort.Sort(f.hashTables[i])
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
	if len(sig) < f.k*f.l {
		panic("signature length does not match LSH parameters")
	}
	if f.numIndexedKeys == 0 {
		return nil
	}
	fillBandedHashKeys(f.tmpKeyBuf, sig, f.hashValueSize, f.keySize, f.k, f.l)
	allKeys := string(f.tmpKeyBuf)
	var results map[T]struct{}
	// Query hash tables using binary search.
	for i := 0; i < f.l; i++ {
		// Only search over the indexed keys.
		hashTable := f.hashTables[i][:f.numIndexedKeys]
		start := i * f.keySize
		hashKey := allKeys[start : start+f.keySize]
		k := sort.Search(len(hashTable), func(x int) bool {
			return hashTable[x].hashKey >= hashKey
		})
		if k < len(hashTable) && hashTable[k].hashKey == hashKey {
			if results == nil {
				results = make(map[T]struct{}, 4)
			}
			for j := k; j < len(hashTable) && hashTable[j].hashKey == hashKey; j++ {
				key := hashTable[j].key
				results[key] = struct{}{}
			}
		}
	}
	return results
}
