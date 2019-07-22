package roboot

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

//==============================================================================
//                                Error
//==============================================================================
type errorString string

func (e errorString) Error() string { return string(e) }
func newError(s string) error {
	return errorString(s)
}

//==============================================================================
//                                Env
//==============================================================================
type (
	Encoder interface {
		Encode(interface{}) error
	}

	Decoder interface {
		Decode(interface{}) error
	}

	Codec interface {
		ContentType() string
		Marshal(interface{}) ([]byte, error)
		NewEncoder(io.Writer) Encoder
		Encode(io.Writer, interface{}) error
		Unmarshal([]byte, interface{}) error
		NewDecoder(io.Reader) Decoder
		Decode(io.Reader, interface{}) error
	}

	Renderer interface {
		Render(io.Writer, string, interface{}) error
	}

	ErrType uint8

	ErrorHandler interface {
		Log(ctx *Context, errType ErrType, err error)
		Handle(ctx *Context, callerDepth, status int, err error)
	}

	Env struct {
		// cant't be nil
		Codec Codec
		Error ErrorHandler

		FileUpload struct {
			MaxMemory int64
		}
		Renderer Renderer
	}
)

const (
	ErrTypeParseQuery ErrType = iota + 1
	ErrTypeParseForm
	ErrTypeParseMultipartForm
	ErrTypePanic
	ErrTypeHandle
	ErrTypeRender
	ErrTypeEncode
)

func (e ErrType) String() string {
	switch e {
	case ErrTypeParseQuery:
		return "ParseQuery"
	case ErrTypeParseForm:
		return "ParseForm"
	case ErrTypeParseMultipartForm:
		return "ParseMultipartForm"
	case ErrTypePanic:
		return "Panic"
	case ErrTypeHandle:
		return "Handle"
	case ErrTypeRender:
		return "Render"
	case ErrTypeEncode:
		return "Encode"
	default:
		return "Unknown"
	}
}

//==============================================================================
//                                Context
//==============================================================================
type (
	ResponseWriter interface {
		StatusCode() int
		http.ResponseWriter
	}

	respWriter struct {
		statusCode int
		http.ResponseWriter
	}

	Params interface {
		Len() int
		Get(name string) string
	}

	Context struct {
		Req   *http.Request
		Resp  ResponseWriter
		Codec Codec

		httpW     http.ResponseWriter
		env       *Env
		encoder   Encoder
		decoder   Decoder
		urlQuery  url.Values
		urlParams Params
		ctxValues map[string]interface{}
	}
)

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

func (ctx *Context) Env() *Env {
	return ctx.env
}

func (ctx *Context) OriginalHTTPResponseWriter() http.ResponseWriter {
	return ctx.httpW
}

func (ctx *Context) ParamValue(name string) string {
	return ctx.urlParams.Get(name)
}

