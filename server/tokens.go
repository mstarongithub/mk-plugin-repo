package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/rs/zerolog/hlog"
	"gitlab.com/mstarongitlab/goutils/other"
)

type returnTokenData struct {
  Tokens map[string]string
}

type newTokenData struct {
	Token string
}

type extendTokenData struct {
	ExtendTo  time.Time
	TokenName string
}

func GetAllTokens(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}
	log := hlog.FromRequest(r)

	tokens, err := store.GetTokensForAccountID(*accId)
	if err != nil {
		log.Error().
			Err(err).
			Uint("account-id", *accId).
			Msg("db failure while getting tokens for an account")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure while getting tokens",
			http.StatusInternalServerError,
		)
		return
	}

	returnTokens := returnTokenData{map[string]string{}}
	for _, token := range tokens {
		returnTokens.Tokens[token.Name] = token.Token
	}

	outBytes, err := json.Marshal(&returnTokens)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to marshal tokens")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"failed to marshal response",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(outBytes))
}

func GenerateNewToken(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse form data")
		other.HttpErr(w, ErrIdBadRequest, "bad form data", http.StatusBadRequest)
		return
	}
	tokenName := r.FormValue("name")
	if tokenName == "" {
		other.HttpErr(
			w,
			ErrIdBadRequest,
			"no token name (form name \"name\") provided",
			http.StatusBadRequest,
		)
		return
	}
	tokenToken, err := store.NewToken(*accId, tokenName, time.Now().Add(time.Hour*24*30*12))
	if err != nil {
		log.Error().Err(err).Msg("DB failure while creating new token")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure while creating new token",
			http.StatusInternalServerError,
		)
		return
	}
	token := newTokenData{tokenToken}
	data, err := json.Marshal(&token)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to parse return json for new token")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"failed to marshal response data",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(data))
}

func ExtendToken(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}
	_ = log

	body, _ := io.ReadAll(r.Body)
	data := extendTokenData{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad json data", http.StatusBadRequest)
		return
	}
	if data.ExtendTo.Before(time.Now()) {
		other.HttpErr(
			w,
			ErrIdCantExtendIntoPast,
			"can't extend a token into the past",
			http.StatusBadRequest,
		)
		return
	}
	token, err := store.FindTokenByName(*accId, data.TokenName)
	switch err {
	case nil:
	case storage.ErrDataNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "token doesn't exist", http.StatusNotFound)
		return
	default:
		log.Error().
			Err(err).
			Uint("account-id", *accId).
			Str("token-name", data.TokenName).
			Msg("Db failure while getting token for update")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure while getting token to update",
			http.StatusInternalServerError,
		)
	}
	token.ExpiresAt = data.ExtendTo
	store.ExtendToken(token)
}

func 
