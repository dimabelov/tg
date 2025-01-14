// GENERATED BY 'T'ransport 'G'enerator. DO NOT EDIT.
package transport

import (
	"context"
	"fmt"
	"github.com/seniorGolang/tg/v2/example/interfaces"
	"time"
)

type metricsExampleRPC struct {
	next interfaces.ExampleRPC
}

func metricsMiddlewareExampleRPC(next interfaces.ExampleRPC) interfaces.ExampleRPC {
	return &metricsExampleRPC{next: next}
}

func (m metricsExampleRPC) Test(ctx context.Context, arg0 int, arg1 string, opts ...interface{}) (ret1 int, ret2 string, err error) {

	defer func(_begin time.Time) {
		RequestCount.WithLabelValues("exampleRPC", "test", fmt.Sprint(err == nil)).Add(1)
		RequestCountAll.WithLabelValues("exampleRPC", "test", fmt.Sprint(err == nil)).Add(1)
		RequestLatency.WithLabelValues("exampleRPC", "test", fmt.Sprint(err == nil)).Observe(time.Since(_begin).Seconds())
	}(time.Now())

	return m.next.Test(ctx, arg0, arg1, opts...)
}

func (m metricsExampleRPC) Test2(ctx context.Context, arg0 int, arg1 string, opts ...interface{}) (ret1 int, ret2 string, err error) {

	defer func(_begin time.Time) {
		RequestCount.WithLabelValues("exampleRPC", "test2", fmt.Sprint(err == nil)).Add(1)
		RequestCountAll.WithLabelValues("exampleRPC", "test2", fmt.Sprint(err == nil)).Add(1)
		RequestLatency.WithLabelValues("exampleRPC", "test2", fmt.Sprint(err == nil)).Observe(time.Since(_begin).Seconds())
	}(time.Now())

	return m.next.Test2(ctx, arg0, arg1, opts...)
}
