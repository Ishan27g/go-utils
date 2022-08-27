package internalApi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/tracing"
	"github.com/Ishan27g/internalApi/request"
	"github.com/Ishan27g/internalApi/test/server"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type clientForX struct {
	*NamedApi
}

// AddUser via the client for the server
func (s *clientForX) AddUser(name string) bool {

	j, _ := json.Marshal(&server.User{Name: name})
	req, _ := request.NewRequest(context.Background(), "/users", bytes.NewReader(j), request.WithAuthBasic("any", "ok"), request.WithQueryParams(map[string]string{
		"user": name,
	}))
	_, err := s.NamedApi.Post(server.AddUserApi, req)

	return err == nil
}

// GetUser via the client for the server
func (s *clientForX) GetUser() *server.User {

	req, _ := request.NewRequest(context.Background(), "/users", nil, request.WithAuthBasic("any", "ok"))
	b, _ := s.NamedApi.Get(server.GetUserApi, req)

	var user = server.User{}
	err := json.Unmarshal(b, &user)
	if err != nil {
		fmt.Println(err.Error())
	}
	return &user
}

func Test_Api_AsHttp(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	provider := tracing.Init("jaeger", "clientForX", "users")
	defer provider.Close()
	// setup named api
	api := NewNamed(server.Version, WithTracingHttpClient(), WithTracingProvider(provider))
	api.Add(server.GetUserApi, "localhost:9999")
	api.Add(server.AddUserApi, "localhost:9999")

	client := &clientForX{api}

	// mock server
	go server.Run(nil) // comment this and run mock/server/server.go separately to see server spans

	<-time.After(1 * time.Second)

	// Send the actual http request
	// post
	assert.True(t, client.AddUser("user123"))

	// get
	user := client.GetUser()
	assert.Equal(t, "user123", user.Name)

}
func Test_Api_AsHook(t *testing.T) {

	log.SetLevel(log.DebugLevel)
	provider := tracing.Init("jaeger", "internalApi", "clientForX")
	defer provider.Close()

	// setup named api
	api := NewNamed(server.Version, WithTracingHttpClient(), WithTracingProvider(provider))
	api.Add(server.GetUserApi, "localhost:9999")
	api.Add(server.AddUserApi, "localhost:9999")

	client := &clientForX{api}

	// Set the hook
	api.ExpectHook(server.GetUserApi, func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(&server.User{Name: "mock-user"})
		if err != nil {
			return
		}
		w.Write(b)
	})
	api.ExpectHook(server.AddUserApi, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		return
	})

	// post
	assert.True(t, client.AddUser("mock-user"))

	//get
	user := client.GetUser()
	assert.Equal(t, "mock-user", user.Name)

}
