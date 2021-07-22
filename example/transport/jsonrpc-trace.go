// GENERATED BY 'T'ransport 'G'enerator. DO NOT EDIT.
package transport

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/seniorGolang/tg/example/interfaces"
)

type traceJsonRPC struct {
	next interfaces.JsonRPC
}

func traceMiddlewareJsonRPC(next interfaces.JsonRPC) interfaces.JsonRPC {
	return &traceJsonRPC{next: next}
}

func (svc traceJsonRPC) Test(ctx context.Context, arg0 int, arg1 string, opts ...interface{}) (ret1 int, ret2 string, err error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "Test")
	return svc.next.Test(ctx, arg0, arg1, opts...)
}
