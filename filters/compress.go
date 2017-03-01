package filters

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/cosiner/roboot"
)

var errHijack = errors.New("Response is not hijackable")

type gzipWriter struct {
	gw *gzip.Writer
	roboot.ResponseWriter
}

func (w gzipWriter) Write(data []byte) (int, error) {
	return w.gw.Write(data)
}

func (w gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, is := w.ResponseWriter.(http.Hijacker)
	if !is {
		return nil, nil, errHijack
	}

	w.gw.Close()

	return hijacker.Hijack()
}

func (w gzipWriter) Close() error {
	err := w.gw.Close()

	return err
}

type flateWriter struct {
	fw *flate.Writer
	roboot.ResponseWriter
}

func (w flateWriter) Write(data []byte) (int, error) {
	return w.fw.Write(data)
}

func (w flateWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, is := w.ResponseWriter.(http.Hijacker)
	if !is {
		return nil, nil, errHijack
	}

	w.fw.Close()

	return hijacker.Hijack()
}

func (w flateWriter) Close() error {
	err := w.fw.Close()

	return err
}

func gzipCompress(w roboot.ResponseWriter) (roboot.ResponseWriter, bool) {
	w.Header().Set(roboot.HeaderContentEncoding, roboot.ContentEncodingGzip)
	return gzipWriter{
		gw:             gzip.NewWriter(w),
		ResponseWriter: w,
	}, true
}

func flateCompress(w roboot.ResponseWriter) (roboot.ResponseWriter, bool) {
	fw, err := flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return w, false
	}

	w.Header().Set(roboot.HeaderContentEncoding, roboot.ContentEncodingDeflate)
	return flateWriter{
		fw:             fw,
		ResponseWriter: w,
	}, true
}

func Compress(ctx *roboot.Context, chain roboot.HandlerFunc) {
	encoding := ctx.Req.Header.Get(roboot.HeaderAcceptEncoding)

	var (
		needClose bool
		oldW      = ctx.Resp
	)
	if strings.Contains(encoding, roboot.ContentEncodingGzip) {
		ctx.Resp, needClose = gzipCompress(oldW)
	} else if strings.Contains(encoding, roboot.ContentEncodingDeflate) {
		ctx.Resp, needClose = flateCompress(oldW)
	}

	chain(ctx)
	if needClose {
		ctx.Resp.Header().Del(roboot.HeaderContentLength)
		cw, ok := ctx.Resp.(io.Closer)
		if ok {
			cw.Close()
		}
	}
	ctx.Resp = oldW
}
