package server

import (
	"context"
	"net/http"
)

type withContextValsMiddleware struct {
	pairs map[any]any
	nextz http.Handler
}

type HandlerBuilder func(http.Handler) http.Handler

func addContextValsMiddleware(next http.Handler, pairs map[any]any) *withContextValsMiddleware {
	return &withContextValsMiddleware{
		pairs: pairs,
		nextz: next,
	}
}

func (c *withContextValsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	for key, val := range c.pairs {
		ctx = context.WithValue(ctx, key, val)
	}
	newRequest := r.WithContext(ctx)
	c.nextz.ServeHTTP(w, newRequest)
}
