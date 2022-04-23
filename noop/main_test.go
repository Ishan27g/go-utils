package main

import (
	"context"
	"testing"
	"time"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/stretchr/testify/assert"
)

func buildTestUrl(isNoop bool) string {
	if isNoop {
		return host + "9999" + "/endpoint1" + data + "something"
	}
	return host + "9999" + "/endpoint1" + data + "something" + urlWithNoop
}
func testRequestWithoutActions(t *testing.T) {
	assert.Equal(t, sendHttpReq(context.Background(), buildTestUrl(false)),
		sendHttpReq(context.Background(), buildTestUrl(true)))
}

func testRequestWithActions(t *testing.T) {
	actions := noop.NewActions()
	actions.AddEvent(noop.Event{
		Name: "From Client",
		Meta: nil,
	})
	ctx := noop.NewCtxWithActions(context.Background(), actions)

	assert.Equal(t, sendHttpReq(ctx, buildTestUrl(false)),
		sendHttpReq(ctx, buildTestUrl(true)))
}
func testRequestWithNoopCtxNoActions(t *testing.T) {
	ctx := noop.NewCtxWithNoop(context.Background(), true)
	assert.Equal(t, sendHttpReq(ctx, buildTestUrl(false)),
		sendHttpReq(ctx, buildTestUrl(true)))
}
func testRequestWithNoopCtxAndActions(t *testing.T) {
	actions := noop.NewActions()
	actions.AddEvent(noop.Event{
		Name: "From Client",
		Meta: nil,
	})
	ctx := noop.NewCtxWithActions(context.Background(), actions)
	ctx = noop.NewCtxWithNoop(ctx, true)
	assert.Equal(t, sendHttpReq(ctx, buildTestUrl(false)),
		sendHttpReq(ctx, buildTestUrl(true)))
}

func Test_Noop(t *testing.T) {
	go startServer()
	<-time.After(1 * time.Second)
	testRequestWithoutActions(t)
	testRequestWithActions(t)
	testRequestWithNoopCtxNoActions(t)
	testRequestWithNoopCtxAndActions(t)
}
