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

const (
	HEADER_ACCEPT_ENCODING  = "Accept-Encoding"
	HEADER_CONTENT_ENCODING = "Content-Encoding"
	HEADER_CONTENT_LENGTH   = "Content-Length"

	CONTENT_ENCODING_GZIP    = "gzip"
	CONTENT_ENCODING_DEFLATE = "deflate"
)

type gzipWriter struct {
	gw *gzip.Writer
	http.ResponseWriter
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
	http.ResponseWriter
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

func gzipCompress(w http.ResponseWriter) (http.ResponseWriter, bool) {
	w.Header().Set(HEADER_CONTENT_ENCODING, CONTENT_ENCODING_GZIP)
	return gzipWriter{
		gw:             gzip.NewWriter(w),
		ResponseWriter: w,
	}, true
}

func flateCompress(w http.ResponseWriter) (http.ResponseWriter, bool) {
	fw, err := flate.NewWriter(w, flate.DefaultCompression)
	if err != nil {
		return w, false
	}

	w.Header().Set(HEADER_CONTENT_ENCODING, CONTENT_ENCODING_DEFLATE)
	return flateWriter{
		fw:             fw,
		ResponseWriter: w,
	}, true
}

func Compress(ctx *roboot.Context, chain roboot.HandlerFunc) {
	encoding := ctx.Req.Header.Get(HEADER_ACCEPT_ENCODING)

	var (
		needClose bool
		oldW      = ctx.Resp
	)
	if strings.Contains(encoding, CONTENT_ENCODING_GZIP) {
		ctx.Resp, needClose = gzipCompress(oldW)
	} else if strings.Contains(encoding, CONTENT_ENCODING_DEFLATE) {
		ctx.Resp, needClose = flateCompress(oldW)
	}

	chain(ctx)
	if needClose {
		ctx.Resp.Header().Del(HEADER_CONTENT_LENGTH)
		cw, ok := ctx.Resp.(io.Closer)
		if ok {
			cw.Close()
		}
	}
	ctx.Resp = oldW
}
