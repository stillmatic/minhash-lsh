package minhashlsh

import (
	"fmt"
	"strconv"
	"testing"
)

func Benchmark_InsertN(b *testing.B) {
	nSigs := []int{1_000, 10_000, 100_000}

	for _, nSig := range nSigs {
		b.Run(fmt.Sprintf("insert-%d", nSig), func(b *testing.B) {
			sigs := make([][]uint64, nSig)
			for i := range sigs {
				sigs[i] = randomSignature(64, int64(i))
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				f := NewMinhashLSH16[string](64, 0.5, nSig)
				for i := range sigs {
					f.Add(strconv.Itoa(i), sigs[i])
				}
				f.Index()
			}
		})

		b.Run(fmt.Sprintf("map-insert-%d", nSig), func(b *testing.B) {
			sigs := make([][]uint64, nSig)
			for i := range sigs {
				sigs[i] = randomSignature(64, int64(i))
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				f := NewMinhashLSHMap[string](64, 0.5)
				for i := range sigs {
					f.Add(strconv.Itoa(i), sigs[i])
				}
			}
		})
		b.Run(fmt.Sprintf("map-presized-insert-%d", nSig), func(b *testing.B) {
			sigs := make([][]uint64, nSig)
			for i := range sigs {
				sigs[i] = randomSignature(64, int64(i))
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				f := NewMinhashLSHMapWithSize[string](64, 0.5, nSig)
				for i := range sigs {
					f.Add(strconv.Itoa(i), sigs[i])
				}
			}
		})
	}
}

func Benchmark_QueryN(b *testing.B) {
	nSigs := []int{1_000, 10_000, 100_000}

	for _, nSig := range nSigs {
		b.Run(fmt.Sprintf("query-%d", nSig), func(b *testing.B) {
			sigs := make([][]uint64, nSig)
			for i := range sigs {
				sigs[i] = randomSignature(64, int64(i))
			}
			f := NewMinhashLSH16[string](64, 0.5, nSig)
			for i := range sigs {
				f.Add(strconv.Itoa(i), sigs[i])
			}
			f.Index()
			querySig := randomSignature(64, 999999)
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				f.Query(querySig)
			}
		})

		b.Run(fmt.Sprintf("map-query-%d", nSig), func(b *testing.B) {
			sigs := make([][]uint64, nSig)
			for i := range sigs {
				sigs[i] = randomSignature(64, int64(i))
			}
			f := NewMinhashLSHMapWithSize[string](64, 0.5, nSig)
			for i := range sigs {
				f.Add(strconv.Itoa(i), sigs[i])
			}
			querySig := randomSignature(64, 999999)
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				f.Query(querySig)
			}
		})
	}
}
