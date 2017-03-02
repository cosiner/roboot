package roboot

import (
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/cosiner/httperrs"
)

type Params interface {
	Len() int
	Get(name string) string
}

type Logger interface {
	Error(...interface{})
	Errorf(string, ...interface{})
}

type stdLoggerT struct{}

var stdLogger Logger = stdLoggerT{}

func (stdLoggerT) Error(v ...interface{}) {
	log.Println(v...)
}

func (stdLoggerT) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

type Encoder interface {
	Encode(interface{}) error
}

type Decoder interface {
	Decode(interface{}) error
}

type Codec interface {
	ContentType() string
	Marshal(interface{}) ([]byte, error)
	NewEncoder(io.Writer) Encoder
	Encode(io.Writer, interface{}) error
	Unmarshal([]byte, interface{}) error
	NewDecoder(io.Reader) Decoder
	Decode(io.Reader, interface{}) error
}

type Renderer interface {
	Render(io.Writer, string, interface{}) error
}

type ErrorHandler interface {
	Handle(ctx *Context, status int, err error)
}

type ErrorHandlerFunc func(*Context, int, error)

func (e ErrorHandlerFunc) Handle(ctx *Context, status int, err error) {
	e(ctx, status, err)
}

var _ ErrorHandler = ErrorHandlerFunc(nil)

type Env struct {
	FileUpload struct {
		MaxMemory int64
	}
	Logger
	Codec
	Renderer
	ErrorHandler
}

func (e *Env) GetLogger() Logger {
	logger := e.Logger
	if logger != nil {
		return logger
	}
	return stdLogger
}

type respWriter struct {
	statusCode int
	http.ResponseWriter
}

func (r *respWriter) WriteHeader(status int) {
	if r.statusCode > 0 {
		return
	}
	r.ResponseWriter.WriteHeader(status)
	r.statusCode = status
}

func (r *respWriter) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	if err == nil && r.statusCode == 0 {
		r.statusCode = http.StatusOK
	}
	return n, err
}

func (r *respWriter) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

type ResponseWriter interface {
	StatusCode() int
	http.ResponseWriter
}

type Context struct {
	Params Params
	Req    *http.Request
	Resp   ResponseWriter
	Env    *Env
	Codec  Codec

	encoder Encoder
	decoder Decoder
	query   url.Values
}

func (ctx *Context) queryValues() url.Values {
	if ctx.query == nil {
		params, err := url.ParseQuery(ctx.Req.URL.RawQuery)
		if err != nil {
			ctx.Env.GetLogger().Errorf("Roboot.Request: parse query failed: %s: %s", ctx.Req.URL.String(), err.Error())
		}
		ctx.query = params
	}
	return ctx.query
}

func (ctx *Context) QueryValue(name string) string {
	return ctx.queryValues().Get(name)
}

func (ctx *Context) QueryValues(name string) []string {
	return ctx.queryValues()[name]
}

func (ctx *Context) bodyValues() url.Values {
	if ctx.Req.PostForm == nil {
		tmpQuery := ctx.Req.URL.RawQuery
		ctx.Req.URL.RawQuery = ""
		err := ctx.Req.ParseForm()
		if err != nil {
			ctx.Env.GetLogger().Errorf("Roboot.Request: parse post form failed: %s: %s", ctx.Req.URL.String(), err.Error())
		}
		ctx.Req.URL.RawQuery = tmpQuery
	}
	return ctx.Req.PostForm
}

func (ctx *Context) BodyValue(name string) string {
	return ctx.bodyValues().Get(name)
}

func (ctx *Context) BodyValues(name string) []string {
	return ctx.bodyValues()[name]
}

func (ctx *Context) multipartFormValues() *multipart.Form {
	if ctx.Req.MultipartForm == nil {
		tmpQuery := ctx.Req.URL.RawQuery
		ctx.Req.URL.RawQuery = ""

		maxMemory := ctx.Env.FileUpload.MaxMemory
		if maxMemory <= 0 {
			maxMemory = 32 << 20 //32M
		}
		err := ctx.Req.ParseMultipartForm(maxMemory)
		if err != nil {
			ctx.Env.GetLogger().Errorf("Roboot.Request: parse multipart form failed: %s: %s", ctx.Req.URL.String(), err.Error())
		}
		ctx.Req.URL.RawQuery = tmpQuery
	}
	return ctx.Req.MultipartForm
}

func (ctx *Context) File(name string) *multipart.FileHeader {
	files := ctx.Files(name)
	if len(files) == 0 {
		return nil
	}
	return files[0]
}

func (ctx *Context) Files(name string) []*multipart.FileHeader {
	form := ctx.multipartFormValues()
	if form == nil || form.File == nil {
		return nil
	}
	return form.File[name]
}

