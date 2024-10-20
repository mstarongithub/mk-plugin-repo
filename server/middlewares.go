package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/mstarongithub/passkey"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/other"
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

func forceCorrectPasskeyAuthFlowMiddleware(
	pkey *passkey.Passkey,
	handler http.Handler,
) http.Handler {
	var username struct {
		Username string `json:"username"`
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't fuck with the request if not intended for starting to register or login
		if !strings.HasSuffix(r.URL.Path, "Begin") {
			log.Debug().Msg("Request to non-begin passkey method, doing nothing")
			handler.ServeHTTP(w, r)
			return
		}
		store := StorageFromRequest(w, r)
		if store == nil {
			return
		}
		// Grab all needed data
		// First read and parse body
		body, _ := io.ReadAll(r.Body)
		log.Debug().Bytes("body", body).Msg("Body of auth begin request")
		err := json.Unmarshal(body, &username)
		if err != nil {
			other.HttpErr(w, ErrIdBadRequest, "Not a username json object", http.StatusBadRequest)
			return
		}
		// Then check if we can read the cookie
		sid, err := r.Cookie("sid")
		// If we can't read the cookie, it doesn't exist
		// Thus no user is logged in
		// NOTE: This assumption *could* cause maybe a problem if the registerBegin endpoint is called with a bearer token
		//       However, I'm pretty certain that such a token won't have any impact on the execution as of right now
		//       Since there's no point in the code before or after this moment where a bearer token is being read
		//       The only point where a bearer token is read is for calls to a login regstricted endpoint
		//       if no passkey session is around
		if err != nil {
			log.Debug().
				Msg("Request has no passkey session cookie, is fresh login/register. Checking if account with requested username already exists")
			_, err = store.FindAccountByName(username.Username)
			switch err {
			case nil:
				log.Info().
					Str("username", username.Username).
					Msg("Account with same name already exists, preventing login")
				other.HttpErr(
					w,
					ErrIdAlreadyExists,
					"Account with that name already exists",
					http.StatusBadRequest,
				)
			case storage.ErrAccountNotFound:
				// Expected case where no account with that name exists yet
				log.Debug().
					Str("username", username.Username).
					Msg("No account with this username exists yet, passing through")
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				r.ContentLength = int64(len(body))
				handler.ServeHTTP(w, r)
			default:
				log.Error().
					Err(err).
					Str("username", username.Username).
					Msg("Failed to check if account with username already exists")
				other.HttpErr(
					w,
					ErrIdDbErr,
					"Failed to check if account with that name already exists",
					http.StatusInternalServerError,
				)
			}
			return
		} else {
			log.Debug().Msg("Account is already logged in, force replacing given username with logged in's username")
			session, ok := store.GetSession(sid.Value)
			if !ok {
				log.Error().Msg("Failed to get passkey session")
				other.HttpErr(w, ErrIdDbErr, "Failed to get passkey session", http.StatusInternalServerError)
				return
			}
			if session.Expires.Before(time.Now()) {
				log.Debug().Msg("Session expired, failing")
				http.SetCookie(w, &http.Cookie{
					Name:    "sid",
					Value:   "",
					Expires: time.Now().Add(time.Hour * 24 * -7),
					MaxAge:  -5,
				})
				other.HttpErr(w, ErrIdNotApproved, "Session expired", http.StatusForbidden)
				return
			}
			acc, err := store.FindAccountByPasskeyId(session.UserID)
			switch err {
			case nil:
				log.Debug().Msg("Found an account")
				// Replace body. Replaces all occurances of the requested username with the logged in's username
				newBody := strings.ReplaceAll(string(body), username.Username, acc.Name)
				log.Debug().Str("new-body", newBody).Msg("New body passed to auth begin endpoint")
				r.Body = io.NopCloser(strings.NewReader(newBody))
				r.ContentLength = int64(len(newBody))
				handler.ServeHTTP(w, r)
			case storage.ErrAccountNotFound:
				log.Debug().Msg("Account from token not found")
				other.HttpErr(w, ErrIdDataNotFound, "Account from token not found", http.StatusForbidden)
			default:
				other.HttpErr(w, ErrIdDbErr, "Failed to get account in db", http.StatusInternalServerError)
				log.Error().Err(err).Msg("Failed to get account from db")
			}
		}
	})
}
