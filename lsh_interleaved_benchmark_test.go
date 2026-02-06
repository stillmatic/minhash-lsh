package minhashlsh

import (
	"fmt"
	"strconv"
	"testing"
)

// Benchmark_Interleaved measures the realistic pattern: add 1 entry, query 1,
// repeat. This is the primary use case for a streaming/online variant.
//
// Two approaches compared:
//   - map:     MinhashLSHMap  (map-backed, O(1) insert + O(1) lookup)
//   - reindex: MinhashLSH      (sorted slices, explicit Index() after every add)
func Benchmark_Interleaved(b *testing.B) {
	const poolSize = 200_000
	sigs := make([][]uint64, poolSize)
	for i := range sigs {
		sigs[i] = randomSignature(64, int64(i))
	}

	baseSizes := []int{1_000, 10_000, 100_000}

	for _, baseSize := range baseSizes {
		addPool := poolSize - baseSize // signatures available for interleaved adds

		b.Run(fmt.Sprintf("map/%d", baseSize), func(b *testing.B) {
			f := NewMinhashLSHMapWithSize[string](64, 0.5, baseSize)
			for i := 0; i < baseSize; i++ {
				f.Add(strconv.Itoa(i), sigs[i])
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := baseSize + (i % addPool)
				f.Add(strconv.Itoa(idx), sigs[idx])
				f.Query(sigs[i%baseSize])
			}
		})

		b.Run(fmt.Sprintf("reindex/%d", baseSize), func(b *testing.B) {
			f := NewMinhashLSH32[string](64, 0.5, baseSize)
			for i := 0; i < baseSize; i++ {
				f.Add(strconv.Itoa(i), sigs[i])
			}
			f.Index()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := baseSize + (i % addPool)
				f.Add(strconv.Itoa(idx), sigs[idx])
				f.Index()
				f.Query(sigs[i%baseSize])
			}
		})
	}
}
