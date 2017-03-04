package filters

import (
	"errors"
	"net/http"
	"runtime"

	"github.com/cosiner/roboot"
)

type Recovery struct {
	Bufsize int
}

func (r Recovery) Filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	const defaultBufsize = 4096
	defer func() {
		if err := recover(); err != nil {

			bufsize := r.Bufsize
			if bufsize <= 0 {
				bufsize = defaultBufsize
			}
			buf := make([]byte, bufsize)
			n := runtime.Stack(buf, false)
			buf = buf[:n]

			ctx.Status(http.StatusInternalServerError)
			ctx.Env().Error.Log(ctx, roboot.ErrTypePanic, errors.New(string(buf)))
		}
	}()

	chain(ctx)
}
