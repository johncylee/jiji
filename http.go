package jiji

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
)

type HTTP struct {
	Client *http.Client
	URL    *url.URL
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

// Send POST to HTTP.URL with Content-Type "application/json". Supports HTTP basic auth and context.
func (t *HTTP) Send(b []byte, ctx context.Context) (err error) {
	body := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.URL.String(), body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", ContentTypeJSON)
	res, err := t.Client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return
	}
	if res.StatusCode < http.StatusInternalServerError {
		Logger.Error("http.Send", "status", res.Status)
		return
	}
	// only return server side error
	return errors.New(res.Status)
}
