package server

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type HandlerBuilder func(http.Handler) http.Handler

func ChainMiddlewares(base http.Handler, links ...HandlerBuilder) http.Handler {
	slices.Reverse(links)
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
	return ChainMiddlewares(h,
		hlog.NewHandler(log.Logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Send()
		}),
		hlog.RemoteAddrHandler("ip"),
		hlog.UserAgentHandler("user-agent"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("request-id", "Request-Id"),
	)
}

// func TokenOrAuthMiddleware(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// First we need the auth layer to use
// 		authLayer := AuthFromRequestContext(w, r)
// 		if authLayer == nil {
// 			return
// 		}
// 		store := StorageFromRequest(w, r)
// 		if store == nil {
// 			return
// 		}
// 		log := LogFromRequestContext(w, r)
// 		if log == nil {
// 			return
// 		}
// 		log.Debugln("TokenOrAuthMiddleware called")
// 		// For the authentication, check the existence of a token first
// 		// If there is a token, ignore basic auth and fail if the token is false
// 		token := r.Header.Get(auth.AUTH_TOKEN_HEADER)
// 		if strings.TrimPrefix(token, "Bearer ") != "" {
// 			accId, ok := authLayer.CheckToken(token)
// 			if !ok {
// 				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
// 				return
// 			} else {
// 				h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CONTEXT_KEY_ACTOR_ID, accId)))
// 				return
// 			}
// 		}
// 		// No token found, try basic auth
// 		username, password, authSet := r.BasicAuth()
// 		if !authSet {
// 			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
// 			return
// 		}
// 		// If the account requires mfa, unlucky. Better generate a token for next time
// 		if status, _ := authLayer.LoginWithPassword(username, password); status != auth.AUTH_SUCCESS {
// 			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
// 			return
// 		}
//
// 		acc, err := store.FindAccountByName(username)
// 		if err != nil {
// 			log.WithError(err).
// 				WithField("middleware", "authentication").
// 				Warningln("Completed authentication but failed to get account afterwards")
// 			http.Error(
// 				w,
// 				"Failed to get account after authentication",
// 				http.StatusInternalServerError,
// 			)
// 		}
// 		if !acc.Approved {
// 			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
// 			return
// 		}
//
// 		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CONTEXT_KEY_ACTOR_ID, acc.ID)))
// 	})
// }

func CanApproveNotesOnlyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accId := AccIdFromRequestContext(w, r)
		if accId == nil {
			return
		}
		store := StorageFromRequest(w, r)
		if store == nil {
			return
		}
		log := hlog.FromRequest(r)
		acc, err := store.FindAccountByID(*accId)
		if err != nil {
			log.Warn().Err(err).
				Msg("Failed to get account from id after acc is already verified")
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
		if !(acc.Approved && acc.CanApprovePlugins) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
func CanApproveUsersOnlyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accId := AccIdFromRequestContext(w, r)
		if accId == nil {
			return
		}
		store := StorageFromRequest(w, r)
		if store == nil {
			return
		}
		log := hlog.FromRequest(r)
		acc, err := store.FindAccountByID(*accId)
		if err != nil {
			log.Warn().Err(err).
				Msg("Failed to get account from id after acc is already verified")
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
		if !(acc.Approved && acc.CanApproveUsers) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func RouteBasedLoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		newRequest := r.WithContext(context.WithValue(
			ctx,
			CONTEXT_KEY_LOG,
			log.With().Str("url-path", r.URL.Path).Logger(),
		))
		h.ServeHTTP(w, newRequest)
	})
}
