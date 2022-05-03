package tracing

//
//import (
//	"context"
//
//	"github.com/Ishan27g/go-utils/noop/noop"
//	"go.opentelemetry.io/otel/attribute"
//	"go.opentelemetry.io/otel/codes"
//	"go.opentelemetry.io/otel/trace"
//)
//
//var status = make(chan bool, 1)
//var exporter, app, service = "", "", ""
//
//func Disable() {
//	<-status
//	defer func() { status <- false }()
//	if tp.Get() != nil {
//		tp.Close()
//	}
//	tp = &provider{
//		exporter:       tp.exporter,
//		app:            tp.app,
//		service:        tp.service,
//		TracerProvider: nil,
//	}
//}
//func Enable() {
//	<-status
//	defer func() { status <- true }()
//	if tp.Get() == nil {
//		initialise(tp.exporter, tp.app, tp.service)
//	}
//}
//
//type NoopSpan interface {
//	End(options ...trace.SpanEndOption)
//	AddEvent(name string, options ...trace.EventOption)
//	IsRecording() bool
//	RecordError(err error, options ...trace.EventOption)
//	SpanContext() trace.SpanContext
//	SetStatus(code codes.Code, description string)
//	SetName(name string)
//	SetAttributes(kv ...attribute.KeyValue)
//	TracerProvider() trace.TracerProvider
//}
//type noopSpan struct {
//	trace.Span
//}
//
//func (s *noopSpan) End(options ...trace.SpanEndOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	s.Span.End(options...)
//}
//
//func (s *noopSpan) AddEvent(name string, options ...trace.EventOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	s.Span.AddEvent(name, options...)
//}
//
//func (s *noopSpan) IsRecording() bool {
//	return s.Span.IsRecording()
//}
//
//func (s *noopSpan) RecordError(err error, options ...trace.EventOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	s.Span.RecordError(err, options...)
//}
//
//func (s *noopSpan) SpanContext() trace.SpanContext {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return trace.NewSpanContext(trace.SpanContextConfig{})
//	}
//	return s.Span.SpanContext()
//
//}
//
//func (s *noopSpan) SetStatus(code codes.Code, description string) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	s.Span.SetStatus(code, description)
//}
//
//func (s *noopSpan) SetName(name string) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	s.Span.SetName(name)
//}
//
//func (s *noopSpan) SetAttributes(kv ...attribute.KeyValue) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//}
//
//func (s *noopSpan) TracerProvider() trace.TracerProvider {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return nt.TraceProvider
//	}
//	return s.Span.TracerProvider()
//}
//
//type NoopTracer interface {
//	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
//	Tracer(instrumentationName string, opts ...trace.TracerOption) trace.Tracer
//}
//
//type noopTracer struct {
//	tr NoopTracer
//	TraceProvider
//}
//
//func (nt *noopTracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
//	if noop.ContainsNoop(ctx) {
//		return ctx, &noopSpan{}
//	}
//	return nt.Start(ctx, spanName, opts...)
//}
//
//var nt = &noopTracer{}
//
////func Init(expo, appName, serviceName string) NoopTracer {
////	exporter, app, service = expo, appName, serviceName
////	nt = &noopTracer{tr: Init(expo, app, service)}
////	return nt
////}
////
////func test(){
////	noopTracer := Init("","","")
////	noopTracer.
////}
