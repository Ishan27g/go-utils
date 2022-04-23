package main

import (
	"context"
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
)

const (
	host        = "http://localhost:"
	data        = "?data="
	urlWithNoop = "&noop=true"
)

var isNoop = false

func timeIt(from time.Time) {
	fmt.Println("took", time.Since(from))
}
func buildUrl(service, d string) string {
	buildUrl := func() string {
		return host + service + data + d
	}
	if !isNoop {
		return buildUrl()
	}
	return buildUrl() + urlWithNoop
}

func sendHttpReq(ctx context.Context, to string) []noop.Event {
	request, err := http.NewRequestWithContext(ctx, "GET", to, nil)
	if err != nil {
		return nil
	}
	client := http.Client{
		Timeout: 6 * time.Second,
	}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error() + " ok")
		return nil
	}
	rsp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var actions noop.Json
	err = json.Unmarshal(rsp, &actions)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var events []noop.Event
	for _, event := range actions.Events {
		e := event.(map[string]interface{})
		events = append(events, noop.Event{Name: e["name"].(string), Meta: e["meta"]})
	}
	return events
}
func sendRequestWithActions(actions *noop.Actions, service, data string) {
	ctx := noop.NewCtxWithActions(context.Background(), actions)
	rsp := sendHttpReq(ctx, buildUrl(service, data))
	actions.AddEvent(rsp...)
}
func TestRequestPipeline() *noop.Actions {
	defer timeIt(time.Now())
	service1 := "9999/endpoint1"
	service2 := "9999/endpoint2"
	service3 := "9999/endpoint3"
	meta := func(to string) interface{} {
		return "This replaces the async call that would trigger " + to
	}

	actions := noop.NewActions()
	actions.AddEvent(noop.Event{
		Name: "From tester to service1",
		Meta: meta("service1"),
	})

	sendRequestWithActions(actions, service1, "data1")

	actions.AddEvent(noop.Event{
		Name: "From tester to service2",
		Meta: meta("service2"),
	})
	sendRequestWithActions(actions, service2, "data2")

	actions.AddEvent(noop.Event{
		Name: "From tester to service3",
		Meta: meta("service3"),
	})
	sendRequestWithActions(actions, service3, "data3")

	return actions
}

func main() {
	// mock three services as a single server
	go func() {
		startServer()
	}()

	// Test normal requests to service's , i.e. without `noop` (noop=false)
	// Each service will do a DB operation
	isNoop = false
	actionsWithoutNoop := TestRequestPipeline()
	fmt.Println("noop = false")
	for _, event := range actionsWithoutNoop.GetEvents() {
		fmt.Println(fmt.Sprintf("[%s]:%v", event.Name, event.Meta))
	}

	// Test same service's with `noop` requests (noop=true)
	// Each service behave as before except returning before DB operations
	isNoop = true
	actionsWithNoop := TestRequestPipeline()

	fmt.Println("\nnoop = true")
	for _, event := range actionsWithNoop.GetEvents() {
		fmt.Println(fmt.Sprintf("[%s]:%v", event.Name, event.Meta))
	}
	fmt.Println()
}
