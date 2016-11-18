package smux

import (
	"math/rand"
	"testing"

	"github.com/fanyang01/httpx/internal/radix"
)

func TestNew(t *testing.T) {
	mux := New()
	_ = mux
}

/*

BenchmarkMap/map-4         	10000000	       226 ns/op
BenchmarkMap/hmap-4        	20000000	        81.8 ns/op
BenchmarkMap/map-random-4  	20000000	        85.0 ns/op
BenchmarkMap/hmap-random-4 	20000000	        58.1 ns/op

*/
func BenchmarkMap(b *testing.B) {
	mux := New()
	tree := &radix.Tree{}
	M := map[string]*radix.Tree{
		xGET:     tree,
		xPOST:    tree,
		xPUT:     tree,
		xHEAD:    tree,
		xDELETE:  tree,
		xCONNECT: tree,
		xOPTIONS: tree,
		xPATCH:   tree,
		xTRACE:   tree,
	}
	methods := []string{
		"GET", "POST", "PUT", "HEAD", "DELETE", "CONNECT", "OPTIONS", "PATCH", "TRACE",
	}
	b.Run("map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, method := range methods {
				tr := M[method]
				_ = tr
			}
		}
	})
	b.Run("hmap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, method := range methods {
				tr := mux.tree(method)
				_ = tr
			}
		}
	})
	b.Run("map-random", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			method := methods[rand.Intn(len(methods))]
			tr := M[method]
			_ = tr
		}
	})
	b.Run("hmap-random", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			method := methods[rand.Intn(len(methods))]
			tr := mux.tree(method)
			_ = tr
		}
	})
}
