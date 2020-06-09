package rolluptracer

import (
	"github.com/opentracing/opentracing-go"
)

type rollupSpanContext struct {
	root          opentracing.Span
	parent        *rollupSpanContext
	operationName string
	effOpName     string
}

var _ opentracing.SpanContext = (*rollupSpanContext)(nil)

func (sc *rollupSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	sc.root.Context().ForeachBaggageItem(handler)
}

type wrapperSpanContext struct {
	root opentracing.Span
}

var _ opentracing.SpanContext = (*wrapperSpanContext)(nil)

func (sc *wrapperSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	sc.root.Context().ForeachBaggageItem(handler)
}
