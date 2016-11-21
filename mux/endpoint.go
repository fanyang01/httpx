package mux

import (
	"net/http"

	"github.com/fanyang01/httpx/internal/radix"
)

type endpoint struct {
	method      string
	middlewares []Middleware
	handler     http.Handler
	combined    *radix.Node
}

func (mux *Mux) record(method string, p string,
	h http.Handler, node *radix.Node, ms []Middleware) {

	cn := mux.combined.Add(p, mux.updateCombined)
	cn.Replace(mux.MethodNotAllowed.ServeHTTP)
	mux.link[cn] = append(mux.link[cn], node)
	mux.endpoint[node] = &endpoint{
		method:      method,
		handler:     h,
		combined:    cn,
		middlewares: ms,
	}
}

func (mux *Mux) updateEndpoint(old, new *radix.Node) {
	p := mux.endpoint[old]
	delete(mux.endpoint, old)
	mux.endpoint[new] = p
	cn := p.combined
	for i, n := range mux.link[cn] {
		if n == old {
			mux.link[cn][i] = new
			break
		}
	}
}

func (mux *Mux) updateCombined(old, new *radix.Node) {
	nodes := mux.link[old]
	delete(mux.link, old)
	mux.link[new] = nodes
	for _, n := range nodes {
		mux.endpoint[n].combined = new
	}
}
