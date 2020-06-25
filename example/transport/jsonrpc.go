// GENERATED BY 'T'ransport 'G'enerator. DO NOT EDIT.
package transport

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/savsgio/gotils"
	"github.com/valyala/fasthttp"
)

const (
	// Version defines the version of the JSON RPC implementation
	Version = "2.0"
	// contentTypeJson defines the content type to be served
	contentTypeJson = "application/json"
	// ParseError defines invalid JSON was received by the server
	// An error occurred on the server while parsing the JSON text
	parseError = -32700
	// InvalidRequestError defines the JSON sent is not a valid Request object
	invalidRequestError = -32600
	// MethodNotFoundError defines the method does not exist / is not available
	methodNotFoundError = -32601
	// InvalidParamsError defines invalid method parameter(s)
	invalidParamsError = -32602
	// InternalError defines a server error
	internalError = -32603
)

type idJsonRPC = json.RawMessage

type baseJsonRPC struct {
	ID      idJsonRPC       `json:"id"`
	Version string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Error   *errorJsonRPC   `json:"error,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type errorJsonRPC struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err errorJsonRPC) Error() string {
	return err.Message
}

type jsonrpcResponses []baseJsonRPC

func (responses *jsonrpcResponses) append(response *baseJsonRPC) {
	if response == nil {
		return
	}
	if response.ID != nil {
		*responses = append(*responses, *response)
	}
}

type methodJsonRPC func(span opentracing.Span, ctx *fasthttp.RequestCtx, requestBase baseJsonRPC) (responseBase *baseJsonRPC)

func (srv *Server) serveBatch(ctx *fasthttp.RequestCtx) {

	batchSpan := extractSpan(srv.log, fmt.Sprintf("jsonRPC:%s", gotils.B2S(ctx.URI().Path())), ctx)
	defer injectSpan(srv.log, batchSpan, ctx)
	defer batchSpan.Finish()
	methodHTTP := gotils.B2S(ctx.Method())

	if methodHTTP != fasthttp.MethodPost {
		ext.Error.Set(batchSpan, true)
		batchSpan.SetTag("msg", "only POST method supported")
		ctx.Error("only POST method supported", fasthttp.StatusMethodNotAllowed)
		return
	}

	for _, handler := range srv.httpBefore {
		handler(ctx)
	}

	if value := ctx.Value(CtxCancelRequest); value != nil {
		return
	}

	ctx.SetContentType(contentTypeJson)

	var err error
	var requests []baseJsonRPC

	if err = json.Unmarshal(ctx.PostBody(), &requests); err != nil {
		ext.Error.Set(batchSpan, true)
		batchSpan.SetTag("msg", "request body could not be decoded: "+err.Error())

		for _, handler := range srv.httpAfter {
			handler(ctx)
		}
		sendResponse(srv.log, ctx, makeErrorResponseJsonRPC([]byte("\"0\""), parseError, "request body could not be decoded: "+err.Error(), nil))
		return
	}

	responses := make(jsonrpcResponses, 0, len(requests))

	var wg sync.WaitGroup

	for _, request := range requests {

		methodNameOrigin := request.Method
		method := strings.ToLower(request.Method)

		span := opentracing.StartSpan(request.Method, opentracing.ChildOf(batchSpan.Context()))
		span.SetTag("batch", true)

		switch method {

		case "jsonrpc.test":

			wg.Add(1)
			go func(request baseJsonRPC) {
				responses.append(srv.httpJsonRPC.test(span, ctx, request))
				wg.Done()
			}(request)

		default:
			ext.Error.Set(span, true)
			span.SetTag("msg", "invalid method '"+methodNameOrigin+"'")
			responses.append(makeErrorResponseJsonRPC(request.ID, methodNotFoundError, "invalid method '"+methodNameOrigin+"'", nil))
		}
		span.Finish()
	}
	wg.Wait()

	for _, handler := range srv.httpAfter {
		handler(ctx)
	}
	sendResponse(srv.log, ctx, responses)
}

func makeErrorResponseJsonRPC(id idJsonRPC, code int, msg string, data interface{}) *baseJsonRPC {

	if id == nil {
		return nil
	}

	return &baseJsonRPC{
		Error: &errorJsonRPC{
			Code:    code,
			Data:    data,
			Message: msg,
		},
		ID:      id,
		Version: Version,
	}
}