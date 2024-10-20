package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/mstarongithub/mk-plugin-repo/storage"
	"github.com/rs/zerolog/hlog"
	"gitlab.com/mstarongitlab/goutils/other"
)

func getPublicAccountDataHandler(w http.ResponseWriter, r *http.Request) {
	type OutData struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Approved     bool     `json:"approved"`
		UserAdmin    bool     `json:"user_admin"`
		PluginAdmin  bool     `json:"plugin_admin"`
		PluginsOwned []uint   `json:"plugins_owned"`
		Links        []string `json:"links"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accountIdString := r.PathValue("accountId")
	if accountIdString == "" {
		other.HttpErr(w, ErrIdBadRequest, "Missing account id", http.StatusBadRequest)
		return
	}
	accId64, err := strconv.ParseUint(accountIdString, 10, 0)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "Account id must be a uint", http.StatusBadRequest)
		return
	}
	accId := uint(accId64)
	log.Debug().Uint("account-id", accId).Msg("Public access to account")
	acc, err := store.FindAccountByID(accId)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		log.Info().Uint("account-id", accId).Msg("Account not found")
		other.HttpErr(w, ErrIdDataNotFound, "Account not found", http.StatusNotFound)
		return
	default:
		log.Error().Uint("account-id", accId).Err(err).Msg("Failed to get account from storage")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Failed to get account from storage",
			http.StatusInternalServerError,
		)
		return
	}
	out := OutData{
		Name:         acc.Name,
		Description:  acc.Description,
		Approved:     acc.Approved,
		UserAdmin:    acc.CanApproveUsers,
		PluginAdmin:  acc.CanApprovePlugins,
		PluginsOwned: acc.PluginsOwned,
		Links:        acc.Links,
	}
	outJson, err := json.Marshal(&out)
	if err != nil {
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"Failed to marshal response",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(outJson))
}

func updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	type InData struct {
		AccountId   *uint    `json:"account_id"`
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Links       []string `json:"links"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	body, _ := io.ReadAll(r.Body)
	inData := InData{}
	err := json.Unmarshal(body, &inData)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "Body content must be proper json", http.StatusBadRequest)
		return
	}

	// If inData.AccountId != nil then check if actor has account admin perms
	var targetAccId uint = 0
	if inData.AccountId == nil {
		targetAccId = *actorId
	} else {
		targetAccId = *inData.AccountId

		actorAcc, err := store.FindAccountByID(*actorId)
		switch err {
		case nil:
		case storage.ErrAccountNotFound:
			other.HttpErr(w, ErrIdDataNotFound, "Couldn't find acting account", http.StatusNotFound)
			return
		default:
			other.HttpErr(w, ErrIdDbErr, "Failed to get acting account from db", http.StatusInternalServerError)
			log.Error().Err(err).Uint("account-id", *actorId).Msg("Failed to get account from db")
			return
		}
		if !actorAcc.Approved || !actorAcc.CanApproveUsers {
			other.HttpErr(w, ErrIdNotApproved, "Acting account can't modify target account's data", http.StatusForbidden)
			return
		}
	}

	acc, err := store.FindAccountByID(targetAccId)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "Account not found", http.StatusNotFound)
		return
	default:
		log.Error().Err(err).Uint("account-id", targetAccId).Msg("Failed to get account from db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Failed to get target account from db",
			http.StatusInternalServerError,
		)
		return
	}
	if inData.Name != nil {
		_, err = store.FindAccountByName(*inData.Name)
		switch err {
		case nil:
			other.HttpErr(
				w,
				ErrIdAlreadyExists,
				"Account with that name already exists",
				ErrIdBadRequest,
			)
			return
		case storage.ErrAccountNotFound: // Only acceptable case: No account with that name exists
		default:
			log.Error().
				Err(err).
				Str("account-name", *inData.Name).
				Msg("Failed to look for account in db")
			other.HttpErr(
				w,
				ErrIdDbErr,
				"Failed to look for account in db",
				http.StatusInternalServerError,
			)
			return
		}
		acc.Name = *inData.Name
	}
	if inData.Description != nil {
		acc.Description = *inData.Description
	}
	if len(inData.Links) > 0 {
		acc.Links = inData.Links
	}
	err = store.UpdateAccount(acc)
	if err != nil {
		log.Error().Err(err).Uint("account-id", acc.ID).Msg("Failed to update account")
		other.HttpErr(w, ErrIdDbErr, "Failed to update account", http.StatusInternalServerError)
	}
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	type InData struct {
		Id uint `json:"id"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	log := hlog.FromRequest(r)
	actor, err := store.FindAccountByID(*actorId)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		log.Warn().Uint("actor", *actorId).Msg("Actor performing account deletion not found")
		other.HttpErr(
			w,
			ErrIdDataNotFound,
			"Actor performing deletion not found",
			http.StatusNotFound,
		)
		return
	default:
		log.Error().
			Err(err).
			Uint("actor-id", *actorId).
			Msg("Problem getting actor for account deletion from db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Db failure while getting actor performing deletion",
			http.StatusInternalServerError,
		)
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad request body", http.StatusBadRequest)
		return
	}
	data := InData{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not required json data", http.StatusBadRequest)
		return
	}
	if data.Id != *actorId && !actor.CanApproveUsers {
		other.HttpErr(w, ErrIdNotApproved, "operation forbidden", http.StatusForbidden)
		return
	}
	if data.Id == 1 {
		log.Warn().Msg("Attempt to delete superuser. Telling them to kindly fuck off")
		other.HttpErr(
			w,
			ErrIdNotApproved,
			"Kindly fuck off and stop trying to delete the superuser",
			http.StatusForbidden,
		)
		return
	}
	store.DeleteAccount(data.Id)
}