func (ctx *Context) queryValues() url.Values {
	if ctx.urlQuery == nil {
		params, err := url.ParseQuery(ctx.Req.URL.RawQuery)
		if err != nil {
			ctx.Env().Error.Log(ctx, ErrTypeParseQuery, err)
		}
		ctx.urlQuery = params
	}
	return ctx.urlQuery
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
			ctx.Env().Error.Log(ctx, ErrTypeParseForm, err)
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

func (ctx *Context) ContextValue(name string) interface{} {
	return ctx.ctxValues[name]
}

func (ctx *Context) SetContextValue(name string, val interface{}) {
	if ctx.ctxValues == nil {
		ctx.ctxValues = make(map[string]interface{})
	}
	ctx.ctxValues[name] = val
}

func (ctx *Context) multipartFormValues() *multipart.Form {
	const defaultMaxMemory = 32 << 20 // 32M
	if ctx.Req.MultipartForm == nil {
		tmpQuery := ctx.Req.URL.RawQuery
		ctx.Req.URL.RawQuery = ""

		maxMemory := ctx.Env().FileUpload.MaxMemory
		if maxMemory <= 0 {
			maxMemory = defaultMaxMemory
		}
		err := ctx.Req.ParseMultipartForm(maxMemory)
		if err != nil {
			ctx.Env().Error.Log(ctx, ErrTypeParseMultipartForm, err)
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

func (ctx *Context) GetCodec() Codec {
	if ctx.Codec != nil {
		return ctx.Codec
	}
	return ctx.Env().Codec
}

func (ctx *Context) Decode(obj interface{}) error {
	if ctx.decoder == nil {
		ctx.decoder = ctx.GetCodec().NewDecoder(ctx.Req.Body)
	}
	return ctx.decoder.Decode(obj)
}

func (ctx *Context) Status(code int) {
	ctx.Resp.WriteHeader(code)
}

func (ctx *Context) Encode(obj interface{}, status int) {
	codec := ctx.GetCodec()
	if ctx.encoder == nil {
		ctx.encoder = codec.NewEncoder(ctx.Resp)
	}
	typ := ctx.Resp.Header().Get(HeaderContentType)
	if typ == "" {
		ctx.Resp.Header().Set(HeaderContentType, codec.ContentType())
	}
	if status == 0 {
		status = http.StatusOK
	}
	ctx.Status(status)
	err := ctx.encoder.Encode(obj)
	if err != nil {
		ctx.env.Error.Log(ctx, ErrTypeEncode, err)
	}
}

var (
	errEmptyRenderer = newError("renderer is empty")
)

func (ctx *Context) Render(name string, v interface{}) {
	renderer := ctx.Env().Renderer
	if renderer == nil {
		ctx.Error(errEmptyRenderer, http.StatusInternalServerError)
		return
	}
	err := renderer.Render(ctx.Resp, name, v)
	if err != nil {
		ctx.env.Error.Log(ctx, ErrTypeRender, err)
	}
}

func (ctx *Context) Error(err error, statusCode int) {
	if err == nil {
		panic(fmt.Errorf("expect non-nil error"))
	}
	ctx.Env().Error.Handle(ctx, 1, statusCode, err)
}

//==============================================================================
//                                Handler
//==============================================================================
type (
	Handler interface {
		Handle(*Context)
	}

	HandlerFunc func(*Context)

	Filter interface {
		Filter(ctx *Context, chain Handler)
	}

	FilterFunc func(ctx *Context, chain Handler)

	MatchedHandler struct {
		Handler
		Params
	}

	MatchedFilter struct {
		Filter
		Params
	}

	Router interface {
		Handle(path string, handler Handler) error
		Filter(path string, filters ...Filter) error
		Group(prefix string) Router
		Merge(prefix string, r Router) error

		MatchHandler(path string) MatchedHandler
		MatchFilters(path string) []MatchedFilter
		MatchHandlerAndFilters(path string) (MatchedHandler, []MatchedFilter)
	}
)

func (f HandlerFunc) Handle(ctx *Context) {
	f(ctx)
}

func (f FilterFunc) Filter(ctx *Context, chain Handler) {
	f(ctx, chain)
}

//==============================================================================
//                                Server
//==============================================================================
type (
	Server interface {
		Env() *Env
		Router(h string) Router
		Host(h string, r Router)
		http.Handler
	}

	server struct {
		defaultRouter Router
		routers       map[string]Router

		env Env
	}

	filterHandler struct {
		filters []MatchedFilter
		handler MatchedHandler
	}
)

func (f *filterHandler) Handle(ctx *Context) {
	oldP := ctx.urlParams
	if len(f.filters) == 0 {
		ctx.urlParams = f.handler.Params
		f.handler.Handler.Handle(ctx)
	} else {
		filter := f.filters[0]
		f.filters = f.filters[1:]
		ctx.urlParams = filter.Params
		filter.Filter.Filter(ctx, f)
	}
	ctx.urlParams = oldP
}

func NewServer(env Env, defaultRouter Router) Server {
	if env.Error == nil || env.Codec == nil {
		panic("error handler and codec should not be empty")
	}

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

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := Context{
		Req:   req,
		Resp:  &respWriter{ResponseWriter: w},
		env:   &s.env,
		httpW: w,
	}
	r := s.Router(req.URL.Host)
	if r == nil {
		ctx.Error(newError("resource not found"), http.StatusNotFound)
		return
	}

	handler, filters := r.MatchHandlerAndFilters(req.URL.Path)
	if handler.Handler == nil {
		handler.Handler = HandlerFunc(func(ctx *Context) {
			ctx.Error(newError("resource not found"), http.StatusNotFound)
		})
	}
	(&filterHandler{
		filters: filters,
		handler: handler,
	}).Handle(&ctx)
}
