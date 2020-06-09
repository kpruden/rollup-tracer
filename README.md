# Rollup Tracer

RollupTracer is a simple OpenTracing tracer that wraps another tracer and "rolls" up intermediate spans to a common ancestor. It's intended to be used in heavily-instrumented codebases, where many of the intermediate spans are useful for debugging during development, but let useful in a production environment.

The tracer rolls up the tags and durations of all the spans it creates to the nearest "real" span, applying a hierarchical namespace scheme based on the operation names of the spans.

The implementation is currently pretty naive - in particular, it does not deal with the creation of multiple child spans at the same level with the same operation name.

The tracer offers an additional way to create spans: `StartRealSpan`. This delegates directly to the parent tracer to create a "real" span, as a direct child of the nearest "real" ancestor of the parent context provided (if any).

The tracer automatically creates a "real" span for any span tagged with the following `span.kind` at span-creation time:

 * `client`
 * `server`
 * `producer`
 * `consumer`

This code is, at best, a proof of concept. It works as far as I know, but testing has been minimal.
