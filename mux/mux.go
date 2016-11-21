package mux

import (
	"fmt"
	"net/http"
	"path"
	"strings"

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
	hmap        hmap
	pathfunc    func(*http.Request) string
	middlewares []Middleware
	endpoint    map[*radix.Node]*endpoint
	combined    radix.Tree
	link        map[*radix.Node][]*radix.Node
	extended    map[string]*radix.Tree
	option
}

func New(options ...Option) *Mux {
	mux := Mux{
		endpoint: make(map[*radix.Node]*endpoint),
		link:     make(map[*radix.Node][]*radix.Node),
		extended: make(map[string]*radix.Tree),
		option: option{
			StrictSlash:      true,
			UseEncodedPath:   false,
			CleanPath:        false,
			NotFound:         http.HandlerFunc(http.NotFound),
			MethodNotAllowed: http.HandlerFunc(MethodNotAllowed),
		},
	}
	mux.hmap.Add(
		xGET, xPOST, xPUT, xHEAD, xDELETE,
		xOPTIONS, xPATCH, xTRACE, xCONNECT,
	)
	for _, f := range options {
		f(&mux)
	}
	switch {
	case mux.UseEncodedPath && mux.CleanPath:
		mux.pathfunc = cleanEncodedPath
	case mux.UseEncodedPath:
		mux.pathfunc = encodedPath
	case mux.CleanPath:
		mux.pathfunc = cleanPath
	default:
		mux.pathfunc = urlPath
	}
	return &mux
}

func urlPath(req *http.Request) string     { return req.URL.Path }
func encodedPath(req *http.Request) string { return req.URL.EscapedPath() }
func cleanPath(req *http.Request) string   { return path.Clean(req.URL.Path) }
func cleanEncodedPath(req *http.Request) string {
	return path.Clean(req.URL.EscapedPath())
}

func (mux *Mux) replace(node *radix.Node, f http.HandlerFunc) bool {
	_, replaced := node.Replace(f)
	return replaced
}

func (mux *Mux) add(method, pattern string, h http.Handler, middlewares ...Middleware) {
	t := mux.tree(method)
	if t == nil {
		t = &radix.Tree{}
		mux.extended[method] = t
	}

	var (
		node    = t.Add(pattern, mux.updateEndpoint)
		handler = h
		mws     = make([]Middleware, 0, len(mux.middlewares)+len(middlewares))
	)
	mws = append(mws, mux.middlewares...)
	mws = append(mws, middlewares...)

	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i].Wrap(handler)
	}
	if replaced := mux.replace(node, handler.ServeHTTP); replaced {
		panic(fmt.Errorf(
			"mux: can't override registered pattern: %s %q",
			method, pattern,
		))
	}
	mux.record(method, pattern, h, node, mws)

	if mux.StrictSlash && node.Type() != radix.MatchAllNode {
		mux.redirect(t, pattern)
	}

}

func (mux *Mux) redirect(t *radix.Tree, pattern string) {
	var f func(string) string

	if strings.HasSuffix(pattern, "/") {
		pattern = pattern[:len(pattern)-1]
		f = func(s string) string { return s + "/" }
	} else {
		pattern = pattern + "/"
		f = func(s string) string { return s[:len(s)-1] }
	}
	if node := t.Lookup(pattern); node == nil || node.HandlerFunc == nil {
		t.Add(pattern, mux.updateEndpoint).Replace(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(
					w, r, f(r.URL.String()), http.StatusMovedPermanently,
				)
			},
		)
	}
}

func (mux *Mux) Use(middlewares ...Middleware) {
	mux.middlewares = append(mux.middlewares, middlewares...)
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
		if node := t.Lookup(mux.pathfunc(req)); node != nil && node.HandlerFunc != nil {
			node.HandlerFunc(rw, req)
			return
		}
		mux.NotFound.ServeHTTP(rw, req)
		return
	}
	mux.MethodNotAllowed.ServeHTTP(rw, req)
}
