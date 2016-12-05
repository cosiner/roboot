package roboot

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/cosiner/router"
)

type Params interface {
	Get(name string) string
}

type Context interface {
	LogErrorf(string, ...interface{})
	FileUploadMaxMemory() int64
}

type Request interface {
	setPathParams(params Params)
	getPathParams() Params

	URL() *url.URL
	Method() string

	PathValue(string) string
	QueryValue(string) string
	QueryValues(string) []string
	BodyValue(string) string
	BodyValues(string) []string
	File(string) *multipart.FileHeader
	Files(string) []*multipart.FileHeader

	Header(string) string
	Headers(string) []string

	Body() io.Reader
}

type request struct {
	req         *http.Request
	queryParams url.Values
	pathParams  Params
	ctx         Context
}

func (req *request) setPathParams(params Params) {
	req.pathParams = params
}

func (req *request) getPathParams() Params {
	return req.pathParams
}

func (r *request) URL() *url.URL {
	return r.req.URL
}

func (r *request) Method() string {
	return r.req.Method
}

func (r *request) PathValue(name string) string {
	return r.pathParams.Get(name)
}

func (r *request) queryValues() url.Values {
	if r.queryParams == nil {
		params, err := url.ParseQuery(r.req.URL.RawQuery)
		if err != nil {
			r.ctx.LogErrorf("Roboot.Request: parse query failed: %s: %s", r.req.URL.String(), err.Error())
		}
		r.queryParams = params
	}
	return r.queryParams
}

func (r *request) QueryValue(name string) string {
	return r.queryValues().Get(name)
}

func (r *request) QueryValues(name string) []string {
	return r.queryValues()[name]
}

func (r *request) bodyValues() url.Values {
	if r.req.PostForm == nil {
		tmpQuery := r.req.URL.RawQuery
		r.req.URL.RawQuery = ""
		err := r.req.ParseForm()
		if err != nil {
			r.ctx.LogErrorf("Roboot.Request: parse post form failed: %s: %s", r.req.URL.String(), err.Error())
		}
		r.req.URL.RawQuery = tmpQuery
	}
	return r.req.PostForm
}

func (r *request) BodyValue(name string) string {
	return r.bodyValues().Get(name)
}

func (r *request) BodyValues(name string) []string {
	return r.bodyValues()[name]
}

func (r *request) multipartFormValues() *multipart.Form {
	if r.req.MultipartForm == nil {
		tmpQuery := r.req.URL.RawQuery
		r.req.URL.RawQuery = ""

		maxMemory := r.ctx.FileUploadMaxMemory()
		if maxMemory <= 0 {
			maxMemory = 32 << 20 //32M
		}
		err := r.req.ParseMultipartForm(maxMemory)
		if err != nil {
			r.ctx.LogErrorf("Roboot.Request: parse multipart form failed: %s: %s", r.req.URL.String(), err.Error())
		}
		r.req.URL.RawQuery = tmpQuery
	}
	return r.req.MultipartForm
}

func (r *request) File(name string) *multipart.FileHeader {
	files := r.Files(name)
	if len(files) == 0 {
		return nil
	}
	return files[0]
}

func (r *request) Files(name string) []*multipart.FileHeader {
	form := r.multipartFormValues()
	if form == nil || form.File == nil {
		return nil
	}
	return form.File[name]
}

func (r *request) Header(name string) string {
	return r.req.Header.Get(name)
}

func (r *request) Headers(name string) []string {
	return r.req.Header[name]
}

func (r *request) Body() io.Reader {
	return r.req.Body
}

type Response interface {
	Status(int)

	Header(string, string)
	Headers(string, ...string)

	Body() io.Writer
}

type response struct {
	w http.ResponseWriter
}

func (r *response) Status(code int) {
	r.w.WriteHeader(code)
}

func (r *response) Header(name, value string) {
	r.w.Header().Set(name, value)
}

func (r *response) Headers(name string, values ...string) {
	if len(values) == 0 {
		r.w.Header().Del(name)
	} else {
		for _, value := range values {
			r.w.Header().Add(name, value)
		}
	}
}

func (r *response) Body() io.Writer {
	return r.w
}

type Handler interface {
	Handle(req Request, res Response)
}

type HandlerFunc func(Request, Response)

func (f HandlerFunc) Handle(req Request, res Response) {
	f(req, res)
}

type Filter interface {
	Filter(req Request, res Response, chain HandlerFunc)
}

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

	ctx Context
}

func NewServer(ctx Context) Server {
	return &server{
		defaultRouter: NewRouter(),

		ctx: ctx,
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

func (f *filterHandler) Handle(req Request, res Response) {
	originParams := req.getPathParams()
	if len(f.filters) == 0 {
		req.setPathParams(f.handler.Params)
		f.handler.Handler.Handle(req, res)
	} else {
		filter := f.filters[0]
		f.filters = f.filters[1:]
		req.setPathParams(filter.Params)
		filter.Filter.Filter(req, res, f.Handle)
	}
	req.setPathParams(originParams)
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r := s.Router(req.URL.Host)
	if r == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	handler, filters := r.MatchHandlerAndFilters(req.URL.Path)
	if handler.Handler == nil {
		handler.Handler = HandlerFunc(NotFoundHandler)
	}

	sreq := request{
		req: req,
		ctx: s.ctx,
	}
	sres := response{
		w: w,
	}
	h := filterHandler{
		filters: filters,
		handler: handler,
	}
	h.Handle(&sreq, &sres)
}
