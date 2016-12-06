package handlers

import "github.com/cosiner/roboot"

type Status int

func (s Status) Handle(ctx *roboot.Context) {
	ctx.Resp.WriteHeader(int(s))
}
