package handlers

import (
	"net/http"

	"github.com/cosiner/roboot"
)

type Methods map[string]roboot.HandlerFunc

func (m Methods) Handle(ctx *roboot.Context) {
	handler, has := m[ctx.Req.Method]
	if !has {
		ctx.Status(http.StatusMethodNotAllowed)
	} else {
		handler(ctx)
	}
}
