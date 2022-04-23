package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
)

var port = flag.String("http", ":9999", "http port")

// simulates data published to some queue
func publishToQueue(data string) bool {
	<-time.After(1 * time.Second)
	if data != "" {
		return true
	}
	return false
}

// genericMethod is any generic async method in a request pipeline
// It would normally be triggered by nats/kafka. After processing
// this method would ideally publish to nats/kafka to trigger the next
// service/stage of the request pipeline

/* Noop-Testing: Before the final publish operation ->
- add publish `msg` to the context
- return if ctx is noop, otherwise publish `msg` to queue
*/
func genericMethod(ctx context.Context, data string) bool {

	// data validation
	// ...
	// ...

	// define event to be added
	event := noop.Event{
		Name: "Async method triggered at service " + *port,
		Meta: data,
	}

	// add this data to actions if actions was previously added to this ctx
	// otherwise, nothing is added
	noop.ActionsFromCtx(ctx).AddEvent(event)

	// finally, return if noop operation
	if noop.ContainsNoop(ctx) {
		return true
	}
	// otherwise, act as normal , i.e. publish the message to trigger the next service
	return publishToQueue(data)
}

/* Noop-Testing:
Wrap async method as its own http handler, which based on the `noop` param
- extracts the action from request
- triggers underlying async method
- responds with actions taken by this method
*/

func triggerAsyncFromHttp(c *gin.Context) {
	dataFromRequest := c.Query("data")

	// is request a `noop`
	isNoop := strings.EqualFold(c.Query("noop"), "true")

	ctx := noop.NewCtxWithNoop(c, isNoop) // todo new ctx or gin-request context?

	// actions for this request
	actions := noop.NewActions()

	// add an event
	// actions.AddEvent(noop.Event{Name: "Endpoint", Meta: c.Request.URL.RequestURI()})
	ctx = noop.NewCtxWithActions(ctx, actions)

	// manually trigger the async method
	if !genericMethod(ctx, dataFromRequest) {
		c.JSON(http.StatusExpectationFailed, nil)
	}

	if rsp, err := actions.Marshal(); err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(rsp)
		return
	}
	c.JSON(http.StatusExpectationFailed, nil)
}

func startServer() {

	flag.Parse()
	if *port == "" {
		return
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/endpoint1", triggerAsyncFromHttp)
	r.GET("/endpoint2", triggerAsyncFromHttp)
	r.GET("/endpoint3", triggerAsyncFromHttp)

	log.Println("starting on", *port)
	err := r.Run(*port)
	if err != nil {
		log.Fatalf(err.Error(), err)
	}
}
