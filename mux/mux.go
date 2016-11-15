package mux

import (
	"github.com/fanyang01/httpx/internal/radix"
	"github.com/y0ssar1an/q"
)

type tree struct {
	method string
	ok     bool
	radix.Tree
}

const nSLOT = 11

type hmap [nSLOT]tree

func hash(s string) byte {
	// for i := 0; i < len(s); i++ {
	// 	b ^= s[i] & 0x3b // magic
	// }
	// var b byte = 121
	// b *= s[0]
	// b ^= s[1]
	var b byte = 131
	b *= s[0]
	b += s[1]
	return b % nSLOT
}

func (h *hmap) Add(methods ...string) {
	for _, s := range methods {
		i := hash(s)
		if (*h)[i].ok {
			panic("hmap: conflict hash value")
		}
		(*h)[i].ok = true
		(*h)[i].method = s
	}
}

func (h *hmap) Get(method string) *radix.Tree {
	i := hash(method)
	if (*h)[i].ok && (*h)[i].method == method {
		return &(*h)[i].Tree
	}
	return nil
}

const (
	sMETHOD = "GET" + "POST" + "PUT" + "HEAD" + "DELETE" + "CONNECT" + "OPTIONS" + "PATCH" + "TRACE"
)

var (
	GET     = sMETHOD[:3]
	POST    = sMETHOD[3:7]
	PUT     = sMETHOD[7:10]
	HEAD    = sMETHOD[10:14]
	DELETE  = sMETHOD[14:20]
	CONNECT = sMETHOD[20:27]
	OPTIONS = sMETHOD[27:34]
	PATCH   = sMETHOD[34:39]
	TRACE   = sMETHOD[39:44]
	METHODS = []string{
		GET, POST, PUT, HEAD, DELETE,
		CONNECT, OPTIONS, PATCH, TRACE,
	}
)

type Mux struct {
	hmap   hmap
	extend map[string]*radix.Tree
}

func New() *Mux {
	mux := &Mux{}
	mux.hmap.Add(
		GET, POST, PUT, HEAD, DELETE,
		OPTIONS, PATCH, TRACE, CONNECT,
	)
	q.Q(mux.hmap)
	return mux
}

func (mux *Mux) tree(method string) *radix.Tree {
	t := mux.hmap.Get(method)
	if t == nil {
		t = mux.extend[method]
	}
	return t
}
