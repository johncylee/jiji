package jiji

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func MockHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	user, pass, ok := req.BasicAuth()
	if ok {
		Logger.Info("MockHTTPServ.BasicAuth", "user", user, "pass", pass)
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	Logger.Info("MockHTTPServ", "req.Body", string(b))
	w.Write([]byte("OK"))
}

func TestHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(MockHandler))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	transport := &HTTP{
		Client: http.DefaultClient,
		URL:    u,
	}
	testDelivery(t, transport, 5)
	u.User = url.UserPassword("user", "pass")
	testDelivery(t, transport, 5)
}
