package mux

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/fanyang01/httpx/internal/radix"
)

func TestNew(t *testing.T) {
	mux := New()
	_ = mux
	f := func(s string, i int) byte {
		b := byte(i)
		b *= s[0]
		b += s[1]
		return b % 11
	}
OUTER:
	for i := 0; i < 256; i++ {
		M := map[byte]bool{}
		for _, method := range METHODS {
			key := f(method, i)
			if M[key] {
				continue OUTER
			}
			M[key] = true
		}
		fmt.Println(i)
	}
}

func BenchmarkMap(b *testing.B) {
	mux := New()
	tree := &radix.Tree{}
	M := map[string]*radix.Tree{
		GET:     tree,
		POST:    tree,
		PUT:     tree,
		HEAD:    tree,
		DELETE:  tree,
		CONNECT: tree,
		OPTIONS: tree,
		PATCH:   tree,
		TRACE:   tree,
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
