package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/hlog"
	"gitlab.com/mstarongitlab/goutils/other"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

func forceCorrectPasskeyAuthFlowMiddleware(
	handler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := hlog.FromRequest(r)
		// Don't fuck with the request if not intended for starting to register or login
		if strings.HasSuffix(r.URL.Path, "loginFinish") {
			log.Debug().Msg("Request to finish login method, doing nothing")
			handler.ServeHTTP(w, r)
			return
		} else if strings.HasSuffix(r.URL.Path, "registerFinish") {
			handler.ServeHTTP(w, r)
			// Force unset session cookie here
			w.Header().Del("Set-Cookie")
			http.SetCookie(w, &http.Cookie{
				Name:    "sid",
				Value:   "",
				Path:    "",
				MaxAge:  0,
				Expires: time.UnixMilli(0),
			})
			return
		} else if strings.HasSuffix(r.URL.Path, "loginBegin") {
			fuckWithLoginRequest(w, r, handler)
		} else if strings.HasSuffix(r.URL.Path, "registerBegin") {
			fuckWithRegisterRequest(w, r, handler)
		}
	})
}

func fuckWithRegisterRequest(
	w http.ResponseWriter,
	r *http.Request,
	nextHandler http.Handler,
) {
	log := hlog.FromRequest(r)
	log.Debug().Msg("Messing with register start request")
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	cookie, cookieErr := r.Cookie("sid")
	var username struct {
		Username string `json:"username"`
	}
	body, _ := io.ReadAll(r.Body)
	log.Debug().Bytes("body", body).Msg("Body of auth begin request")
	err := json.Unmarshal(body, &username)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "Not a username json object", http.StatusBadRequest)
		return
	}
	if cookieErr == nil {
		// Already authenticated, overwrite username to logged in account's name
		// Get session from cookie
		log.Debug().Msg("Session token exists, force overwriting username of register request")
		session, ok := store.GetSession(cookie.Value)
		if !ok {
			log.Error().Str("session-id", cookie.Value).Msg("Passkey session missing")
			other.HttpErr(w, ErrIdDbErr, "Passkey session missing", http.StatusInternalServerError)
			return
		}
		acc, err := store.FindAccountByPasskeyId(session.UserID)
		// Assume account must exist if a session for it exists
		if err != nil {
			log.Error().Err(err).Msg("Failed to get account from passkey id from session")
			other.HttpErr(
				w,
				ErrIdDbErr,
				"Failed to get authenticated account",
				http.StatusInternalServerError,
			)
			return
		}
		// Replace whatever username may be given with username of logged in account
		newBody := strings.ReplaceAll(string(body), username.Username, acc.Name)
		// Assign to request
		r.Body = io.NopCloser(strings.NewReader(newBody))
		r.ContentLength = int64(len(newBody))
		// And pass on
		nextHandler.ServeHTTP(w, r)
	} else {
		// Not authenticated, ensure that no existing name is registered with
		_, err = store.FindAccountByName(username.Username)
		switch err {
		case nil:
			// No error while getting account means account exists, refuse access
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
			// Didn't find account with that name, give access
			log.Debug().
				Str("username", username.Username).
				Msg("No account with this username exists yet, passing through")
				// Copy original body since previous reader hit EOF
			r.Body = io.NopCloser(strings.NewReader(string(body)))
			r.ContentLength = int64(len(body))
			nextHandler.ServeHTTP(w, r)
		default:
			// Some other error, log it and return appropriate message
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
	}
}

func fuckWithLoginRequest(
	w http.ResponseWriter,
	r *http.Request,
	nextHandler http.Handler,
) {
	log := hlog.FromRequest(r)
	log.Debug().Msg("Messing with login start request")
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	cookie, cookieErr := r.Cookie("sid")
	var username struct {
		Username string `json:"username"`
	}
	// Force ignore cookie for now
	_ = cookieErr
	var err error = errors.New("placeholder")
	if err == nil {
		// Someone is logged in, overwrite username with logged in account's one
		body, _ := io.ReadAll(r.Body)
		log.Debug().Bytes("body", body).Msg("Body of auth begin request")
		err := json.Unmarshal(body, &username)
		if err != nil {
			other.HttpErr(w, ErrIdBadRequest, "Not a username json object", http.StatusBadRequest)
			return
		}
		session, ok := store.GetSession(cookie.Value)
		if !ok {
			log.Error().Str("session-id", cookie.Value).Msg("Passkey session missing")
			other.HttpErr(w, ErrIdDbErr, "Passkey session missing", http.StatusInternalServerError)
			return
		}
		acc, err := store.FindAccountByPasskeyId(session.UserID)
		// Assume account must exist if a session for it exists
		if err != nil {
			log.Error().Err(err).Msg("Failed to get account from passkey id from session")
			other.HttpErr(
				w,
				ErrIdDbErr,
				"Failed to get authenticated account",
				http.StatusInternalServerError,
			)
			return
		}
		// Replace whatever username may be given with username of logged in account
		newBody := strings.ReplaceAll(string(body), username.Username, acc.Name)
		// Assign to request
		r.Body = io.NopCloser(strings.NewReader(newBody))
		r.ContentLength = int64(len(newBody))
		// And pass on
		nextHandler.ServeHTTP(w, r)
	} else {
		// No one logged in, check if user exists to prevent creating a bugged account
		body, _ := io.ReadAll(r.Body)
		log.Debug().Bytes("body", body).Msg("Body of auth begin request")
		err := json.Unmarshal(body, &username)
		if err != nil {
			other.HttpErr(w, ErrIdBadRequest, "Not a username json object", http.StatusBadRequest)
			return
		}
		_, err = store.FindAccountByName(username.Username)
		switch err {
		case nil:
		// All good, account exists, keep going
		// Do nothing in this branch
		case storage.ErrAccountNotFound:
			// Account doesn't exist, catch it
			other.HttpErr(w, ErrIdDataNotFound, "Username not found", http.StatusNotFound)
			return
		default:
			// catch db failures
			log.Error().Err(err).Str("username", username.Username).Msg("Db failure while getting account")
			other.HttpErr(w, ErrIdDbErr, "Failed to check for account in db", http.StatusInternalServerError)
			return
		}
		// Restore body as new reader of the same content
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		nextHandler.ServeHTTP(w, r)
	}
}
