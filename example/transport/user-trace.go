// GENERATED BY 'T'ransport 'G'enerator. DO NOT EDIT.
package transport

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/seniorGolang/tg/example/interfaces"
	"github.com/seniorGolang/tg/example/interfaces/types"
)

type traceUser struct {
	next interfaces.User
}

func traceMiddlewareUser(next interfaces.User) interfaces.User {
	return &traceUser{next: next}
}

func (svc traceUser) GetUser(ctx context.Context, cookie string, userAgent string) (user *types.User, err error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "GetUser")
	return svc.next.GetUser(ctx, cookie, userAgent)
}

func (svc traceUser) UploadFile(ctx context.Context, fileBytes []byte) (err error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "UploadFile")
	return svc.next.UploadFile(ctx, fileBytes)
}

func (svc traceUser) CustomResponse(ctx context.Context, arg0 int, arg1 string, opts ...interface{}) (err error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "CustomResponse")
	return svc.next.CustomResponse(ctx, arg0, arg1, opts...)
}

func (svc traceUser) CustomHandler(ctx context.Context, arg0 int, arg1 string, opts ...interface{}) (err error) {
	span := opentracing.SpanFromContext(ctx)
	span.SetTag("method", "CustomHandler")
	return svc.next.CustomHandler(ctx, arg0, arg1, opts...)
}
