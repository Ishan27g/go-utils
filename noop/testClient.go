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
	fmt.Println(fmt.Sprintf("\nisNoop:%v took:%v", isNoop, time.Since(from)))
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

// TestRequestPipeline
/* Noop-Testing: --> manually trigger the async methods across the pipeline
- http trigger returns actions for that service
- send requests with `actions` iteratively to each service in the pipeline
- each service returns its actions

when `noop=true` -> services return before publishing, client triggers each service
when `noop=false` -> services act as normal and publish for the next service
*/
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

	// Test the pipeline as is -> Async
	// Test a normal request -> trigger first service in pipeline without `noop` i.e. noop=false
	// The first service would behave as normal and trigger next service in the pipeline
	// All subsequent services in the pipeline should get triggered
	isNoop = false

	// todo send only to 1
	//actionsWithoutNoop := TestRequestPipeline()
	//for _, event := range actionsWithoutNoop.GetEvents() {
	//	fmt.Println(fmt.Sprintf("[%s]:%v", event.Name, event.Meta))
	//}

	// Noop-Testing: Test the pipeline synchronously
	// Test the same service's with `noop` requests i.e. noop=true.
	// Each service behave as before except returning before publish operation
	// Manually trigger each service based in the pipeline
	isNoop = true
	actionsWithNoop := TestRequestPipeline()

	for _, event := range actionsWithNoop.GetEvents() {
		fmt.Println(fmt.Sprintf("[%s]:%v", event.Name, event.Meta))
	}
	fmt.Println()
}
