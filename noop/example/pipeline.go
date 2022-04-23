package example

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const (
	host        = "http://localhost:"
	urlWithNoop = "?noop=true"
)

var isNoop = false

func timeIt(from time.Time) {
	fmt.Println(fmt.Sprintf("\nisNoop:%v took:%v", isNoop, time.Since(from)))
}
func buildUrl(service string) string {
	buildUrl := func() string {
		return host + service
	}
	if !isNoop {
		return buildUrl()
	}
	return buildUrl() + urlWithNoop
}

func setup() (*Service, *Service, *Service, *Service) {
	s1 := New("service-1", 1)
	s2 := New("service-2", 2)
	s3 := New("service-3", 3)
	s4 := New("service-4", 4)
	return s1, s2, s3, s4
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
func RunAsync() {
	s1, s2, s3, s4 := setup()

	g, _ := errgroup.WithContext(context.Background())

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

	<-time.After(1 * time.Second)

	g.Go(func() error {
		NatsPublish(s1.subjForNext(), "pipeline-data", nil)
		return nil
	})
	_ = g.Wait()
}

func sendRequestWithActions(actions *noop.Actions, service, data string) {
	ctx := noop.NewCtxWithActions(context.Background(), actions)
	rsp := sendHttpReq(ctx, buildUrl(service))
	actions.AddEvent(rsp...)
}

func startServer(service *Service, port string) {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/endpoint", service.TriggerAsyncFromHttp)

	log.Println("starting on", port)
	err := r.Run(port)
	if err != nil {
		log.Fatalf(err.Error(), err)
	}
}

func RunSync() {
	data := "pipeline-data"
	endpoint := "endpoint"

	s1, s2, s3, s4 := setup()

	go startServer(s1, ":9991")
	go startServer(s2, ":9992")
	go startServer(s3, ":9993")
	go startServer(s4, ":9994")

	actions := noop.NewActions()
	actions.AddEvent(noop.Event{
		Name:        "From tester to service1",
		NextSubject: s1.subjForNext(),
		Meta:        data,
	})
	sendRequestWithActions(actions, "9991/"+endpoint, data)

	actions.AddEvent(noop.Event{
		Name:        "From tester to service2",
		NextSubject: s2.subjForNext(),
		Meta:        data,
	})
	sendRequestWithActions(actions, "9992/"+endpoint, data)

	actions.AddEvent(noop.Event{
		Name:        "From tester to service3",
		NextSubject: s3.subjForNext(),
		Meta:        data,
	})
	sendRequestWithActions(actions, "9993/"+endpoint, data)

	actions.AddEvent(noop.Event{
		Name:        "From tester to service4",
		NextSubject: s4.subjForNext(),
		Meta:        data,
	})
	sendRequestWithActions(actions, "9994/"+endpoint, data)

	for _, event := range actions.GetEvents() {
		fmt.Println(fmt.Sprintf("[%s]:%v", event.Name, event.Meta))
	}

}
