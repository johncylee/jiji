package jiji

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
)

func MockHTTPServ(quit chan struct{}) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		user, pass, ok := req.BasicAuth()
		if ok {
			Logger.Printf("MockHTTPServ.BasicAuth: %s, %s",
				user, pass)
		}
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		Logger.Println("MockHTTPServ:", string(b))
		w.Write([]byte("OK"))
	})
	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		Logger.Println(srv.ListenAndServe())
	}()
	<-quit
	srv.Shutdown(context.Background())
}

func TestHTTP(t *testing.T) {
	quit := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		MockHTTPServ(quit)
	}()
	transport := HTTP{
		Client: http.DefaultClient,
		URL:    "http://localhost:8080/",
	}
	testDelivery(&transport, t)
	transport.User = "johndoe"
	transport.Password = "secret"
	testDelivery(&transport, t)
	close(quit)
	<-done
}
