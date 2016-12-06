package handlers

import (
	"sync"

	"github.com/cosiner/roboot"
)

type cachedErrors struct {
	mu       sync.RWMutex
	handlers map[int]roboot.Handler
	base     roboot.ErrorHandler
}

func (e *cachedErrors) Handler(status int) roboot.Handler {
	e.mu.RLock()
	handler := e.handlers[status]
	e.mu.RUnlock()
	if handler != nil {
		return handler
	}

	e.mu.Lock()
	handler = e.handlers[status]
	if handler == nil {
		if e.handlers == nil {
			e.handlers = make(map[int]roboot.Handler)
		}
		handler = e.base.Handler(status)
		e.handlers[status] = handler
	}
	e.mu.Unlock()
	return handler
}

func StatusError(status int) roboot.Handler {
	return Status(status)
}

func CacheError(base roboot.ErrorHandlerFunc) roboot.ErrorHandler {
	if base == nil {
		return roboot.ErrorHandlerFunc(StatusError)
	}

	return &cachedErrors{
		base: base,
	}
}
