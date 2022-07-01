package rabbitMq

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Ishan27g/go-utils/tracing"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ExampleHandler that propagates span over rabbitmq.
// Extracts span from http request, publishes it as rabbitmq headers, then consumes and extracts
// a span which is a child of the original span.
func ExampleHandler(c *gin.Context) {
	provider := tracing.Init("jaeger", "api-gw", "proxy")
	defer provider.Close()

	go func() {
		headers := consumeOnce() // consume from queue
		ctx := ExtractSpan(headers)
		ctx, span := provider.Get().Start(ctx, "consumer-from-rabbit")
		defer span.End()
		span.SetAttributes(attribute.Key("consumer").Int(1))
	}()

	_, span := provider.Get().Start(c.Request.Context(), "producer-to-rabbit")
	span.SetAttributes(attribute.Key("producer").Int(1))
	defer span.End()

	headers := ExtractHeader(span)
	publish(headers) // publish to queue

	c.JSON(http.StatusOK, "OK\n")

}

const traceId = "TraceId"
const spanId = "SpanId"

// ExtractHeader injects otel span into ampq headers
func ExtractHeader(span trace.Span) (header amqp.Table) {
	header = make(amqp.Table)
	header[traceId] = span.SpanContext().TraceID().String()
	header[spanId] = span.SpanContext().SpanID().String()
	return
}

// ExtractSpan returns a context with underlying span extracted from ampq headers
func ExtractSpan(header amqp.Table) context.Context {
	traceID, err := trace.TraceIDFromHex(header[traceId].(string))
	if err != nil {
		fmt.Println("no traceId in header: ", err)
	}
	spanID, err := trace.SpanIDFromHex(header[spanId].(string))
	if err != nil {
		fmt.Println("no spanId in header: ", err)
	}
	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = false
	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(spanContextConfig))
}
