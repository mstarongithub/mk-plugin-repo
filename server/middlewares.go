package server

import (
	"context"
	"net/http"

	"github.com/mstarongithub/mk-plugin-repo/auth"
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

func TokenOrAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First we need the auth layer to use
		authLayer := AuthFromRequestContext(r)
		if authLayer == nil {
			http.Error(w, "missing auth reference in request", http.StatusInternalServerError)
			return
		}
		// For the authentication, check the existence of a token first
		// If there is a token, ignore basic auth and fail if the token is false
		token := r.Header.Get(auth.AUTH_TOKEN_HEADER)
		if token != "" {
			if !authLayer.CheckToken(token) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			} else {
				h.ServeHTTP(w, r)
				return
			}
		}
		// No token found, try basic auth
		username, password, authSet := r.BasicAuth()
		if !authSet {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		// If the account requires mfa, unlucky. Better generate a token for next time
		if status, _ := authLayer.LoginWithPassword(username, password); status != auth.AUTH_SUCCESS {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
