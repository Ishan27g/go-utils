package request

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// Request is wrapper for internal *http.Request
// Caller can access and modify Request.(*http.Request) before making the http call
type Request struct {
	user *struct {
		username string
		password string
	}
	query   map[string]string
	headers map[string][]string
	*http.Request
}
type Option func(*Request)

func WithHeaders(headers map[string][]string) Option {
	return func(request *Request) {
		request.headers = headers
	}
}
func WithQueryParams(query map[string]string) Option {
	return func(request *Request) {
		request.query = query
	}
}
func WithAuthBasic(username, password string) Option {
	return func(request *Request) {
		request.user = &struct {
			username string
			password string
		}{username: username, password: password}
	}
}

func NewRequest(ctx context.Context, urlEndpoint string, body io.Reader, o ...Option) (*Request, error) {
	var (
		err error
		r   = &Request{
			user:    nil,
			headers: map[string][]string{},
			query:   map[string]string{},
			Request: nil,
		}
	)

	for _, option := range o {
		option(r)
	}

	r.Request, err = http.NewRequestWithContext(ctx, "", urlEndpoint, body)
	if err != nil {
		return nil, err
	}

	// optional headers
	for key, val := range r.headers {
		if r.Header.Get(key) == "" {
			r.Header.Set(key, val[0])
			for _, v := range val[1:] {
				r.Header.Add(key, v)
			}
		}
	}

	// optional query params
	if len(r.query) > 0 {
		q, _ := url.ParseQuery(r.URL.RawQuery)
		for k, v := range r.query {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}

	// basic auth
	if r.user != nil {
		r.SetBasicAuth(r.user.username, r.user.password)
	}

	return r, err

}
