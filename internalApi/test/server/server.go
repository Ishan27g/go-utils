package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Ishan27g/go-utils/tracing"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	Version    = "v1"
	GetUserApi = "GetUser"
	AddUserApi = "AddUser"
)

type User struct {
	Name string `json:"Name"`
}
type ServiceX struct {
	user            *User
	tracingProvider tracing.TraceProvider
}

// GetUser server handler
func (s *ServiceX) GetUser(w http.ResponseWriter, r *http.Request) {
	var span = trace.SpanFromContext(context.Background())
	if s.tracingProvider != nil {
		_, span = s.tracingProvider.Get().Start(r.Context(), GetUserApi)
		//span = trace.SpanFromContext(r.Context())
		//span.SetName(GetUserApi)
	}
	defer span.End()

	b, err := json.Marshal(s.user)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("error unmarshalling : %v", err.Error()))
		return
	}

	span.SetAttributes(attribute.String("userName", s.user.Name))
	span.SetStatus(codes.Ok, "returned user")

	w.Write(b)
	return
}

// AddUser server handler
func (s *ServiceX) AddUser(w http.ResponseWriter, r *http.Request) {
	var span = trace.SpanFromContext(context.Background())
	if s.tracingProvider != nil {
		_, span = s.tracingProvider.Get().Start(r.Context(), AddUserApi)
	}
	defer span.End()
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("error reading body : %v", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user = User{}
	err = json.Unmarshal(b, &user)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("error unmarshaling : %v", err.Error()))
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
	s.user = &user

	span.SetAttributes(attribute.String("userName", s.user.Name))
	span.SetStatus(codes.Ok, "added user")

	w.WriteHeader(http.StatusOK)
	return
}
func MBasicAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, _, ok := r.BasicAuth(); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}
func Run(provider tracing.TraceProvider) {

	server := &ServiceX{nil, provider}
	r := mux.NewRouter()
	r.Use(otelmux.Middleware("users"))

	v := r.PathPrefix("/" + Version).Subrouter()
	{
		v.HandleFunc("/users", server.GetUser).Methods("GET")
		v.HandleFunc("/users", server.AddUser).Methods("POST")
	}

	http.Handle("/", MBasicAuth(r))
	fmt.Println("listening on :9999")
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		return
	}
}
