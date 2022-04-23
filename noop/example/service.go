package example

import (
	"context"
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
}

func (s *Service) subjForThis() string {
	return "pipeline.stage." + strconv.Itoa(s.Stage)
}
func (s *Service) subjForNext() string {
	return "pipeline.stage." + strconv.Itoa(s.Stage+1)
}
func New(name string, stage int) *Service {
	s := Service{
		Name:  name,
		Stage: stage,
	}
	return &s
}

// GenericMethod is any generic async method in a request pipeline
// It would normally be triggered by nats/kafka. After processing
// this method would ideally publish to nats/kafka to trigger the next
// service/stage of the request pipeline
/* Noop-Testing: Before the final publish operation ->
- add publish `msg` to the context
- return if ctx is noop, otherwise publish `msg` to queue
*/
func (s *Service) GenericMethod(ctx context.Context, data *nats.Msg) bool {

	// data validation
	// ...
	// ...
	<-time.After(1 * time.Second)

	// define event to be added
	event := noop.Event{
		Name:        "Async method triggered at " + s.Name,
		NextSubject: s.subjForNext(),
		Meta:        string(data.Data),
	}

	// add this data to actions if actions was previously added to this ctx
	// otherwise, nothing is added
	noop.ActionsFromCtx(ctx).AddEvent(event)

	// finally, return if noop operation
	if noop.ContainsNoop(ctx) {
		return true
	}

	// otherwise, act as normal , i.e. publish the message to trigger the next service
	return NatsPublish(s.subjForNext(), string(data.Data), nil)
}

func (s *Service) TriggerAsyncFromHttp(c *gin.Context) {

	/* Noop-Testing:
	   Wrap async method as its own http handler, which based on the `noop` param
	   - extracts the action from request
	   - triggers underlying async method
	   - responds with actions taken by this method
	*/

	// is request a `noop`
	isNoop := strings.EqualFold(c.Query("noop"), "true")

	ctx := noop.NewCtxWithNoop(c, isNoop) // todo new ctx or gin-request context?

	// actions for this request
	actions := noop.NewActions()

	// add an event
	// actions.AddEvent(noop.Event{Name: "Endpoint", Meta: c.Request.URL.RequestURI()})
	ctx = noop.NewCtxWithActions(ctx, actions)

	// manually trigger the async method
	if !s.GenericMethod(ctx, nats.NewMsg("")) {
		c.JSON(http.StatusExpectationFailed, nil)
	}

	if rsp, err := actions.Marshal(); err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(rsp)
		return
	}
	c.JSON(http.StatusExpectationFailed, nil)
}