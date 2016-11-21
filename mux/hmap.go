package mux

import "github.com/fanyang01/httpx/internal/radix"

type tree struct {
	radix.Tree
	method string
}

const nSLOT = 11

type hmap [nSLOT]tree

func hash(s string) byte {
	b := (s[0] ^ 1) | s[1]
	return b % nSLOT
}

func (h *hmap) Add(methods ...string) {
	for _, s := range methods {
		i := hash(s)
		if (*h)[i].method != "" {
			panic("hmap: conflict hash value")
		}
		(*h)[i].method = s
	}
}

const (
	xMETHOD = "GET" + "POST" + "PUT" + "HEAD" + "DELETE" + "CONNECT" + "OPTIONS" + "PATCH" + "TRACE"
)

var (
	xGET     = xMETHOD[:3]
	xPOST    = xMETHOD[3:7]
	xPUT     = xMETHOD[7:10]
	xHEAD    = xMETHOD[10:14]
	xDELETE  = xMETHOD[14:20]
	xCONNECT = xMETHOD[20:27]
	xOPTIONS = xMETHOD[27:34]
	xPATCH   = xMETHOD[34:39]
	xTRACE   = xMETHOD[39:44]
	xMETHODS = []string{
		xGET, xPOST, xPUT, xHEAD, xDELETE,
		xCONNECT, xOPTIONS, xPATCH, xTRACE,
	}
)

func (mux *Mux) tree(method string) *radix.Tree {
	if len(method) > 1 {
		i := hash(method)
		if mux.hmap[i].method == method {
			return &mux.hmap[i].Tree
		}
	}
	return mux.extended[method]
}
