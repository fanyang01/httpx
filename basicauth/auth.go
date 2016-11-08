package basicauth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	HeaderProxyAuthorization = "Proxy-Authorization"
	HeaderProxyAuthenticate  = "Proxy-Authenticate"
	HeaderAuthorization      = "Authorization"
	HeaderWWWAuthenticate    = "WWW-Authenticate"
)

type contextKey int

const (
	UserContextKey      = contextKey(0)
	ProxyUserContextKey = contextKey(1)
)

type Config struct {
	Auth  func(username, password string) bool
	Realm string
}

func auth(config *Config, chdr, shdr string, code int, ck contextKey) func(http.Handler) http.Handler {
	if config == nil {
		panic("basicauth: the config parameter can't be nil")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			header := req.Header.Get(chdr)
			defer req.Header.Del(chdr)

			username, password, ok := Decode(header)
			if !ok || !config.Auth(username, password) {
				rw.Header().Set(shdr, `Basic realm="`+config.Realm+`"`)
				rw.WriteHeader(code)
				return
			}

			ctx := context.WithValue(req.Context(), ck, username)
			req.WithContext(ctx)

			next.ServeHTTP(rw, req)
			return
		})
	}
}

func AuthProxy(config *Config) func(http.Handler) http.Handler {
	return auth(config, HeaderProxyAuthorization, HeaderProxyAuthenticate, http.StatusProxyAuthRequired, ProxyUserContextKey)
}

func Auth(config *Config) func(http.Handler) http.Handler {
	return auth(config, HeaderAuthorization, HeaderWWWAuthenticate, http.StatusUnauthorized, UserContextKey)
}

func Encode(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func Decode(s string) (username, password string, ok bool) {
	ss := strings.Fields(s)
	if len(ss) != 2 || ss[0] != "Basic" {
		return
	}
	b, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		return
	}
	if ss = strings.SplitN(string(b), ":", 2); len(ss) != 2 {
		return
	}
	return ss[0], ss[1], true
}
