package internalApi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Ishan27g/go-utils/tracing"
	"github.com/Ishan27g/internalApi/request"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

/*
Named APIs
- named according to the service's http.handler
- should match operationId (Openapi) spec or request name (postman)
*/
//
//type NamedApi interface {
//
//	// Add a named-NamedApi endpoint.
//	Add(apiName string, host string)
//
//	// ExpectHook marks the NamedApi endpoint as a mock hook
//	ExpectHook(apiName string, handler HttpHook)
//
//	// http methods
//	Get(apiName string, req *request.Request) ([]byte, error)
//	Post(apiName string, req *request.Request) ([]byte, error)
//	Put(apiName string, req *request.Request) ([]byte, error)
//	Patch(apiName string, req *request.Request) ([]byte, error)
//	Delete(apiName string, req *request.Request) ([]byte, error)
//}

type NamedApi struct {
	version         string
	tracingProvider tracing.TraceProvider
	httpClient      *http.Client
	defaultHeaders  map[string]string

	hosts     host                  // host: apiNames...
	apiClient map[string]*apiClient // host : apiClient{apiNames...}

}
type host map[string][]string

// get host for some apiName
func (h *host) getHost(apiName string) string {
	for host, apiNames := range *h {
		for _, name := range apiNames {
			if name == apiName {
				return host
			}
		}
	}
	return ""
}

func NewNamed(version string, options ...Option) *NamedApi {
	api := &NamedApi{
		version:         version,
		tracingProvider: nil,
		httpClient:      nil,

		hosts:     host{},
		apiClient: map[string]*apiClient{},

		defaultHeaders: map[string]string{
			//"X-CUSTOM-HEADER": "custom",
			//"Accept":          "application/json",
		},
	}

	for _, option := range options {
		option(api)
	}

	if api.httpClient == nil {
		WithDefaultHttpClient()(api)
	}

	return api
}

func (na *NamedApi) Add(apiName string, host string) {

	// add apiName to host
	if na.hosts[host] == nil {
		na.hosts[host] = []string{}
	}
	na.hosts[host] = append(na.hosts[host], apiName)

	// add NamedApi client for this host
	if na.apiClient[host] == nil {
		na.apiClient[host] = &apiClient{
			tracingProvider: na.tracingProvider,
			httpClient:      *na.httpClient,
			hooks:           map[string]*HttpHook{},
		}
	}

}
func (na *NamedApi) ExpectHook(apiName string, handler HttpHook) {

	host := na.hosts.getHost(apiName)
	if host == "" {
		panic("expecting an api hook which was not added " + apiName)
	}

	if na.apiClient[host].hooks[apiName] != nil {
		panic("already added hook " + apiName)
	}

	na.apiClient[host].hooks[apiName] = &handler

}

func (na *NamedApi) Get(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodGet, req)
}
func (na *NamedApi) Post(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodPost, req)
}
func (na *NamedApi) Put(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodPut, req)
}
func (na *NamedApi) Patch(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodPatch, req)
}
func (na *NamedApi) Delete(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodDelete, req)
}
func (na *NamedApi) Options(apiName string, req *request.Request) ([]byte, error) {
	return na.do(apiName, http.MethodOptions, req)
}

// does bulk of http request preparation before calling the client for this host
func (na *NamedApi) do(apiName string, method string, r *request.Request) (rsp []byte, err error) {

	host := na.hosts.getHost(apiName)
	if na.apiClient[host] == nil {
		return nil, errors.New(fmt.Sprintf("%s NamedApi not added", apiName))
	}

	var (
		doRequest *http.Request
		ctx       context.Context
		span      = trace.SpanFromContext(r.Request.Context())
	)

	if na.tracingProvider != nil {
		_, span = na.tracingProvider.Get().Start(r.Request.Context(), apiName)
	}
	defer span.End()

	ctx = trace.ContextWithSpan(r.Request.Context(), span)

	doRequest, err = http.NewRequestWithContext(ctx, method, fmt.Sprintf("http://%s/%s%s", host, na.version, r.Request.URL.String()), r.Request.Body)
	if err != nil {
		return nil, err
	}

	// copy optional headers
	for key, val := range r.Request.Header {
		doRequest.Header.Set(key, val[0])
		for _, v := range val[1:] {
			doRequest.Header.Add(key, v)
		}
	}

	// add default headers
	for key, val := range na.defaultHeaders {
		if doRequest.Header.Get(key) == "" {
			doRequest.Header.Set(key, val)
			span.SetAttributes(attribute.Key("internal-header" + key).String(val))
		}
	}

	// post,put,patch
	if strings.Contains(doRequest.Method, "P") && doRequest.Header.Get("Content-Type") == "" {
		doRequest.Header.Set("Content-Type", "application/json")
	}

	log.WithFields(log.Fields{
		"url":    doRequest.URL.String(),
		"method": doRequest.Method,
	}).Debug("sending request")

	return na.apiClient[host].do(ctx, na.apiClient[host].hooks[apiName], doRequest)

}
