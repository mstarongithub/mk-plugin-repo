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

func GetAllTokens(w http.ResponseWriter, r *http.Request) {
	type Token struct {
		Name      string    `json:"name"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	type OutData struct {
		Tokens []Token `json:"tokens"`
	}
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

	returnTokens := OutData{[]Token{}}
	for _, token := range tokens {
		returnTokens.Tokens = append(
			returnTokens.Tokens,
			Token{token.Name, token.Token, token.ExpiresAt},
		)
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
	type InData struct {
		Name           string    `json:"name"`
		ExpirationDate time.Time `json:"expiration_date"`
	}
	type OutData struct {
		Name           string    `json:"name"`
		Token          string    `json:"token"`
		ExpirationDate time.Time `json:"expiration_date"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}
	body, _ := io.ReadAll(r.Body)
	inData := InData{}
	err := json.Unmarshal(body, &inData)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "invalid body data", http.StatusBadRequest)
		return
	}
	if inData.Name == "" || inData.ExpirationDate.Before(time.Now().Add(time.Minute)) {
		other.HttpErr(w, ErrIdBadRequest, "invalid name or expiration date", http.StatusBadRequest)
		return
	}

	tokenToken, err := store.NewToken(*accId, inData.Name, inData.ExpirationDate)
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
	token := OutData{Name: inData.Name, ExpirationDate: inData.ExpirationDate, Token: tokenToken}
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
	type InData struct {
		ExtendTo time.Time `json:"extend_to"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}

	name := r.PathValue("tokenName")
	body, _ := io.ReadAll(r.Body)
	data := InData{}
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
	token, err := store.FindTokenByName(*accId, name)
	switch err {
	case nil:
	case storage.ErrDataNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "token doesn't exist", http.StatusNotFound)
		return
	default:
		log.Error().
			Err(err).
			Uint("account-id", *accId).
			Str("token-name", name).
			Msg("Db failure while getting token for update")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure while getting token to update",
			http.StatusInternalServerError,
		)
	}
	log.Info().
		Uint("token-id", token.ID).
		Time("new-expiry", data.ExtendTo).
		Msg("Extending token lifetime")
	token.ExpiresAt = data.ExtendTo
	store.ExtendToken(token)
}

func InvalidateToken(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}

	name := r.PathValue("name")
	if name == "" {
		other.HttpErr(w, ErrIdBadRequest, "name path value must be set", ErrIdBadRequest)
		return
	}
	log.Info().Str("token-name", name).Uint("account-id", *accId).Msg("Invalidating token")
	store.InvalidateTokenByName(name, *accId)
}
