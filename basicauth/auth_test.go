package basicauth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		username string
		password string
		want     string
	}{
		{"user", "password", "Basic dXNlcjpwYXNzd29yZA=="},
	}
	for _, tt := range tests {
		t.Run(tt.username+":"+tt.password, func(t *testing.T) {
			if got := Encode(tt.username, tt.password); got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		wantUsername string
		wantPassword string
		wantOk       bool
	}{
		{"empty string", "", "", "", false},
		{"invalid header 1", "Basic", "", "", false},
		{"invalid header 2", "Basic abcd", "", "", false},
		{"invalid header 3", "Bearer dXNlcjpwYXNzd29yZA==", "", "", false},
		{"normal", "Basic dXNlcjpwYXNzd29yZA==", "user", "password", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsername, gotPassword, gotOk := Decode(tt.s)
			if gotUsername != tt.wantUsername {
				t.Errorf("Decode() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("Decode() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Decode() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestAuth(t *testing.T) {
	config := &Config{
		Realm: "Basic Auth",
		Auth: func(username, password string) bool {
			return username == "foo" && password == "bar"
		},
	}
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, r.Context().Value(UserContextKey))
	})
	server := httptest.NewServer(Auth(config)(handler))
	defer server.Close()

	tests := []struct {
		name       string
		headerName string
		user       string
		password   string
		wantStatus int
	}{
		{"without header", "", "", "", 401},
		{"proxy authorization header", HeaderProxyAuthorization, "foo", "bar", 401},
		{"wrong password", HeaderAuthorization, "foo", "baz", 401},
		{"normal", HeaderAuthorization, "foo", "bar", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL, nil)
			if tt.headerName != "" {
				req.Header.Add(tt.headerName, Encode(tt.user, tt.password))
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
			switch resp.StatusCode {
			case 401:
				if want, got := `Basic realm="Basic Auth"`, resp.Header.Get(HeaderWWWAuthenticate); got != want {
					t.Errorf("got header = %v, want %v", got, want)
				}
			case 200:
				b, _ := ioutil.ReadAll(resp.Body)
				if got, want := string(b), tt.user+"\n"; got != want {
					t.Errorf("got body = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestAuthProxy(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, "hello")
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	srvURL, _ := url.Parse(server.URL)

	config := &Config{
		Realm: "HTTP Proxy",
		Auth: func(username, password string) bool {
			return username == "foo" && password == "bar"
		},
	}
	proxy := httptest.NewServer(
		AuthProxy(config)(httputil.NewSingleHostReverseProxy(srvURL)),
	)
	defer proxy.Close()
	proxyURL, _ := url.Parse(proxy.URL)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	tests := []struct {
		name       string
		headerName string
		user       string
		password   string
		wantStatus int
	}{
		{"without header", "", "", "", 407},
		{"authorization header", HeaderAuthorization, "foo", "bar", 407},
		{"wrong password", HeaderProxyAuthorization, "foo", "baz", 407},
		{"normal", HeaderProxyAuthorization, "foo", "bar", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", server.URL, nil)
			if tt.headerName != "" {
				req.Header.Add(tt.headerName, Encode(tt.user, tt.password))
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("got status = %v, want %v", resp.StatusCode, tt.wantStatus)
			}
			switch resp.StatusCode {
			case 407:
				if want, got := `Basic realm="HTTP Proxy"`, resp.Header.Get(HeaderProxyAuthenticate); got != want {
					t.Errorf("got header = %v, want %v", got, want)
				}
			case 200:
				b, _ := ioutil.ReadAll(resp.Body)
				if got, want := string(b), "hello\n"; got != want {
					t.Errorf("got body = %v, want %v", got, want)
				}
			}
		})
	}
}
