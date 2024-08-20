## old code

```
goos: darwin
goarch: arm64
pkg: github.com/ekzhu/minhash-lsh
cpu: Apple M1 Max
Benchmark_Insert10000-10    	1000000000	         0.02405 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise64-10       	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.35937500, Absolute Error: 0.05937500
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.31250000, Absolute Error: 0.01250000
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.37500000, Absolute Error: 0.07500000
Data size: 14726530, Real resemblance: 0.30000000, Estimated resemblance: 0.34375000, Absolute Error: 0.04375000
14726530	        88.30 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise128-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.34375000, Absolute Error: 0.04375000
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.28906250, Absolute Error: 0.01093750
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.35156250, Absolute Error: 0.05156250
Data size:  9596599, Real resemblance: 0.30000003, Estimated resemblance: 0.26562500, Absolute Error: 0.03437503
 9596599	       127.6 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise256-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.34375000, Absolute Error: 0.04375000
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.30078125, Absolute Error: 0.00078125
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.30468750, Absolute Error: 0.00468750
Data size:  5955696, Real resemblance: 0.30000003, Estimated resemblance: 0.36328125, Absolute Error: 0.06328122
 5955696	       201.7 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise512-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.33593750, Absolute Error: 0.03593750
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.27148438, Absolute Error: 0.02851562
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.32226562, Absolute Error: 0.02226563
Data size:  3942560, Real resemblance: 0.30000000, Estimated resemblance: 0.26953125, Absolute Error: 0.03046875
 3942560	       304.4 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/ekzhu/minhash-lsh	11.410s
```

## with generics

```
goos: darwin
goarch: arm64
pkg: github.com/ekzhu/minhash-lsh
cpu: Apple M1 Max
Benchmark_Insert10000-10    	1000000000	         0.02392 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise64-10       	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.35937500, Absolute Error: 0.05937500
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.31250000, Absolute Error: 0.01250000
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.37500000, Absolute Error: 0.07500000
Data size: 14795398, Real resemblance: 0.29999997, Estimated resemblance: 0.34375000, Absolute Error: 0.04375003
14795398	        86.32 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise128-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.34375000, Absolute Error: 0.04375000
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.28906250, Absolute Error: 0.01093750
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.35156250, Absolute Error: 0.05156250
Data size: 10006194, Real resemblance: 0.30000008, Estimated resemblance: 0.31250000, Absolute Error: 0.01249992
10006194	       123.6 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise256-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.34375000, Absolute Error: 0.04375000
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.30078125, Absolute Error: 0.00078125
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.30468750, Absolute Error: 0.00468750
Data size:  6179806, Real resemblance: 0.29999987, Estimated resemblance: 0.31640625, Absolute Error: 0.01640638
 6179806	       196.3 ns/op	       0 B/op	       0 allocs/op

BenchmarkMinWise512-10      	Data size:      100, Real resemblance: 0.30000000, Estimated resemblance: 0.33593750, Absolute Error: 0.03593750
Data size:    10000, Real resemblance: 0.30000000, Estimated resemblance: 0.27148438, Absolute Error: 0.02851562
Data size:  1000000, Real resemblance: 0.30000000, Estimated resemblance: 0.32226562, Absolute Error: 0.02226563
Data size:  4001493, Real resemblance: 0.30000002, Estimated resemblance: 0.28320312, Absolute Error: 0.01679690
 4001493	       301.4 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/ekzhu/minhash-lsh	11.453s
```

generic implementation matches exactly and looks like it saves a few nanoseconds.