package roboot

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/cosiner/router"
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
	Handler(int) Handler
}

type ErrorHandlerFunc func(int) Handler

func (e ErrorHandlerFunc) Handler(status int) Handler {
	return e(status)
}

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

type Context struct {
	Params Params
	Req    *http.Request
	Resp   http.ResponseWriter
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

func (ctx *Context) Encode(obj interface{}) error {
	if ctx.encoder == nil {
		codec := ctx.GetCodec()
		if codec == nil {
			return ErrEmptyCodec
		}
		ctx.encoder = codec.NewEncoder(ctx.Resp)
	}
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

func (ctx *Context) Error(status int) {
	var (
		handler Handler
		eh      = ctx.Env.ErrorHandler
	)
	if eh != nil {
		handler = eh.Handler(status)
		return
	}
	if handler == nil {
		ctx.Resp.WriteHeader(status)
		return
	}

	handler.Handle(ctx)
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

type routeHandler struct {
	handler Handler
	filters []Filter
}

type serverRouter struct {
	router router.Tree
}

func NewRouter() Router {
	return &serverRouter{}
}

func (r *serverRouter) addRoute(path string, fn func(routeHandler) (routeHandler, error)) error {
	return r.router.Add(path, func(h interface{}) (interface{}, error) {
		var (
			hd routeHandler
			ok bool
		)
		if h != nil {
			hd, ok = h.(routeHandler)
			if !ok {
				return nil, fmt.Errorf("illegal route handler type: %s", path)
			}
		}
		hd, err := fn(hd)
		return hd, err
	})
}

func (s *serverRouter) Handle(path string, handler Handler) error {
	return s.addRoute(path, func(hd routeHandler) (routeHandler, error) {
		if hd.handler != nil {
			return hd, fmt.Errorf("duplicate route handler: %s", path)
		}
		hd.handler = handler
		return hd, nil
	})
}

func (s *serverRouter) Filter(path string, filters ...Filter) error {
	return s.addRoute(path, func(hd routeHandler) (routeHandler, error) {
		if hd.handler != nil {
			return hd, fmt.Errorf("duplicate route handler: %s", path)
		}
		c := cap(hd.filters)
		if c == 0 {
			hd.filters = filters
		} else if c-len(hd.filters) < len(filters) {
			newFilters := make([]Filter, len(hd.filters)+len(filters))
			copy(newFilters, hd.filters)
			copy(newFilters[len(hd.filters):], filters)
			hd.filters = newFilters
		} else {
			hd.filters = append(hd.filters, filters...)
		}
		return hd, nil
	})
}

func (s *serverRouter) Group(prefix string) Router {
	return groupRouter{
		prefix: prefix,
		Router: s,
	}
}

func (s *serverRouter) Merge(path string, r Router) error {
	switch sr := r.(type) {
	case *serverRouter:
		return s.router.Add(path, sr.router)
	case groupRouter:
		return errors.New("grouped router already be handled and should not to be handled again")
	default:
		return errors.New("unsupported rotuer type")
	}
}

func (s *serverRouter) parseMatchedHandler(result router.MatchResult) MatchedHandler {
	if result.Handler == nil {
		return MatchedHandler{}
	}
	return MatchedHandler{
		Handler: result.Handler.(routeHandler).handler,
		Params:  result.KeyValues,
	}
}

func (s *serverRouter) MatchHandler(path string) MatchedHandler {
	result := s.router.MatchOne(path)
	return s.parseMatchedHandler(result)
}

func (s *serverRouter) parseMatchedFilters(results []router.MatchResult) []MatchedFilter {
	filters := make([]MatchedFilter, 0, len(results))
	for i := range results {
		for _, filter := range results[i].Handler.(routeHandler).filters {
			filters = append(filters, MatchedFilter{
				Filter: filter,
				Params: results[i].KeyValues,
			})
		}
	}
	return filters
}

func (s *serverRouter) MatchFilters(path string) []MatchedFilter {
	results := s.router.MatchAll(path)
	return s.parseMatchedFilters(results)
}

func (s *serverRouter) MatchHandlerAndFilters(path string) (MatchedHandler, []MatchedFilter) {
	h, f := s.router.MatchBoth(path)
	return s.parseMatchedHandler(h), s.parseMatchedFilters(f)
}

type groupRouter struct {
	prefix string
	Router
}

func (g groupRouter) Handle(path string, handler Handler) error {
	return g.Router.Handle(g.prefix+path, handler)
}

func (g groupRouter) Filter(path string, filters ...Filter) error {
	return g.Router.Filter(g.prefix+path, filters...)
}

func (g groupRouter) Group(prefix string) Router {
	return groupRouter{
		prefix: g.prefix + prefix,
		Router: g.Router,
	}
}

func (g groupRouter) Merge(prefix string, r Router) error {
	return g.Router.Merge(g.prefix+prefix, r)
}

func (g groupRouter) MatchHandler(path string) MatchedHandler {
	return g.Router.MatchHandler(g.prefix + path)
}

func (g groupRouter) MatchFilters(path string) []MatchedFilter {
	return g.Router.MatchFilters(g.prefix + path)
}

func (g groupRouter) MatchHandlerAndFilters(path string) (MatchedHandler, []MatchedFilter) {
	return g.Router.MatchHandlerAndFilters(g.prefix + path)
}

type Server interface {
	Router(h string) Router
	Host(h string, r Router)
	http.Handler
}

type server struct {
	defaultRouter Router
	routers       map[string]Router

	env *Env
}

func NewServer(env *Env) Server {
	return &server{
		defaultRouter: NewRouter(),

		env: env,
	}
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
		Resp: w,
		Env:  s.env,
	}
	r := s.Router(req.URL.Host)
	if r == nil {
		ctx.Error(http.StatusNotFound)
		return
	}

	handler, filters := r.MatchHandlerAndFilters(req.URL.Path)
	if handler.Handler == nil {
		handler.Handler = HandlerFunc(func(ctx *Context) {
			ctx.Error(http.StatusNotFound)
		})
	}
	h := filterHandler{
		filters: filters,
		handler: handler,
	}

	h.Handle(&ctx)
}
