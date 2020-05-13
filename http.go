package jiji

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
)

type HTTP struct {
	Client   *http.Client
	URL      string
	user     string
	password string
}

const (
	ContentTypeJSON = "application/json"
)

func (t *HTTP) Connect() (err error) {
	url, err := url.Parse(t.URL)
	if err != nil {
		return
	}
	if url.User != nil {
		if password, set := url.User.Password(); set {
			t.user = url.User.Username()
			t.password = password
		}
	}
	return
}

func (t *HTTP) Close() {
	t.Client.CloseIdleConnections()
}

func (t *HTTP) Send(b []byte) (err error) {
	var res *http.Response
	body := bytes.NewReader(b)
	if t.user == "" {
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
		req.SetBasicAuth(t.user, t.password)
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
