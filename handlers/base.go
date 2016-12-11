package handlers

import (
	"net/http"

	"github.com/cosiner/roboot"
)

type MethodBase interface {
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

type NopMethodBase struct {
}

var _ MethodBase = NopMethodBase{}

func (NopMethodBase) Get(ctx *roboot.Context)     { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Post(ctx *roboot.Context)    { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Delete(ctx *roboot.Context)  { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Put(ctx *roboot.Context)     { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Patch(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Head(ctx *roboot.Context)    { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Options(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Trace(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopMethodBase) Connect(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }

type wrappedMethodBase struct {
	MethodBase
}

func (w wrappedMethodBase) Handle(ctx *roboot.Context) {
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

type ActionBase interface {
	Query(*roboot.Context)
	Create(*roboot.Context)
	Delete(*roboot.Context)
	Update(*roboot.Context)
	CreateOrUpdate(*roboot.Context)
	Head(*roboot.Context)
	Options(*roboot.Context)
	Trace(*roboot.Context)
	Connect(*roboot.Context)
}

type NopActionBase struct {
}

var _ ActionBase = NopActionBase{}

func (NopActionBase) Query(ctx *roboot.Context)  { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Create(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Delete(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Update(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Head(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) CreateOrUpdate(ctx *roboot.Context) {
	ctx.Resp.WriteHeader(http.StatusMethodNotAllowed)
}
func (NopActionBase) Options(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Trace(ctx *roboot.Context)   { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }
func (NopActionBase) Connect(ctx *roboot.Context) { ctx.Resp.WriteHeader(http.StatusMethodNotAllowed) }

type wrappedActionBase struct {
	ActionBase
}

func (w wrappedActionBase) Handle(ctx *roboot.Context) {
	switch ctx.Req.Method {
	case roboot.METHOD_GET:
		w.Query(ctx)
	case roboot.METHOD_POST:
		w.Create(ctx)
	case roboot.METHOD_DELETE:
		w.Delete(ctx)
	case roboot.METHOD_PUT:
		w.CreateOrUpdate(ctx)
	case roboot.METHOD_PATCH:
		w.Update(ctx)
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

func WrapBase(b interface{}, wrapper ...func(interface{}) roboot.Handler) roboot.Handler {
	switch t := b.(type) {
	case MethodBase:
		return wrappedMethodBase{
			MethodBase: t,
		}
	case ActionBase:
		return wrappedActionBase{
			ActionBase: t,
		}
	default:
		if len(wrapper) > 0 {
			return wrapper[0](b)
		}

		panic("unsupported base type")
	}
}
