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

## heap

We can save a decent amount of time by using a heap to store the minhashes. This means that it's already sorted and there is no need to call `index`. This improves speed ~2-3x.

However, it increases memory usage nearly 2x! We can add back the constraint that the heap is a fixed size to reduce memory usage, to about 20% more.

```
goos: darwin
goarch: arm64
pkg: github.com/stillmatic/minhash-lsh
cpu: Apple M1 Max
Benchmark_InsertN/insert-1000-10                                     363           3231877 ns/op          914724 B/op      29931 allocs/op
Benchmark_InsertN/heap-insert-1000-10                                523           2258580 ns/op         1658782 B/op      30071 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-1000-10                     541           2198811 ns/op         1134168 B/op      29931 allocs/op
Benchmark_InsertN/insert-10000-10                                     50          23703342 ns/op         9147493 B/op     299931 allocs/op
Benchmark_InsertN/heap-insert-10000-10                                99          10907846 ns/op        27354528 B/op     300183 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-10000-10                    135           8828093 ns/op        11346948 B/op     299931 allocs/op
Benchmark_InsertN/insert-100000-10                                     4         282012521 ns/op        90444296 B/op    2999932 allocs/op
Benchmark_InsertN/heap-insert-100000-10                               12          97260559 ns/op        316428901 B/op   3000323 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-100000-10                    15          71587672 ns/op        112562433 B/op   2999931 allocs/op
```

## fixed size / alloc hunted hash functions

We can optimize by taking advantage of the fact that we know the size of the key being passed to the hash function and preallocate arrays of the correct size. This saves a lot of allocations because we simply reuse the same arrays over and over.

This is 'sorta' unsafe because the same array could be used in multiple goroutines and concurrent access to a slice is a big no-no. However - this library is already NOT goroutine safe, since we are continually appending to the original slices anyways. In order to be safe, we need to add a mutex. At least we are not worse than the original implementation.

```
Benchmark_InsertN/insert-1000-10         	                     366	   3238409 ns/op	  914725 B/op	   29931 allocs/op
Benchmark_InsertN/heap-insert-1000-10    	                     588	   2036035 ns/op	 1211105 B/op	   15075 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-1000-10         	     612	   1953235 ns/op	  686482 B/op	   14935 allocs/op
Benchmark_InsertN/insert-10000-10                        	      50	  23594039 ns/op	 9147602 B/op	  299931 allocs/op
Benchmark_InsertN/heap-insert-10000-10                   	     145	   8191531 ns/op	22874765 B/op	  150187 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-10000-10        	     175	   6486692 ns/op	 6867263 B/op	  149935 allocs/op
Benchmark_InsertN/insert-100000-10                       	       4	 286056521 ns/op	90443000 B/op	 2999931 allocs/op
Benchmark_InsertN/heap-insert-100000-10                  	      16	  70306562 ns/op	271629217 B/op	 1500327 allocs/op
Benchmark_InsertN/fixed-size-heap-insert-100000-10       	      24	  47045970 ns/op	67762742 B/op	 1499935 allocs/op
```

At this point we've improved speed up to 7x over the original implementation, cut allocations in half, and reduced memory usage by 25%. Not bad!