package noop

import (
	"context"
	"net/http"
)

type keyType string

const noopKey keyType = "noop_key"

func ContainsNoop(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v := ctx.Value(noopKey); v != nil {
		return v.(bool)
	}
	return false
}

func NewCtxWithNoop(ctx context.Context, isNoop bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ContainsNoop(ctx) { // todo needed?
		return ctx
	}
	return context.WithValue(ctx, noopKey, isNoop)
}

func NewHttpRequest(req *http.Request) *http.Request {
	if ContainsNoop(req.Context()) {
		req.Header.Add(string(noopKey), "true")
	}
	r, _ := http.NewRequestWithContext(NewCtxWithNoop(req.Context(), true),
		req.Method, req.URL.String(), req.Body)
	for key, value := range req.Header {
		for _, v := range value {
			r.Header.Add(key, v)
		}
	}
	return r
}
func Middleware(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if isNoop := req.Header.Get(string(noopKey)); isNoop == "true" {
			req, _ = http.NewRequestWithContext(NewCtxWithNoop(req.Context(), true),
				req.Method, req.URL.String(), req.Body)
		}
		n.ServeHTTP(w, req)
	})
}
