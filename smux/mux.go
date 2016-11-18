package smux

import (
	"net/http"

	"github.com/fanyang01/httpx/internal/radix"
)

type Method string

const (
	GET     Method = "GET"
	POST           = "POST"
	PUT            = "PUT"
	DELETE         = "DELETE"
	HEAD           = "HEAD"
	PATCH          = "PATCH"
	OPTIONS        = "OPTIONS"
	TRACE          = "TRACE"
	CONNECT        = "CONNECT"
)

type Mux struct {
	hmap   hmap
	extend map[string]*radix.Tree
}

func New() *Mux {
	mux := &Mux{}
	mux.hmap.Add(
		xGET, xPOST, xPUT, xHEAD, xDELETE,
		xOPTIONS, xPATCH, xTRACE, xCONNECT,
	)
	return mux
}

func (mux *Mux) add(method, pattern string, h http.Handler) {
	mux.tree(method).Add(pattern, radix.Payload{Handler: h})
}

func (mux *Mux) Handle(m Method, pattern string, h http.Handler) {
	mux.add(string(m), pattern, h)
}

func (mux *Mux) GET(pattern string, h http.Handler) {
	mux.add(xGET, pattern, h)
}

func (mux *Mux) HEAD(pattern string, h http.Handler) {
	mux.add(xHEAD, pattern, h)
}

func (mux *Mux) POST(pattern string, h http.Handler) {
	mux.add(xPOST, pattern, h)
}

func (mux *Mux) PUT(pattern string, h http.Handler) {
	mux.add(xPUT, pattern, h)
}

func (mux *Mux) DELETE(pattern string, h http.Handler) {
	mux.add(xDELETE, pattern, h)
}
