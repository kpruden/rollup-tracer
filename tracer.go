package rolluptracer

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// RollupTracer wraps another tracer, and produces spans which roll up
// tags and durations to the nearest ancestor span from the parent tracer.
type RollupTracer struct {
	parent opentracing.Tracer
}

// NewRollupTracer creates a RollupTracer wrapping the parent tracer provided
func NewRollupTracer(parent opentracing.Tracer) opentracing.Tracer {
	return &RollupTracer{parent}
}

func findParentContext(sso opentracing.StartSpanOptions) opentracing.SpanContext {
	for _, ref := range sso.References {
		if ref.Type == opentracing.ChildOfRef {
			return ref.ReferencedContext
		}
	}
	return nil
}

// StartSpan creates, starts, and returns a new Span with the given `operationName` and
// incorporate the given StartSpanOption `opts`.
func (t *RollupTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}

	if sso.Tags != nil {
		// Always start a new "real" span for the two ends of an RPC call or producer/consumer request
		kind := sso.Tags[string(ext.SpanKind)]
		if kind == ext.SpanKindRPCClientEnum ||
			kind == ext.SpanKindRPCServerEnum ||
			kind == ext.SpanKindConsumerEnum ||
			kind == ext.SpanKindProducerEnum {
			return t.StartRealSpan(operationName, opts...)
		}
	}

	if sso.StartTime.IsZero() {
		sso.StartTime = time.Now()
	}

	parentCtx := findParentContext(sso)

	if parentCtx == nil {
		return &wrapperSpan{
			tracer: t,
			context: &wrapperSpanContext{
				root: t.parent.StartSpan(operationName, opts...),
			},
		}
	}

	var context *rollupSpanContext

	if rc, ok := parentCtx.(*rollupSpanContext); ok {
		context = &rollupSpanContext{
			root:   rc.root,
			parent: rc,
		}
	} else if wc, ok := parentCtx.(*wrapperSpanContext); ok {
		context = &rollupSpanContext{
			root: wc.root,
		}
	} else {
		// This should never happen, but we'll return a wrapped "real" span just in case
		return &wrapperSpan{
			tracer: t,
			context: &wrapperSpanContext{
				root: t.parent.StartSpan(operationName, opts...),
			},
		}
	}

	span := &rollupSpan{
		tracer:    t,
		context:   context,
		startTime: sso.StartTime,
	}

	span.SetOperationName(operationName)

	if sso.Tags != nil {
		for k, v := range sso.Tags {
			span.SetTag(k, v)
		}
	}

	return span
}

// StartRealSpan delegates to the parent tracer to start a "real" span. This creates a new
// rollup context.
func (t *RollupTracer) StartRealSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	for i, o := range opts {
		if ref, ok := o.(opentracing.SpanReference); ok && ref.Type == opentracing.ChildOfRef {
			ctx := ref.ReferencedContext
			if rc, ok := ctx.(*rollupSpanContext); ok {
				ref.ReferencedContext = rc.root.Context()
			} else if rc, ok := ctx.(*wrapperSpanContext); ok {
				ref.ReferencedContext = rc.root.Context()
			}
			opts[i] = ref
		}
	}

	return &wrapperSpan{
		tracer: t,
		context: &wrapperSpanContext{
			root: t.parent.StartSpan(operationName, opts...),
		},
	}
}

// Inject takes the `sm` SpanContext instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`.
func (t *RollupTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	if rc, ok := sm.(*rollupSpanContext); ok {
		sm = rc.root.Context()
	}
	if rc, ok := sm.(*wrapperSpanContext); ok {
		sm = rc.root.Context()
	}
	return t.parent.Inject(sm, format, carrier)
}

// Extract returns a SpanContext instance given `format` and `carrier`.
func (t *RollupTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return t.parent.Extract(format, carrier)
}
