package router

import "github.com/cosiner/roboot"

type groupRouter struct {
	prefix string
	roboot.Router
}

func Group(r roboot.Router, pathPrefix string) roboot.Router {
	return groupRouter{
		prefix: pathPrefix,
		Router: r,
	}
}

func (g groupRouter) Handle(path string, handler roboot.Handler) error {
	return g.Router.Handle(g.prefix+path, handler)
}

func (g groupRouter) Filter(path string, filters ...roboot.Filter) error {
	return g.Router.Filter(g.prefix+path, filters...)
}

func (g groupRouter) Group(prefix string) roboot.Router {
	return groupRouter{
		prefix: g.prefix + prefix,
		Router: g.Router,
	}
}

func (g groupRouter) Merge(prefix string, r roboot.Router) error {
	return g.Router.Merge(g.prefix+prefix, r)
}

func (g groupRouter) MatchHandler(path string) roboot.MatchedHandler {
	return g.Router.MatchHandler(g.prefix + path)
}

func (g groupRouter) MatchFilters(path string) []roboot.MatchedFilter {
	return g.Router.MatchFilters(g.prefix + path)
}

func (g groupRouter) MatchHandlerAndFilters(path string) (roboot.MatchedHandler, []roboot.MatchedFilter) {
	return g.Router.MatchHandlerAndFilters(g.prefix + path)
}
