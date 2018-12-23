package filters

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/cosiner/roboot"
)

func unsafeToString(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))

	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len

	return
}

type Recovery struct {
	Bufsize int
}

var _ roboot.Filter = &Recovery{}

func (r Recovery) Filter(ctx *roboot.Context, chain roboot.Handler) {
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
			ctx.Env().Error.Log(ctx, roboot.ErrTypePanic, fmt.Errorf("panic: %v, stack: %s", err, unsafeToString(buf)))
		}
	}()

	chain.Handle(ctx)
}
