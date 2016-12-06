package router

import (
	"errors"
	"fmt"

	"github.com/cosiner/roboot"
	"github.com/cosiner/router"
)

type routeHandler struct {
	handler roboot.Handler
	filters []roboot.Filter
}

type serverRouter struct {
	router router.Tree
}

func New() roboot.Router {
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

func (s *serverRouter) Handle(path string, handler roboot.Handler) error {
	return s.addRoute(path, func(hd routeHandler) (routeHandler, error) {
		if hd.handler != nil {
			return hd, fmt.Errorf("duplicate route handler: %s", path)
		}
		hd.handler = handler
		return hd, nil
	})
}

func (s *serverRouter) Filter(path string, filters ...roboot.Filter) error {
	return s.addRoute(path, func(hd routeHandler) (routeHandler, error) {
		if hd.handler != nil {
			return hd, fmt.Errorf("duplicate route handler: %s", path)
		}
		c := cap(hd.filters)
		if c == 0 {
			hd.filters = filters
		} else if c-len(hd.filters) < len(filters) {
			newFilters := make([]roboot.Filter, len(hd.filters)+len(filters))
			copy(newFilters, hd.filters)
			copy(newFilters[len(hd.filters):], filters)
			hd.filters = newFilters
		} else {
			hd.filters = append(hd.filters, filters...)
		}
		return hd, nil
	})
}

func (s *serverRouter) Group(prefix string) roboot.Router {
	return Group(s, prefix)
}

func (s *serverRouter) Merge(path string, r roboot.Router) error {
	switch sr := r.(type) {
	case *serverRouter:
		return s.router.Add(path, sr.router)
	case groupRouter:
		return errors.New("grouped router already be handled and should not to be handled again")
	default:
		return errors.New("unsupported rotuer type")
	}
}

func (s *serverRouter) parseMatchedHandler(result router.MatchResult) roboot.MatchedHandler {
	if result.Handler == nil {
		return roboot.MatchedHandler{}
	}
	return roboot.MatchedHandler{
		Handler: result.Handler.(routeHandler).handler,
		Params:  result.KeyValues,
	}
}

func (s *serverRouter) MatchHandler(path string) roboot.MatchedHandler {
	result := s.router.MatchOne(path)
	return s.parseMatchedHandler(result)
}

func (s *serverRouter) parseMatchedFilters(results []router.MatchResult) []roboot.MatchedFilter {
	filters := make([]roboot.MatchedFilter, 0, len(results))
	for i := range results {
		for _, filter := range results[i].Handler.(routeHandler).filters {
			filters = append(filters, roboot.MatchedFilter{
				Filter: filter,
				Params: results[i].KeyValues,
			})
		}
	}
	return filters
}

func (s *serverRouter) MatchFilters(path string) []roboot.MatchedFilter {
	results := s.router.MatchAll(path)
	return s.parseMatchedFilters(results)
}

func (s *serverRouter) MatchHandlerAndFilters(path string) (roboot.MatchedHandler, []roboot.MatchedFilter) {
	h, f := s.router.MatchBoth(path)
	return s.parseMatchedHandler(h), s.parseMatchedFilters(f)
}
