package radix

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type rw struct{ s string }

func (r *rw) Header() http.Header {
	panic("not implemented")
}
func (r *rw) Write(b []byte) (int, error) {
	r.s = string(b)
	return len(b), nil
}
func (r *rw) WriteHeader(i int) {
	panic("not implemented")
}

func makeFunc(s string) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write([]byte(s))
	})
}

func fromFunc(f http.HandlerFunc) string {
	rw := new(rw)
	f(rw, nil)
	return rw.s
}

func (t *Tree) _add(path string, v http.HandlerFunc) (ov http.HandlerFunc, replaced bool) {

	return t.Add(path).Replace(v)
}

func TestTree_Add(t *testing.T) {
	tests := []struct {
		path        string
		wantReplace bool
	}{
		{"", false},
		{"/", false},
		{"/pkg", false},
		{"/pkg/", false},
		{"/pkg/net", false},
		{"/doc/", false},
		{"/pkg/net/http/httputil", false},
		{"/pkg/net/http", false},
		{"/pkg/net/http", true},
		{"/pkg/", true},
		{"/pkg", true},
		{"/", true},
		{"", true},
		{"/pkg/net/html", false},
		{"/pkg/net/http/httptest", false},
		{"/pkg/nnn", false},
		{"/pkg/nnnn", false},
		{"/pkg/nn", false},
		{"/pkg/nnn", true},
		{"/pkg/:first/:second/*rest", false},
		{"/pkg/:first", false},
		{"/pkg/:first/:second", false},
		{"/pkg/:first/:second/*rest", true},
	}
	tree := &Tree{}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q(%v)", tt.path, tt.wantReplace), func(t *testing.T) {
			gotOv, gotReplace := tree._add(tt.path, makeFunc(tt.path))
			if gotReplace != tt.wantReplace {
				t.Errorf("Tree.Add() gotReplace = %v, want %v", gotReplace, tt.wantReplace)
			}
			if !gotReplace {
				if gotOv != nil {
					t.Errorf("Tree.Add() gotOv != nil")
				}
				return
			}
			if got := fromFunc(gotOv); got != tt.path {
				t.Errorf("Tree.Add() got = %v, want %v", got, tt.path)
			}
		})
	}
}

func (t *Tree) _lookup(path string) (v http.HandlerFunc, ok bool) {
	node := t.Lookup(path)
	if node == nil || node.HandlerFunc == nil {
		return nil, false
	}
	return node.HandlerFunc, true
}

func testTree_Lookup(t *testing.T, optimize bool) {
	paths := []struct {
		path string
	}{
		{""},
		{"/"},
		{"/pkg"},
		{"/pkg/"},
		{"/pkg/net"},
		{"/doc/"},
		{"/pkg/net/http/httputil"},
		{"/pkg/net/http"},
		{"/pkg/net/html"},
		{"/pkg/net/http/httptest"},
		{"/pkg/nnn"},
		{"/pkg/nnnn"},
		{"/pkg/nn"},
		{"/pkg/:first/:second/*rest"},
		{"/pkg/:first"},
		{"/pkg/:first/:second"},
	}
	tree := &Tree{}
	for _, pp := range paths {
		tree._add(pp.path, makeFunc(pp.path))
	}
	if optimize {
		tree.Optimize()
	}
	fmt.Println(tree)

	tests := []struct {
		path   string
		want   string
		wantOK bool
	}{
		{"", "", true},
		{"/", "/", true},
		{"/pkg", "/pkg", true},
		{"/pkg/", "/pkg/", true},
		{"/pkg/net", "/pkg/net", true},
		{"/doc/", "/doc/", true},
		{"/pkg/net/http/httputil", "/pkg/net/http/httputil", true},
		{"/pkg/net/http", "/pkg/net/http", true},
		{"/pkg/net/html", "/pkg/net/html", true},
		{"/pkg/net/http/httptest", "/pkg/net/http/httptest", true},
		{"/pkg/nnn", "/pkg/nnn", true},
		{"/pkg/nnnn", "/pkg/nnnn", true},
		{"/pkg/nn", "/pkg/nn", true},

		{"/pkg/1", "/pkg/:first", true},
		{"/pkg/1/", "/pkg/:first/:second", true},
		{"/pkg/1/2", "/pkg/:first/:second", true},
		{"/pkg/1/2/", "/pkg/:first/:second/*rest", true},
		{"/pkg/1/2/3", "/pkg/:first/:second/*rest", true},
		{"/pkg/1/2/3/4", "/pkg/:first/:second/*rest", true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q(%v)", tt.path, tt.wantOK), func(t *testing.T) {
			gotV, ok := tree._lookup(tt.path)
			if ok != tt.wantOK {
				t.Errorf("Tree.Lookup() gotOK = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				if gotV != nil {
					t.Errorf("Tree.Lookup() gotV != nil")
				}
				return
			}
			if got := fromFunc(gotV); got != tt.want {
				t.Errorf("Tree.Lookup() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTree_Lookup(t *testing.T) {
	t.Run("non-optimized", func(t *testing.T) {
		testTree_Lookup(t, false)
	})
	t.Run("optimized", func(t *testing.T) {
		testTree_Lookup(t, true)
	})
}

/*

BenchmarkLookup/optimized-4         	30000000	        40.9 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        42.0 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.3 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.9 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.5 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.7 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.3 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.6 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.5 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.8 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.4 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        42.3 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.5 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        42.1 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.6 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.8 ns/op

BenchmarkLookup/optimized-4         	30000000	        40.9 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        42.1 ns/op

BenchmarkLookup/optimized-4         	30000000	        41.4 ns/op
BenchmarkLookup/non-optimized-4     	30000000	        41.8 ns/op

*/
func BenchmarkLookup(b *testing.B) {
	tree := &Tree{}
	tree_o := &Tree{}
	ps := []string{
		"",
		"/src/",
		"/src/*path",
		"/pkg/",
		"/pkg/*path",
		"/doc/",
		"/doc/:doc",
		"/doc/articles/:article",
		"/cmd/",
		"/cmd/:cmd",
		"/cmd/:cmd/",
		"/blog/",
		"/blog/:blog",
		"/*",
	}
	f := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	for _, p := range ps {
		tree.Add(p).Replace(f)
		tree_o.Add(p).Replace(f)
	}
	tree_o.Optimize()

	bt, err := ioutil.ReadFile(filepath.Join("testdata", "url.log"))
	if err != nil {
		b.Fatal(err)
	}
	urls := strings.Split(string(bt), "\n")

	test := func(b *testing.B, T *Tree) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(urls)
			v := T.Lookup(urls[idx])
			if v != nil && v.HandlerFunc != nil {
				v.HandlerFunc(nil, nil)
			}
		}
	}
	b.Run("optimized", func(b *testing.B) {
		test(b, tree_o)
	})
	b.Run("non-optimized", func(b *testing.B) {
		test(b, tree)
	})
}
