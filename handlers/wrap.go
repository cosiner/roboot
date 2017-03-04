package handlers

import (
	"net/http"

	"github.com/cosiner/roboot"
)

type MethodHandler interface {
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

type NopMethodHandler struct {
}

var _ MethodHandler = NopMethodHandler{}

func (NopMethodHandler) Get(ctx *roboot.Context)     { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Post(ctx *roboot.Context)    { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Delete(ctx *roboot.Context)  { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Put(ctx *roboot.Context)     { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Patch(ctx *roboot.Context)   { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Head(ctx *roboot.Context)    { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Options(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Trace(ctx *roboot.Context)   { ctx.Status(http.StatusMethodNotAllowed) }
func (NopMethodHandler) Connect(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }

type methodHandlerWrapper struct {
	MethodHandler
}

func (w methodHandlerWrapper) Handle(ctx *roboot.Context) {
	switch ctx.Req.Method {
	case roboot.MethodGet:
		w.Get(ctx)
	case roboot.MethodPost:
		w.Post(ctx)
	case roboot.MethodDelete:
		w.Delete(ctx)
	case roboot.MethodPut:
		w.Put(ctx)
	case roboot.MethodPatch:
		w.Patch(ctx)
	case roboot.MethodHead:
		w.Head(ctx)
	case roboot.MethodOptions:
		w.Options(ctx)
	case roboot.MethodTrace:
		w.Trace(ctx)
	case roboot.MethodConnect:
		w.Connect(ctx)
	default:
		ctx.Status(http.StatusMethodNotAllowed)
	}
}

type ActionHandler interface {
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

type NopActionHandler struct {
}

var _ ActionHandler = NopActionHandler{}

func (NopActionHandler) Query(ctx *roboot.Context)  { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Create(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Delete(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Update(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Head(ctx *roboot.Context)   { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) CreateOrUpdate(ctx *roboot.Context) {
	ctx.Status(http.StatusMethodNotAllowed)
}
func (NopActionHandler) Options(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Trace(ctx *roboot.Context)   { ctx.Status(http.StatusMethodNotAllowed) }
func (NopActionHandler) Connect(ctx *roboot.Context) { ctx.Status(http.StatusMethodNotAllowed) }

type actionHandlerWrapper struct {
	ActionHandler
}

func (w actionHandlerWrapper) Handle(ctx *roboot.Context) {
	switch ctx.Req.Method {
	case roboot.MethodGet:
		w.Query(ctx)
	case roboot.MethodPost:
		w.Create(ctx)
	case roboot.MethodDelete:
		w.Delete(ctx)
	case roboot.MethodPut:
		w.CreateOrUpdate(ctx)
	case roboot.MethodPatch:
		w.Update(ctx)
	case roboot.MethodHead:
		w.Head(ctx)
	case roboot.MethodOptions:
		w.Options(ctx)
	case roboot.MethodTrace:
		w.Trace(ctx)
	case roboot.MethodConnect:
		w.Connect(ctx)
	default:
		ctx.Status(http.StatusMethodNotAllowed)
	}
}

func Wrap(b interface{}, wrapper ...func(interface{}) roboot.Handler) roboot.Handler {
	switch t := b.(type) {
	case roboot.Handler:
		return t
	case MethodHandler:
		return methodHandlerWrapper{
			MethodHandler: t,
		}
	case ActionHandler:
		return actionHandlerWrapper{
			ActionHandler: t,
		}
	default:
		if len(wrapper) > 0 {
			return wrapper[0](b)
		}

		panic("unsupported base type")
	}
}
