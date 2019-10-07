package jiji

import (
	"bytes"
	"errors"
	"net/http"
)

type HTTP struct {
	Client   *http.Client
	URL      string
	User     string // For basic auth. Ignored if empty.
	Password string
}

const (
	ContentTypeJSON = "application/json"
)

func (t *HTTP) Connect() (err error) {
	return nil
}

func (t *HTTP) Close() {
	t.Client.CloseIdleConnections()
}

func (t *HTTP) Send(b []byte) (err error) {
	var res *http.Response
	body := bytes.NewReader(b)
	if t.User == "" {
		res, err = t.Client.Post(t.URL, ContentTypeJSON, body)
		if err != nil {
			return
		}
	} else {
		var req *http.Request
		req, err = http.NewRequest(http.MethodPost, t.URL, body)
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", ContentTypeJSON)
		req.SetBasicAuth(t.User, t.Password)
		res, err = t.Client.Do(req)
		if err != nil {
			return
		}
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return
	}
	return errors.New(res.Status)
}
