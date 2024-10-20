package server

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/other"

	"github.com/mstarongithub/mk-plugin-repo/config"
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
		logger := hlog.FromRequest(r)
		acc, err := store.FindAccountByID(*accId)
		if err != nil {
			logger.Warn().Err(err).
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
		logger := hlog.FromRequest(r)
		acc, err := store.FindAccountByID(*accId)
		if err != nil {
			logger.Warn().Err(err).
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

func passkeyAuthInsertUidMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := StorageFromRequest(w, r)
		if s == nil {
			http.Error(w, "failed to get storage", http.StatusInternalServerError)
			return
		}
		str, ok := r.Context().Value(CONTEXT_KEY_ACTOR_NAME).(string)
		if !ok {
			http.Error(w, "actor name not in context", http.StatusInternalServerError)
			return
		}
		acc, err := s.FindAccountByPasskeyId([]byte(str))
		if err != nil {
			http.Error(w, "Failed to get account", http.StatusInternalServerError)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), CONTEXT_KEY_ACTOR_ID, acc.ID))
		h.ServeHTTP(w, r)
	})
}

func profilingAuthenticationMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("password") != config.GlobalConfig.Superuser.MetricsPassword {
			other.HttpErr(w, ErrIdNotApproved, "Bad password", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
