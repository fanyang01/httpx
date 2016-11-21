package mux

import "net/http"

type option struct {
	StrictSlash      bool
	UseEncodedPath   bool
	CleanPath        bool
	NotFound         http.Handler
	MethodNotAllowed http.Handler
}

type Option func(*Mux)

func HandleNotFound(h http.Handler) Option {
	return func(mux *Mux) { mux.NotFound = h }
}

func HandleMethodNotAllowed(h http.Handler) Option {
	return func(mux *Mux) { mux.MethodNotAllowed = h }
}

func StrictSlash(value bool) Option {
	return func(mux *Mux) { mux.StrictSlash = value }
}

func UseEncodedPath(value bool) Option {
	return func(mux *Mux) { mux.UseEncodedPath = value }
}

func CleanPath(value bool) Option {
	return func(mux *Mux) { mux.CleanPath = value }
}

func MethodNotAllowed(rw http.ResponseWriter, req *http.Request) {
	code := http.StatusMethodNotAllowed
	http.Error(rw, http.StatusText(code), code)
}