var (
	ErrEmptyCodec = errors.New("codec is empty")
)

func (ctx *Context) GetCodec() Codec {
	if ctx.Codec != nil {
		return ctx.Codec
	}
	return ctx.Env.Codec
}

func (ctx *Context) Decode(obj interface{}) error {
	if ctx.decoder == nil {
		codec := ctx.GetCodec()
		if codec == nil {
			return ErrEmptyCodec
		}
		ctx.decoder = codec.NewDecoder(ctx.Req.Body)
	}
	return ctx.decoder.Decode(obj)
}

func (ctx *Context) Status(code int) {
	ctx.Resp.WriteHeader(code)
}

func (ctx *Context) Encode(obj interface{}, status int) error {
	codec := ctx.GetCodec()
	if ctx.encoder == nil {
		if codec == nil {
			return ErrEmptyCodec
		}
		ctx.encoder = codec.NewEncoder(ctx.Resp)
	}
	typ := ctx.Resp.Header().Get(HeaderContentType)
	if typ == "" {
		ctx.Resp.Header().Set(HeaderContentType, codec.ContentType())
	}
	if status == 0 {
		status = http.StatusOK
	}
	ctx.Resp.WriteHeader(status)
	return ctx.encoder.Encode(obj)
}

var (
	ErrEmptyRenderer = errors.New("renderer is empty")
)

func (ctx *Context) Render(name string, v interface{}) error {
	renderer := ctx.Env.Renderer
	if renderer == nil {
		return ErrEmptyRenderer
	}
	return renderer.Render(ctx.Resp, name, v)
}

func (ctx *Context) Error(err error) {
	type errorInfo struct {
		Error string `json:"error"`
	}
	statusCode := httperrs.StatusCode(err, http.StatusInternalServerError)
	if ctx.Env.ErrorHandler == nil {
		if statusCode >= http.StatusInternalServerError {
			ctx.Env.GetLogger().Error("server failed:", err.Error())
			ctx.Status(statusCode)
		} else {
			ctx.Encode(errorInfo{Error: err.Error()}, statusCode)
		}
	} else {
		ctx.Env.ErrorHandler.Handle(ctx, statusCode, err)
	}
}

type Handler interface {
	Handle(*Context)
}

type HandlerFunc func(*Context)

func (f HandlerFunc) Handle(ctx *Context) {
	f(ctx)
}

type Filter interface {
	Filter(ctx *Context, chain HandlerFunc)
}

type FilterFunc func(ctx *Context, chain HandlerFunc)

type MatchedHandler struct {
	Handler
	Params
}

type MatchedFilter struct {
	Filter
	Params
}

type Router interface {
	Handle(path string, handler Handler) error
	Filter(path string, filters ...Filter) error
	Group(prefix string) Router
	Merge(prefix string, r Router) error

	MatchHandler(path string) MatchedHandler
	MatchFilters(path string) []MatchedFilter
	MatchHandlerAndFilters(path string) (MatchedHandler, []MatchedFilter)
}

type Server interface {
	Env() *Env
	Router(h string) Router
	Host(h string, r Router)
	http.Handler
}

type server struct {
	defaultRouter Router
	routers       map[string]Router

	env Env
}

func NewServer(env Env, defaultRouter Router) Server {
	return &server{
		defaultRouter: defaultRouter,

		env: env,
	}
}

func (s *server) Env() *Env {
	return &s.env
}

func (s *server) Router(host string) Router {
	r, has := s.routers[host]
	if !has || r == nil {
		r = s.defaultRouter
	}
	return r
}

func (s *server) Host(host string, r Router) {
	if host == "" {
		s.defaultRouter = r
	} else {
		if s.routers == nil {
			s.routers = make(map[string]Router)
		}
		s.routers[host] = r
	}
}

type filterHandler struct {
	filters []MatchedFilter
	handler MatchedHandler
}

func (f *filterHandler) Handle(ctx *Context) {
	oldP := ctx.Params
	if len(f.filters) == 0 {
		ctx.Params = f.handler.Params
		f.handler.Handler.Handle(ctx)
	} else {
		filter := f.filters[0]
		f.filters = f.filters[1:]
		ctx.Params = filter.Params
		filter.Filter.Filter(ctx, f.Handle)
	}
	ctx.Params = oldP
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := Context{
		Req:  req,
		Resp: &respWriter{ResponseWriter: w},
		Env:  &s.env,
	}
	r := s.Router(req.URL.Host)
	if r == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	handler, filters := r.MatchHandlerAndFilters(req.URL.Path)
	if handler.Handler == nil {
		handler.Handler = HandlerFunc(func(ctx *Context) {
			ctx.Status(http.StatusNotFound)
		})
	}
	h := filterHandler{
		filters: filters,
		handler: handler,
	}

	h.Handle(&ctx)
}
