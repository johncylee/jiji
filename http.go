package jiji

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
)

// POST to HTTP.URL with Content-Type "application/json". Supports HTTP basic auth via *url.URL.
// If HTTP.Drop is true, it will drop messages except network or http server errors.
type HTTP struct {
	Client *http.Client
	URL    *url.URL
	// Drop messages except network or server errors
	Drop bool
}

const (
	ContentTypeJSON = "application/json"
)

// Connect is a dummy function in case of HTTP
func (t *HTTP) Connect(ctx context.Context) (err error) {
	return nil
}

func (t *HTTP) Close() {
}

// Send will pace itself with Retry in case of error because HTTP.Connect will always return early.
func (t *HTTP) Send(b []byte, ctx context.Context) (err error) {
	body := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.URL.String(), body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", ContentTypeJSON)
	res, err := t.Client.Do(req)
	if err != nil {
		<-ctx.Done()
		return
	}
	// err == nil
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return // normal exit
	}
	<-ctx.Done()
	if t.Drop && res.StatusCode < http.StatusInternalServerError {
		Logger.Error("HTTP.Send", "dropped", res.Status)
		return
	}
	return errors.New(res.Status)
}
