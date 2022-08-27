package internalApi

import (
	"net/http"
	"time"

	"github.com/Ishan27g/go-utils/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Option func(*NamedApi)

func WithDefaultHttpClient() Option {
	return func(api *NamedApi) {
		api.httpClient = &http.Client{Timeout: 1 * time.Minute}
	}
}
func WithHttpClient(httpClient http.Client) Option {
	return func(api *NamedApi) {
		api.httpClient = &httpClient
	}
}
func WithTracingHttpClient() Option {
	return func(api *NamedApi) {
		api.httpClient = &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport), Timeout: 1 * time.Minute}
	}
}
func WithTracingProvider(provider tracing.TraceProvider) Option {
	return func(api *NamedApi) {
		api.tracingProvider = provider
	}
}
