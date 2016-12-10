package handlers

import (
	"net/http"

	"github.com/cosiner/roboot"
)

type Base interface {
	Get(*roboot.Context)
	Post(*roboot.Context)
	Delete(*roboot.Context)
	Patch(*roboot.Context)
	Put(*roboot.Context)
	Options(*roboot.Context)
	Head(*roboot.Context)
	Trace(*roboot.Context)
	Connect(*roboot.Context)
}

type NopBase struct {
}

func (NopBase) Get(ctx *roboot.Context)     { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Post(ctx *roboot.Context)    { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Delete(ctx *roboot.Context)  { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Put(ctx *roboot.Context)     { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Patch(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Head(ctx *roboot.Context)    { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Options(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Trace(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopBase) Connect(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }

type wrappedBase struct {
	Base
}

func (w wrappedBase) Handle(ctx *roboot.Context) {
	switch ctx.Req.Method {
	case roboot.METHOD_GET:
		w.Get(ctx)
	case roboot.METHOD_POST:
		w.Post(ctx)
	case roboot.METHOD_DELETE:
		w.Delete(ctx)
	case roboot.METHOD_PUT:
		w.Put(ctx)
	case roboot.METHOD_PATCH:
		w.Patch(ctx)
	case roboot.METHOD_HEAD:
		w.Head(ctx)
	case roboot.METHOD_OPTIONS:
		w.Options(ctx)
	case roboot.METHOD_TRACE:
		w.Trace(ctx)
	case roboot.METHOD_CONNECT:
		w.Connect(ctx)
	default:
		ctx.Resp.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func WrapBase(b Base) roboot.Handler {
	return wrappedBase{Base: b}
}
