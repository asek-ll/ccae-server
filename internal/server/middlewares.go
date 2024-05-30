package server

import "net/http"

type MiddlewaresGroup struct {
	mux         *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewMiddlewareGroup(mux *http.ServeMux) *MiddlewaresGroup {
	return &MiddlewaresGroup{
		mux:         mux,
		middlewares: nil,
	}
}

func (mg *MiddlewaresGroup) wrap(handler http.Handler) http.Handler {
	for i := range mg.middlewares {
		handler = mg.middlewares[len(mg.middlewares)-1-i](handler)
	}
	return handler
}

func (mg *MiddlewaresGroup) HandleFunc(pattern string, handler http.HandlerFunc) {
	mg.mux.Handle(pattern, mg.wrap(handler))
}

func (mg *MiddlewaresGroup) Group() *MiddlewaresGroup {
	middlewares := make([]func(http.Handler) http.Handler, len(mg.middlewares))
	copy(middlewares, mg.middlewares)
	return &MiddlewaresGroup{
		mux:         mg.mux,
		middlewares: mg.middlewares,
	}
}

func (mg *MiddlewaresGroup) Use(mw func(http.Handler) http.Handler) *MiddlewaresGroup {
	mg.middlewares = append(mg.middlewares, mw)
	return mg
}
