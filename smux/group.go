package smux

import (
	"net/http"
	"strings"
)

type Group struct {
	mux         *Mux
	prefix      string
	middlewares []Middleware
}

func (mux *Mux) Group(prefix string) *Group {
	return &Group{
		mux:    mux,
		prefix: prefix,
	}
}

type Middleware interface {
	Name() string
	Wrap(http.Handler) http.Handler
}

func (g *Group) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) Group(prefix string) *Group {
	return &Group{
		mux:         g.mux,
		prefix:      concat(g.prefix, prefix),
		middlewares: g.middlewares,
	}
}

func concat(prefix, s string) string {
	has0, has1 := strings.HasSuffix(prefix, "/"), strings.HasPrefix(s, "/")
	switch {
	case has0 && has1:
		return prefix + s[1:]
	case !has0 && !has1:
		return prefix + "/" + s
	}
	return prefix + s
}

func (g *Group) add(method, pattern string, h http.Handler) {
	pattern = concat(g.prefix, pattern)
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		h = g.middlewares[i].Wrap(h)
	}
	g.mux.add(method, pattern, h)
}

func (g *Group) Handle(m Method, pattern string, h http.Handler) {
	g.add(string(m), pattern, h)
}

func (g *Group) GET(pattern string, h http.Handler) {
	g.add(xGET, pattern, h)
}

func (g *Group) HEAD(pattern string, h http.Handler) {
	g.add(xHEAD, pattern, h)
}

func (g *Group) POST(pattern string, h http.Handler) {
	g.add(xPOST, pattern, h)
}

func (g *Group) PUT(pattern string, h http.Handler) {
	g.add(xPUT, pattern, h)
}

func (g *Group) DELETE(pattern string, h http.Handler) {
	g.add(xDELETE, pattern, h)
}
