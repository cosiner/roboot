package filters

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/cosiner/roboot"
)

type compressWriter struct {
	hijacked bool
	cw       io.WriteCloser
	roboot.ResponseWriter
}

func (w *compressWriter) Write(data []byte) (int, error) {
	return w.cw.Write(data)
}

func (w *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, is := w.ResponseWriter.(http.Hijacker)
	if !is {
		return nil, nil, roboot.ErrHijack
	}

	w.hijacked = true
	w.ResponseWriter.Header().Del(roboot.HeaderContentEncoding)

	return hijacker.Hijack()
}

func (w *compressWriter) Close() error {
	if !w.hijacked {
		return w.cw.Close()
	}

	return nil
}

func gzipCompress(w roboot.ResponseWriter) (roboot.ResponseWriter, bool) {
	w.Header().Set(roboot.HeaderContentEncoding, roboot.ContentEncodingGzip)
	return &compressWriter{
		cw:             gzip.NewWriter(w),
		ResponseWriter: w,
	}, true
}

func flateCompress(w roboot.ResponseWriter) (roboot.ResponseWriter, bool) {
	fw, err := flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return w, false
	}

	w.Header().Set(roboot.HeaderContentEncoding, roboot.ContentEncodingDeflate)
	return &compressWriter{
		cw:             fw,
		ResponseWriter: w,
	}, true
}

func Compress(ctx *roboot.Context, chain roboot.Handler) {
	encoding := ctx.Req.Header.Get(roboot.HeaderAcceptEncoding)

	var (
		compressed bool
		oldW       = ctx.Resp
	)
	if strings.Contains(encoding, roboot.ContentEncodingGzip) {
		ctx.Resp, compressed = gzipCompress(oldW)
	} else if strings.Contains(encoding, roboot.ContentEncodingDeflate) {
		ctx.Resp, compressed = flateCompress(oldW)
	}

	chain.Handle(ctx)
	if compressed {
		ctx.Resp.Header().Del(roboot.HeaderContentLength)
		cw, ok := ctx.Resp.(io.Closer)
		if ok {
			cw.Close()
		}
	}
	ctx.Resp = oldW
}

var _ roboot.FilterFunc = Compress
