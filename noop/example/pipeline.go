package example

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"golang.org/x/sync/errgroup"
)

const (
	host        = "http://localhost"
	urlWithNoop = "?noop=true"
)

var isNoop = false

func buildUrl(service string) string {
	buildUrl := func() string {
		return host + service
	}
	if !isNoop {
		return buildUrl()
	}
	return buildUrl() + urlWithNoop
}
func SetNoop(as bool) {
	isNoop = as
}
func Setup() (*Service, *Service, *Service, *Service) {
	s1 := New("service-1", ":9991", 1)
	s2 := New("service-2", ":9992", 2)
	s3 := New("service-3", ":9993", 3)
	s4 := New("service-4", ":9994", 4)
	return s1, s2, s3, s4
}
func sendHttpReq(ctx context.Context, to string, data string) []noop.Event {
	payload, err := json.Marshal(d{Data: data})
	if err != nil {
		return nil
	}
	request, err := http.NewRequestWithContext(ctx, "POST", to, bytes.NewReader(payload))
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
		events = append(events, noop.Event{Name: e["name"].(string),
			Meta: e["meta"], NextSubject: e["nextSubject"].(string)})
	}
	return events
}
func sendRequestWithActions(actions *noop.Actions, service, data string) bool {
	ctx := noop.NewCtxWithActions(context.Background(), actions)
	rsp := sendHttpReq(ctx, buildUrl(service), data)
	if rsp != nil {
		actions.AddEvent(rsp...)
		return true
	}
	return false
}

func TriggerAsync(s1, s2, s3, s4 *Service) {
	g, _ := errgroup.WithContext(context.Background())

	// subscribe for all stages except the first one
	g.Go(func() error {
		NatsSubscribe(s4.subjForThis(), s4.GenericMethod)
		return nil
	})
	g.Go(func() error {
		NatsSubscribe(s3.subjForThis(), s3.GenericMethod)
		return nil
	})
	g.Go(func() error {
		NatsSubscribe(s2.subjForThis(), s2.GenericMethod)
		return nil
	})
	_ = g.Wait()

	// publish from first stage
	g.Go(func() error {
		NatsPublish(s1.subjForNext(), "pipeline-data", nil)
		return nil
	})
	_ = g.Wait()

}

// SendHttpNoop sends http to individually trigger each service in order
func SendHttpNoop(s1, s2, s3, s4 *Service, triggerAll bool) {

	data := "pipeline-data"
	endpoint := "/endpoint"

	actions := noop.NewActions()
	if !sendRequestWithActions(actions, s1.Port+endpoint, data) {
		fmt.Println("error at " + s1.Name)
		return
	}
	if triggerAll {
		if !sendRequestWithActions(actions, s2.Port+endpoint, data) {
			fmt.Println("error at " + s2.Name)
			return
		}

		if !sendRequestWithActions(actions, s3.Port+endpoint, data) {
			fmt.Println("error at " + s3.Name)
			return
		}

		if !sendRequestWithActions(actions, s4.Port+endpoint, data) {
			fmt.Println("error at " + s4.Name)
			return
		}
	}

	fmt.Println("All actions:")
	for _, event := range actions.GetEvents() {
		fmt.Println(fmt.Sprintf("\t- [%s]:[%v][%v] ", event.Name, event.NextSubject, event.Meta))
	}

}
