package jiji

import (
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type MockHandler struct {
	StatusCode int
}

func (t *MockHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	b, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	user, pass, ok := req.BasicAuth()
	if ok {
		Logger.Debug("MockHandler", "user", user, "pass", pass, "body", string(b))
	} else {
		Logger.Debug("MockHandler", "body", string(b))
	}
	Logger.Debug("MockHandler", "StatusCode", t.StatusCode)
	http.Error(w, http.StatusText(t.StatusCode), t.StatusCode)
}

func TestHTTP(t *testing.T) {
	handler := MockHandler{StatusCode: http.StatusOK}
	srv := httptest.NewServer(&handler)
	defer srv.Close()
	u, _ := url.Parse(srv.URL)
	u.User = url.UserPassword("user", "pass")
	transport := &HTTP{
		Client: http.DefaultClient,
		URL:    u,
	}
	var d time.Duration
	done := make(chan struct{})
	n := 10
	for i := range n {
		d += time.Duration(rand.Int64N(50)+50) * time.Millisecond
		mod := i % 3
		switch {
		case mod == 2:
			time.AfterFunc(d, func() {
				handler.StatusCode = http.StatusInternalServerError
			})
		case mod == 1:
			time.AfterFunc(d, func() {
				handler.StatusCode = http.StatusBadRequest
			})
		case i == n-1: // last one
			time.AfterFunc(d, func() {
				handler.StatusCode = http.StatusOK
				close(done)
			})
		default:
			time.AfterFunc(d, func() {
				handler.StatusCode = http.StatusOK
			})
		}
	}
	testDelivery(t, transport, n)
	<-done
}
