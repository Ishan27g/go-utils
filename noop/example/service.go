package example

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

type Service struct {
	Name  string `json:"name"`
	Stage int    `json:"subj"`
	Port  string
}

func New(name string, port string, stage int) *Service {
	s := Service{
		Name:  name,
		Stage: stage,
		Port:  port,
	}
	defer startServer(&s, port)
	return &s
}
func (s *Service) TriggerAsyncFromHttp(c *gin.Context) {

	var data d
	if e := c.ShouldBindJSON(&data); e != nil {
		c.JSON(http.StatusExpectationFailed, nil)
		return
	}
	/* Noop-Testing:
	   Wrap async method as its own http handler, which based on the `noop` param
	   - extracts the action from request
	   - triggers underlying async method
	   - responds with actions taken by this method
	*/

	fmt.Println("from context - ", noop.ContainsNoop(c.Request.Context()))

	// get `noop=` from url
	isNoop := strings.EqualFold(c.Query("noop"), "true")

	ctx := noop.NewCtxWithNoop(context.Background(), isNoop) // todo new ctx or gin-request context?

	// actions for this request
	actions := noop.NewActions()

	// add an event
	actions.AddEvent(noop.Event{Name: "Endpoint hit at " + s.Name, Meta: c.Request.URL.RequestURI()})
	ctx = noop.NewCtxWithActions(ctx, actions)

	// manually trigger the async method
	if !s.GenericMethod(ctx, &nats.Msg{Data: []byte(data.Data)}) {
		c.JSON(http.StatusExpectationFailed, nil)
		return
	}

	if rsp, err := actions.Marshal(); err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		_, _ = c.Writer.Write(rsp)
		return
	}
	c.JSON(http.StatusExpectationFailed, nil)
}
func startServer(service *Service, port string) {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/endpoint", service.TriggerAsyncFromHttp)

	log.Println("starting on", port)
	go func() {
		err := r.Run(port)
		if err != nil {
			log.Fatalf(err.Error(), err)
		}
	}()
}

// GenericMethod is any generic async method in a request pipeline
// It would normally be triggered by nats/kafka. After processing
// this method would ideally publish to nats/kafka to trigger the next
// service/stage of the request pipeline
/* Noop-Testing: Before the final publish operation ->
- add publish `msg` to the context
- return if ctx is noop, otherwise publish `msg` to queue
*/
func (s *Service) GenericMethod(ctx context.Context, msg *nats.Msg) bool {

	// data validation
	// ...
	// ...
	<-time.After(1 * time.Second)

	// define event to be added
	event := noop.Event{
		Name:        "Async method triggered at " + s.Name,
		NextSubject: "Next stage -> " + s.subjForNext(),
		Meta:        string(msg.Data),
	}
	// add this data to actions if actions was previously added to this ctx
	// otherwise, nothing is added
	noop.ActionsFromCtx(ctx).AddEvent(event)

	// finally, return if noop operation
	if noop.ContainsNoop(ctx) {
		return true
	}

	// otherwise, act as normal , i.e. publish the message to trigger the next service
	return NatsPublish(s.subjForNext(), string(msg.Data), nil)
}

type d struct {
	Data string `json:"Data"`
}

func (s *Service) subjForThis() string {
	return "pipeline.stage." + strconv.Itoa(s.Stage)
}
func (s *Service) subjForNext() string {
	return "pipeline.stage." + strconv.Itoa(s.Stage+1)
}
