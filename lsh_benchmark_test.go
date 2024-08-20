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
			f := NewMinhashLSH16[string](64, 0.5, nSig)
			for i := range sigs {
				f.Add(strconv.Itoa(i), sigs[i])
			}
			f.Index()
		})
	}
}
