package server

import (
	"context"
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/sirupsen/logrus"
	"gitlab.com/mstarongitlab/weblogger"
)

type HandlerBuilder func(http.Handler) http.Handler

func ChainMiddlewares(base http.Handler, links ...HandlerBuilder) http.Handler {
	for _, f := range links {
		base = f(base)
	}
	return base
}

func ContextValsMiddleware(pairs map[any]any) HandlerBuilder {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for key, val := range pairs {
				ctx = context.WithValue(ctx, key, val)
			}
			newRequest := r.WithContext(ctx)
			h.ServeHTTP(w, newRequest)
		})
	}
}

func WebLoggerWrapper(h http.Handler) http.Handler {
	return weblogger.LoggingMiddleware(
		h,
		&weblogger.Config{
			DefaultLogLevel:    weblogger.LOG_LEVEL_DEBUG,
			FailedRequestLevel: weblogger.LOG_LEVEL_WARN,
		},
	)
}

func NosurfCheckWrapper(h http.Handler) http.Handler {
	n := nosurf.New(h)
	n.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("reason", nosurf.Reason(r)).Warnln("Failed to validate CSRF token")
		w.WriteHeader(http.StatusBadRequest)
	}))
	return n
}

func NosurfTokenInsertMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newReq := r.WithContext(
			context.WithValue(r.Context(), CONTEXT_KEY_CSRF_TOKEN, nosurf.Token(r)),
		)
		w.Header().Set("CSRF-Token", nosurf.Token(r))
		h.ServeHTTP(w, newReq)
	})
}
