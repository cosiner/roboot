package roboot

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/cosiner/httperrs"
)

//======================================================================================================================
//    Env
//

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
)

type (
	Renderer interface {
		Render(io.Writer, string, interface{}) error
	}
)

type (
	ErrType uint8

	ErrorHandler interface {
		Log(ctx *Context, errType ErrType, err error)
		Handle(ctx *Context, status int, err error)
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
	default:
		return "Unknown"
	}
}

//======================================================================================================================
//   Context
//

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

		env     *Env
		encoder Encoder
		decoder Decoder
		query   url.Values
		params  Params
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

func (ctx *Context) ParamValue(name string) string {
	return ctx.params.Get(name)
}

func (ctx *Context) queryValues() url.Values {
	if ctx.query == nil {
		params, err := url.ParseQuery(ctx.Req.URL.RawQuery)
		if err != nil {
			ctx.Env().Error.Log(ctx, ErrTypeParseQuery, err)
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

func (ctx *Context) Encode(obj interface{}, status int) error {
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
	return ctx.encoder.Encode(obj)
}

var (
	ErrEmptyRenderer = errors.New("renderer is empty")
)

func (ctx *Context) Render(name string, v interface{}) error {
	renderer := ctx.Env().Renderer
	if renderer == nil {
		return ErrEmptyRenderer
	}
	return renderer.Render(ctx.Resp, name, v)
}

func (ctx *Context) Error(err error) {
	ctx.Env().Error.Handle(ctx, httperrs.StatusCode(err, http.StatusInternalServerError), err)
}

//======================================================================================================================
//   Handler
//

type (
	Handler interface {
		Handle(*Context)
	}

	HandlerFunc func(*Context)

	Filter interface {
		Filter(ctx *Context, chain HandlerFunc)
	}

	FilterFunc func(ctx *Context, chain HandlerFunc)

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

//======================================================================================================================
//   Server
//

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
	oldP := ctx.params
	if len(f.filters) == 0 {
		ctx.params = f.handler.Params
		f.handler.Handler.Handle(ctx)
	} else {
		filter := f.filters[0]
		f.filters = f.filters[1:]
		ctx.params = filter.Params
		filter.Filter.Filter(ctx, f.Handle)
	}
	ctx.params = oldP
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

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := Context{
		Req:  req,
		Resp: &respWriter{ResponseWriter: w},
		env:  &s.env,
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
	(&filterHandler{
		filters: filters,
		handler: handler,
	}).Handle(&ctx)
}
