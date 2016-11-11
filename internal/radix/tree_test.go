package radix

import (
	"fmt"
	"testing"
)

func TestTree_Add(t *testing.T) {
	type args struct {
		path string
		v    int
	}
	tests := []struct {
		args        args
		wantOv      int
		wantReplace bool
	}{
		{args{"", 0}, None, false},
		{args{"/", 1}, None, false},
		{args{"/pkg", 2}, None, false},
		{args{"/pkg/", 3}, None, false},
		{args{"/pkg/net", 4}, None, false},
		{args{"/doc/", 5}, None, false},
		{args{"/pkg/net/http/httputil", 6}, None, false},
		{args{"/pkg/net/http", 7}, None, false},
		{args{"/pkg/net/http", 8}, 7, true},
		{args{"/pkg/", 9}, 3, true},
		{args{"/pkg", 10}, 2, true},
		{args{"/", 11}, 1, true},
		{args{"", 12}, 0, true},
	}
	tree := &Tree{}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q=%d", tt.args.path, tt.args.v), func(t *testing.T) {
			gotOv, gotReplace := tree.Add(tt.args.path, tt.args.v)
			if gotOv != tt.wantOv {
				t.Errorf("Tree.Add() gotOv = %v, want %v", gotOv, tt.wantOv)
			}
			if gotReplace != tt.wantReplace {
				t.Errorf("Tree.Add() gotReplace = %v, want %v", gotReplace, tt.wantReplace)
			}
		})
	}
}

func TestTree_Lookup(t *testing.T) {
	paths := []struct {
		path string
		v    int
	}{
		{"", 0},
		{"/", 1},
		{"/pkg", 2},
		{"/pkg/", 3},
		{"/pkg/net", 4},
		{"/doc/", 5},
		{"/pkg/net/http/httputil", 6},
		{"/pkg/net/http", 7},
	}
	tree := &Tree{}
	for _, pp := range paths {
		tree.Add(pp.path, pp.v)
	}

	tests := []struct {
		path      string
		wantV     int
		wantExact bool
	}{
		{"", 0, true},
		{"/", 1, true},
		{"/pkg", 2, true},
		{"/pkg/", 3, true},
		{"/pkg/net", 4, true},
		{"/doc/", 5, true},
		{"/pkg/net/http/httputil", 6, true},
		{"/pkg/net/http", 7, true},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			gotV, gotExact := tree.Lookup(tt.path)
			if gotV != tt.wantV {
				t.Errorf("Tree.Lookup() gotV = %v, want %v", gotV, tt.wantV)
			}
			if gotExact != tt.wantExact {
				t.Errorf("Tree.Lookup() gotExact = %v, want %v", gotExact, tt.wantExact)
			}
		})
	}
}
