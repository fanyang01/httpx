package radix

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type I int

func (I) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {}

func (t *Tree) _add(path string, v http.Handler) (ov http.Handler, replace bool) {
	old, replace := t.Add(path, Payload{v})
	return old.Handler, replace
}

func TestTree_Add(t *testing.T) {
	type args struct {
		path string
		v    http.Handler
	}
	tests := []struct {
		args        args
		wantOv      http.Handler
		wantReplace bool
	}{
		{args{"", I(0)}, nil, false},
		{args{"/", I(1)}, nil, false},
		{args{"/pkg", I(2)}, nil, false},
		{args{"/pkg/", I(3)}, nil, false},
		{args{"/pkg/net", I(4)}, nil, false},
		{args{"/doc/", I(5)}, nil, false},
		{args{"/pkg/net/http/httputil", I(6)}, nil, false},
		{args{"/pkg/net/http", I(7)}, nil, false},
		{args{"/pkg/net/http", I(8)}, I(7), true},
		{args{"/pkg/", I(9)}, I(3), true},
		{args{"/pkg", I(10)}, I(2), true},
		{args{"/", I(11)}, I(1), true},
		{args{"", I(12)}, I(0), true},
		{args{"/pkg/net/html", I(13)}, nil, false},
		{args{"/pkg/net/http/httptest", I(14)}, nil, false},
		{args{"/pkg/nnn", I(15)}, nil, false},
		{args{"/pkg/nnnn", I(16)}, nil, false},
		{args{"/pkg/nn", I(17)}, nil, false},
		{args{"/pkg/nnn", I(18)}, I(15), true},
		{args{"/pkg/:first/:second/*rest", I(19)}, nil, false},
		{args{"/pkg/:first", I(20)}, nil, false},
		{args{"/pkg/:first/:second", I(21)}, nil, false},
		{args{"/pkg/:first/:second/*rest", I(22)}, I(19), true},
	}
	tree := &Tree{}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q=%v", tt.args.path, tt.args.v), func(t *testing.T) {
			gotOv, gotReplace := tree._add(tt.args.path, tt.args.v)
			if gotOv != tt.wantOv {
				t.Errorf("Tree.Add() gotOv = %v, want %v", gotOv, tt.wantOv)
			}
			if gotReplace != tt.wantReplace {
				t.Errorf("Tree.Add() gotReplace = %v, want %v", gotReplace, tt.wantReplace)
			}
		})
	}
}

func (t *Tree) _lookup(path string) (v http.Handler, ok bool) {
	node := t.Lookup(path)
	if node == nil || node.Handler == nil {
		return nil, false
	}
	return node.Handler, true
}

func TestTree_Lookup(t *testing.T) {
	paths := []struct {
		path string
		v    http.Handler
	}{
		{"", I(0)},
		{"/", I(1)},
		{"/pkg", I(2)},
		{"/pkg/", I(3)},
		{"/pkg/net", I(4)},
		{"/doc/", I(5)},
		{"/pkg/net/http/httputil", I(6)},
		{"/pkg/net/http", I(7)},
		{"/pkg/net/html", I(12)},
		{"/pkg/net/http/httptest", I(13)},
		{"/pkg/nnn", I(14)},
		{"/pkg/nnnn", I(15)},
		{"/pkg/nn", I(16)},
		{"/pkg/:first/:second/*rest", I(17)},
		{"/pkg/:first", I(18)},
		{"/pkg/:first/:second", I(19)},
	}
	tree := &Tree{}
	for _, pp := range paths {
		tree._add(pp.path, pp.v)
	}
	tree.Optimize()
	fmt.Println(tree)

	tests := []struct {
		path        string
		wantV       http.Handler
		wantReplace bool
	}{
		{"", I(0), true},
		{"/", I(1), true},
		{"/pkg", I(2), true},
		{"/pkg/", I(3), true},
		{"/pkg/net", I(4), true},
		{"/doc/", I(5), true},
		{"/pkg/net/http/httputil", I(6), true},
		{"/pkg/net/http", I(7), true},
		{"/pkg/net/html", I(12), true},
		{"/pkg/net/http/httptest", I(13), true},
		{"/pkg/nnn", I(14), true},
		{"/pkg/nnnn", I(15), true},
		{"/pkg/nn", I(16), true},

		{"/pkg/1", I(18), true},
		{"/pkg/1/", I(19), true},
		{"/pkg/1/2", I(19), true},
		{"/pkg/1/2/", I(17), true},
		{"/pkg/1/2/3", I(17), true},
		{"/pkg/1/2/3/4", I(17), true},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotV, gotReplace := tree._lookup(tt.path)
			if gotV != tt.wantV {
				t.Errorf("Tree.Lookup() gotV = %v, want %v", gotV, tt.wantV)
			}
			if gotReplace != tt.wantReplace {
				t.Errorf("Tree.Lookup() gotReplace = %v, want %v", gotReplace, tt.wantReplace)
			}
		})
	}
}

// BenchmarkLookup/normal-4         	20000000	        82.9 ns/op
// BenchmarkLookup/optimized-4      	20000000	        82.1 ns/op
func BenchmarkLookup(b *testing.B) {
	tr := &Tree{}
	tro := &Tree{}
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
		tr.Add(p, Payload{f})
		tro.Add(p, Payload{f})
	}
	tro.Optimize()

	bt, err := ioutil.ReadFile(filepath.Join("testdata", "url.log"))
	if err != nil {
		b.Fatal(err)
	}
	urls := strings.Split(string(bt), "\n")
	b.Run("optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(urls)
			v := tro.Lookup(urls[idx])
			if v != nil && v.Handler != nil {
				v.Handler.ServeHTTP(nil, nil)
			}
		}
	})
	b.Run("non-optimized", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(urls)
			v := tr.Lookup(urls[idx])
			if v != nil && v.Handler != nil {
				v.Handler.ServeHTTP(nil, nil)
			}
		}
	})
}
