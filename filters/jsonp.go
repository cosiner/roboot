package filters

import (
	"bytes"

	"github.com/cosiner/roboot"
)

// use as callback parameter name such as ?callback=xxx
type JSONP string

var _ roboot.Filter = JSONP(0)

type buffRespWriter struct {
	roboot.ResponseWriter
	Buffer *bytes.Buffer
}

func (w *buffRespWriter) Write(b []byte) (int, error) {
	if w.Buffer == nil {
		return w.ResponseWriter.Write(b)
	}
	return w.Buffer.Write(b)
}

func (j JSONP) Filter(ctx *roboot.Context, chain roboot.Handler) {
	if ctx.Req.Method != roboot.MethodGet {
		chain.Handle(ctx)
		return
	}

	callback := ctx.QueryValue(string(j))
	if callback == "" {
		chain.Handle(ctx)
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 256))
	bw := buffRespWriter{ // to avoid write header 200 first when write callback name
		ResponseWriter: ctx.Resp,
		Buffer:         buffer,
	}
	ctx.Resp = &bw
	chain.Handle(ctx)
	bw.Buffer = nil

	ctx.Resp.Write([]byte("("))
	ctx.Resp.Write(bw.Buffer.Bytes())
	ctx.Resp.Write([]byte(")"))
}
