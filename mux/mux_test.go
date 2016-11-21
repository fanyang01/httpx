package mux_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fanyang01/httpx/mux"
)

type H struct {
	t       *testing.T
	method  string
	pattern string
	i       int
	path    string
}

func (h H) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if got, want := req.Method, h.method; got != want {
		h.t.Errorf("got http method %v, want %v", got, want)
	}
	fmt.Fprint(rw, h.i)
}

func TestMuxMethod(t *testing.T) {
	tests := []H{
		{t, "GET", "/", 0, "/"},
		{t, "POST", "/", 1, "/"},
		{t, "PUT", "/", 2, "/"},
		{t, "DELETE", "/", 3, "/"},
		{t, "CONNECT", "/", 4, "/"},
		{t, "OPTIONS", "/", 5, "/"},
		{t, "TRACE", "/", 6, "/"},
		{t, "PATCH", "/", 7, "/"},
		{t, "DIY", "/", 8, "/"},
	}
	mux := mux.New()
	for _, tt := range tests {
		mux.Handle(tt.method, tt.pattern, tt)
	}
	server := httptest.NewServer(mux)
	defer server.Close()

	for _, tt := range tests {
		client := http.DefaultClient
		t.Run(tt.method, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, server.URL+tt.path, nil)
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			if got, want := string(body), strconv.Itoa(tt.i); got != want {
				t.Errorf("got response body %v, want %v", got, want)
			}
		})
	}
}
