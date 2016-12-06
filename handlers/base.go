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
	case http.MethodGet:
		w.Get(ctx)
	case http.MethodPost:
		w.Post(ctx)
	case http.MethodDelete:
		w.Delete(ctx)
	case http.MethodPut:
		w.Put(ctx)
	case http.MethodPatch:
		w.Patch(ctx)
	case http.MethodHead:
		w.Head(ctx)
	case http.MethodOptions:
		w.Options(ctx)
	case http.MethodTrace:
		w.Trace(ctx)
	case http.MethodConnect:
		w.Connect(ctx)
	default:
		ctx.Resp.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func WrapBase(b Base) roboot.Handler {
	return wrappedBase{Base: b}
}
