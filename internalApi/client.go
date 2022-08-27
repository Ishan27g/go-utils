package internalApi

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/Ishan27g/go-utils/tracing"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

type HttpHook func(w http.ResponseWriter, r *http.Request)

type apiClient struct {
	httpClient http.Client

	hooks           map[string]*HttpHook // apiName:HttpHook
	tracingProvider tracing.TraceProvider
}

func (c *apiClient) do(ctx context.Context, hook *HttpHook, req *http.Request) (rsp []byte, err error) {

	var resp *http.Response
	var hooked = false

	span := trace.SpanFromContext(ctx)

	span.SetAttributes(semconv.HTTPRouteKey.String(req.URL.Path))
	span.SetAttributes(semconv.HTTPMethodKey.String(req.Method))

	var logFields = func() log.Fields {
		return log.Fields{"url": req.URL.String(), "method": req.Method}
	}

	// if hooks is set
	if hook != nil {
		hooked = true
		rr := httptest.NewRecorder()
		rr.Code = -1
		(*hook)(rr, req)
		if rr.Code != -1 {
			resp = rr.Result()
		}
		log.WithFields(logFields()).Debug("hooked")
	}

	// if no hooks is set, do actual http call
	if resp == nil {
		resp, err = c.httpClient.Do(req)
	}

	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("error sending request : %v", err.Error()))
		log.WithFields(logFields()).Error(fmt.Sprintf("error sending request : %v", err.Error()))
		return nil, err
	}

	if resp.StatusCode > http.StatusAccepted {
		log.WithFields(logFields()).Debug(fmt.Sprintf("Bad response status : %s", resp.Status))
		return nil, errors.New(fmt.Sprintf("Bad response status : %s", resp.Status))
	}

	rsp, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("error reading response [%s] : %v", resp.Status, err.Error()))
		log.WithFields(logFields()).Error(fmt.Sprintf("error reading response : %v", err.Error()))
		return nil, err
	}

	span.SetAttributes(attribute.Key("api-hooked?").Bool(hooked))
	span.SetAttributes(semconv.HTTPStatusCodeKey.Int(resp.StatusCode))
	span.SetAttributes(semconv.HTTPResponseContentLengthKey.Int64(resp.ContentLength))

	log.WithFields(logFields()).Debug(fmt.Sprintf("response [%s] - %s", resp.Status, string(rsp)))

	return

}
