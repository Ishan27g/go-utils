package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
)

var port = flag.String("http", ":9999", "http port")

// simulates database operation
func dbCrud(data string) bool {
	<-time.After(1 * time.Second)
	if data != "" {
		return true
	}
	return false
}

// databaseCall is any generic async method that would normally be triggered by nats/kafka...
// Before the final database operation ->
// - adds event to the context
// - returns if ctx is noop, otherwise proceeds with db operation
func databaseCall(ctx context.Context, data string) bool {

	// generic data validation
	// ...
	// ...

	// define event to be added
	event := noop.Event{
		Name: "Database call at service " + *port,
		Meta: data,
	}

	// add this event to actions if actions was previously added to this ctx, otherwise
	// nothing is added
	noop.ActionsFromCtx(ctx).AddEvent(event)

	// finally, return if noop operation
	if noop.ContainsNoop(ctx) {
		return true
	}
	// otherwise, write to DB
	return dbCrud(data)
}

// manually trigger async methods via a http call
// the request is handled based on noop
func triggerAsyncFromHttp(c *gin.Context) {
	dataFromRequest := c.Query("data")

	// is request a `noop`
	isNoop := strings.EqualFold(c.Query("noop"), "true")

	fmt.Println(dataFromRequest, isNoop)

	ctx := noop.NewCtxWithNoop(c, isNoop) // todo new ctx or gin-request context?
	actions := noop.NewActions()

	// add an event
	// actions.AddEvent(noop.Event{Name: "Endpoint", Meta: c.Request.URL.RequestURI()})
	ctx = noop.NewCtxWithActions(ctx, actions)

	// manually trigger the async method
	if !databaseCall(ctx, dataFromRequest) {
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
