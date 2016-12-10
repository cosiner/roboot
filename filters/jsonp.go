package filters

import (
	"bytes"
	"net/http"

	"github.com/cosiner/roboot"
)

// use as callback parameter name such as ?callback=xxx
type JSONP string

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

func (j JSONP) Filter(ctx *roboot.Context, chain roboot.HandlerFunc) {
	if ctx.Req.Method != http.MethodGet {
		chain(ctx)
		return
	}

	callback := ctx.QueryValue(string(j))
	if callback == "" {
		chain(ctx)
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 256))
	bw := buffRespWriter{ // to avoid write header 200 first when write callback name
		ResponseWriter: ctx.Resp,
		Buffer:         buffer,
	}
	ctx.Resp = &bw
	chain(ctx)
	bw.Buffer = nil

	ctx.Resp.Write([]byte("("))
	ctx.Resp.Write(bw.Buffer.Bytes())
	ctx.Resp.Write([]byte(")"))
}
