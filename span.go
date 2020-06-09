package rolluptracer

import (
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	tlog "github.com/opentracing/opentracing-go/log"
)

func trimCommonPrefix(s1, s2 string) string {
	r1 := []rune(s1)
	r2 := []rune(s2)
	i := 0
	for ; i < len(r1) && i < len(r2) && r1[i] == r2[i]; i++ {
	}

	return strings.TrimPrefix(string(r2[i:]), ".")
}

type rollupSpan struct {
	tracer    *RollupTracer
	context   *rollupSpanContext
	startTime time.Time
	finished  bool
}

var _ opentracing.Span = (*rollupSpan)(nil)

func (s *rollupSpan) Finish() {
	s.FinishWithOptions(opentracing.FinishOptions{})
}

func (s *rollupSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	if s.finished {
		return
	}
	if opts.FinishTime.IsZero() {
		opts.FinishTime = time.Now()
	}

	duration := float64(float64(opts.FinishTime.Sub(s.startTime)) / 1000000.0)
	s.SetTag("durationMs", duration)

	// TODO handle opts.LogRecords
	s.finished = true
}

func (s *rollupSpan) Context() opentracing.SpanContext {
	return s.context
}

func (s *rollupSpan) SetOperationName(operationName string) opentracing.Span {
	s.context.operationName = operationName
	if s.context.parent != nil {
		operationName = trimCommonPrefix(s.context.parent.operationName, operationName)
		operationName = fmt.Sprintf("%s.%s", s.context.parent.effOpName, operationName)
	}
	s.context.effOpName = operationName
	return s
}

func (s *rollupSpan) SetTag(key string, value interface{}) opentracing.Span {
	effKey := fmt.Sprintf("%s.%s", s.context.effOpName, key)
	s.context.root.SetTag(effKey, value)
	return s
}

func (s *rollupSpan) LogFields(fields ...tlog.Field) {
	s.context.root.LogFields(fields...)
}

func (s *rollupSpan) LogKV(alternatingKeyValues ...interface{}) {
	s.context.root.LogKV(alternatingKeyValues...)
}

func (s *rollupSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	s.context.root.SetBaggageItem(restrictedKey, value)
	return s
}

func (s *rollupSpan) BaggageItem(restrictedKey string) string {
	return s.context.root.BaggageItem(restrictedKey)
}

func (s *rollupSpan) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *rollupSpan) LogEvent(event string) {
	s.context.root.LogEvent(event)
}

func (s *rollupSpan) LogEventWithPayload(event string, payload interface{}) {
	s.context.root.LogEventWithPayload(event, payload)
}

func (s *rollupSpan) Log(data opentracing.LogData) {
	s.context.root.Log(data)
}

type wrapperSpan struct {
	tracer  *RollupTracer
	context *wrapperSpanContext
}

var _ opentracing.Span = (*wrapperSpan)(nil)

func (s *wrapperSpan) Finish() {
	s.context.root.Finish()
}

func (s *wrapperSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	s.context.root.FinishWithOptions(opts)
}

func (s *wrapperSpan) Context() opentracing.SpanContext {
	return s.context
}

func (s *wrapperSpan) SetOperationName(operationName string) opentracing.Span {
	s.context.root.SetOperationName(operationName)
	return s
}

func (s *wrapperSpan) SetTag(key string, value interface{}) opentracing.Span {
	s.context.root.SetTag(key, value)
	return s
}

func (s *wrapperSpan) LogFields(fields ...tlog.Field) {
	s.context.root.LogFields(fields...)
}

func (s *wrapperSpan) LogKV(alternatingKeyValues ...interface{}) {
	s.context.root.LogKV(alternatingKeyValues...)
}

func (s *wrapperSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	s.context.root.SetBaggageItem(restrictedKey, value)
	return s
}

func (s *wrapperSpan) BaggageItem(restrictedKey string) string {
	return s.context.root.BaggageItem(restrictedKey)
}

func (s *wrapperSpan) Tracer() opentracing.Tracer {
	return s.tracer
}

func (s *wrapperSpan) LogEvent(event string) {
	s.context.root.LogEvent(event)
}

func (s *wrapperSpan) LogEventWithPayload(event string, payload interface{}) {
	s.context.root.LogEventWithPayload(event, payload)
}

func (s *wrapperSpan) Log(data opentracing.LogData) {
	s.context.root.Log(data)
}
