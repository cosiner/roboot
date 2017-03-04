package handlers

import "github.com/cosiner/roboot"

type interceptor struct {
	filter  roboot.Filter
	handler roboot.Handler
}

func (i *interceptor) Handle(ctx *roboot.Context) {
	if i.filter == nil {
		i.handler.Handle(ctx)
	} else {
		i.filter.Filter(ctx, i.handler.Handle)
	}
}

func Intercept(handler roboot.Handler, filters ...roboot.Filter) roboot.Handler {
	l := len(filters)
	if l == 0 {
		return handler
	}

	return Intercept(&interceptor{handler: handler, filter: filters[l-1]}, filters[:l-1]...)
}
