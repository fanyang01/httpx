package smux

import (
	"fmt"
	"net/http"

	"github.com/fanyang01/httpx/internal/radix"
)

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
	TRACE   = "TRACE"
	CONNECT = "CONNECT"
)

type Mux struct {
	hmap     hmap
	extend   map[string]*radix.Tree
	notFound http.Handler
}

func New() *Mux {
	mux := &Mux{
		extend:   make(map[string]*radix.Tree),
		notFound: http.HandlerFunc(http.NotFound),
	}
	mux.hmap.Add(
		xGET, xPOST, xPUT, xHEAD, xDELETE,
		xOPTIONS, xPATCH, xTRACE, xCONNECT,
	)
	return mux
}

func (mux *Mux) add(method, pattern string, h http.Handler) {
	t := mux.tree(method)
	if t == nil {
		t = &radix.Tree{}
		mux.extend[method] = t
	}
	if _, replace := t.Add(pattern, radix.Payload{Handler: h}); replace {
		panic(fmt.Errorf("mux: can't override registered pattern %q", pattern))
	}
}

func (mux *Mux) Handle(method, pattern string, h http.Handler) {
	mux.add(method, pattern, h)
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

func (mux *Mux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if t := mux.tree(req.Method); t != nil {
		if node := t.Lookup(req.URL.Path); node != nil && node.Handler != nil {
			node.Handler.ServeHTTP(rw, req)
			return
		}
	}
	mux.notFound.ServeHTTP(rw, req)
}
