package radix

import (
	"strconv"
	"strings"
	"testing"
)

/*

BenchmarkStringsIndexByte/length=1-4         	100000000	        13.1 ns/op
BenchmarkStringsIndexByte/length=2-4         	100000000	        20.3 ns/op
BenchmarkStringsIndexByte/length=4-4         	50000000	        36.0 ns/op
BenchmarkStringsIndexByte/length=8-4         	20000000	        66.7 ns/op
BenchmarkStringsIndexByte/length=16-4        	10000000	       135 ns/op
BenchmarkStringsIndexByte/length=20-4        	10000000	       179 ns/op
BenchmarkIndexByte/length=1-4                	200000000	         9.51 ns/op
BenchmarkIndexByte/length=2-4                	100000000	        15.6 ns/op
BenchmarkIndexByte/length=4-4                	50000000	        31.6 ns/op
BenchmarkIndexByte/length=8-4                	20000000	        78.3 ns/op
BenchmarkIndexByte/length=16-4               	10000000	       230 ns/op
BenchmarkIndexByte/length=20-4               	 5000000	       366 ns/op

*/

func BenchmarkStringsIndexByte(b *testing.B) {
	const s = "fobarhelwdincxyjvpqz"
	for _, length := range []int{
		1, 2, 4, 8, 16, 20,
	} {
		b.Run("length="+strconv.Itoa(length), func(b *testing.B) {
			s := s[:length]
			k := -1
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := range s {
					k = strings.IndexByte(s, s[i])
					_ = k
				}
			}
		})
	}
}

func BenchmarkIndexByte(b *testing.B) {
	const s = "fobarhelwdincxyjvpqz"
	for _, length := range []int{
		1, 2, 4, 8, 16, 20,
	} {
		b.Run("length="+strconv.Itoa(length), func(b *testing.B) {
			s := s[:length]
			k := -1
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for i := range s {
					for j := 0; i < len(s); j++ {
						if s[j] == s[i] {
							k = j
							break
						}
					}
					_ = k
				}
			}
		})
	}
}
