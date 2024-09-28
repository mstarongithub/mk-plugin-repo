package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/hlog"
	"gitlab.com/mstarongitlab/goutils/other"
	"gitlab.com/mstarongitlab/goutils/sliceutils"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type VerifyThing struct {
	Id uint `json:"id"`
}

type IdList struct {
	Ids []uint `json:"ids"`
}

type AccountData struct {
	Name         string   `json:"name"`
	Mail         *string  `json:"mail"`
	Description  string   `json:"description"`
	Approved     bool     `json:"approved"`
	UserAdmin    bool     `json:"user_admin"`
	PluginAdmin  bool     `json:"plugin_admin"`
	PluginsOwned []uint   `json:"plugins_owned"`
	Links        []string `json:"links"`
}

func VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "invalid request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not json data", http.StatusBadRequest)
		return
	}
	target, err := store.FindAccountByID(data.Id)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "account not found", http.StatusNotFound)
	default:
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while looking for account",
			http.StatusInternalServerError,
		)
	}
	if target.Approved {
		return
	}
	target.Approved = true
	err = store.UpdateAccount(target)
	if err != nil {
		log.Warn().Err(err).
			Uint("target-id", target.ID).
			Msg("Failed to approve user in db")
		other.HttpErr(w, ErrIdDbErr, "failed to update in db", http.StatusInternalServerError)
		return
	}
}

func VerifyNewPluginHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "invalid request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not json data", http.StatusBadRequest)
		return
	}
	plugin, err := store.GetPluginByID(data.Id)
	switch err {
	case nil:
	case storage.ErrPluginNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "plugin not found", http.StatusNotFound)
		return
	default:
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db problem while looking for plugin",
			http.StatusInternalServerError,
		)
		return
	}
	plugin.Approved = true
	err = store.UpdatePlugin(plugin)
	if err != nil {
		log.Warn().Err(err).Uint("plugin-id", plugin.ID).Msg("Failed to approve plugin")
		other.HttpErr(w, ErrIdDbErr, "failed to update db", http.StatusInternalServerError)
	}
}

func GetAllUnverifiedPluginshandler(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	plugins, err := store.GetUnapprovedPlugins()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get list of unapproved plugins from db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while looking for plugins",
			http.StatusInternalServerError,
		)
		return
	}
	ids := sliceutils.Map(plugins, func(p storage.Plugin) uint {
		return p.ID
	})
	data, err := json.Marshal(&ids)
	if err != nil {
		log.Warn().Err(err).Uints("ids", ids).Msg("Failed to marshal ids to json")
		other.HttpErr(w, ErrIdJsonMarshal, "json marshal fail", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

func GetAllUnverifiedAccountsHandler(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	plugins, err := store.GetAllUnapprovedAccounts()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get list of unapproved accounts from db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while getting unapproved accounts",
			http.StatusInternalServerError,
		)
		return
	}
	ids := sliceutils.Map(plugins, func(p storage.Account) uint {
		return p.ID
	})
	data, err := json.Marshal(&ids)
	if err != nil {
		log.Warn().Err(err).Uints("ids", ids).Msg("Failed to marshal ids to json")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"json marshalling failed",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(data))
}

func PromotePluginAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "target account doesn't exist", http.StatusNotFound)
		return
	default:
		log.Error().Err(err).Msg("Failed to get account to promote to plugin admin")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while getting target account",
			http.StatusInternalServerError,
		)
		return
	}
	if !acc.Approved {
		other.HttpErr(w, ErrIdNotApproved, "target account not approved", http.StatusBadRequest)
		return
	}
	acc.CanApprovePlugins = true
	store.UpdateAccount(acc)
}

func PromoteAccountAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not required json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "Target account not found", http.StatusNotFound)
		return
	default:
		log.Error().Err(err).Msg("Db failure while getting account for promotion")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while getting account for promotion",
			http.StatusInternalServerError,
		)
		return
	}
	if !acc.Approved {
		other.HttpErr(w, ErrIdNotApproved, "target account not approved", http.StatusBadRequest)
		return
	}
	acc.CanApproveUsers = true
	store.UpdateAccount(acc)
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
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
	data := VerifyThing{}
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

func DemotePluginAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "body not expected json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "account to demote not found", http.StatusNotFound)
		return
	default:
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db error while getting account to demote",
			http.StatusInternalServerError,
		)
		log.Error().Err(err).Uint("account-id", data.Id).Msg("Db failure while getting an account")
		return
	}
	if acc.ID == 1 {
		log.Warn().Uint("actor", *actorId).
			Msg("Account admin tried to demote the superuser! Refusing attempt")
		other.HttpErr(
			w,
			ErrIdNotApproved,
			"How dare try to demote superuser!",
			http.StatusForbidden,
		)
		return
	}
	acc.CanApprovePlugins = false
	store.UpdateAccount(acc)
}

func DemoteAccountAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "bad request body", http.StatusBadRequest)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "invalid json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "target to demote not found", http.StatusNotFound)
		return
	default:
		log.Error().
			Err(err).
			Uint("target-id", data.Id).
			Msg("DB failure while getting account to demote")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure while getting account to demote",
			http.StatusInternalServerError,
		)
		return
	}
	if acc.ID == 1 {
		log.Warn().Uint("actor", *actorId).
			Msg("Account admin tried to demote the superuser! Refusing attempt")
		other.HttpErr(
			w,
			ErrIdNotApproved,
			"How dare trying to demote the superuser",
			http.StatusForbidden,
		)
		return
	}
	acc.CanApproveUsers = false
	store.UpdateAccount(acc)
}

func InspectAccountAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	idString := r.PathValue("id")
	id, err := strconv.ParseUint(idString, 0, 0)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "id must be uint number", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(uint(id))
	switch err {
	case nil:
	case storage.ErrAccountNotFound:
		other.HttpErr(w, ErrIdDataNotFound, "Target account doesn't exist", http.StatusBadRequest)
		return
	default:
		log.Error().Err(err).Uint64("target-id", id).Msg("Db failure searching for account")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"db failure searching for account",
			http.StatusInternalServerError,
		)
		return
	}
	retStruct := AccountData{
		Name:         acc.Name,
		Description:  acc.Description,
		Mail:         acc.Mail,
		Approved:     acc.Approved,
		UserAdmin:    acc.CanApproveUsers,
		PluginAdmin:  acc.CanApprovePlugins,
		PluginsOwned: acc.PluginsOwned,
		Links:        acc.Links,
	}
	retData, err := json.Marshal(&retStruct)
	if err != nil {
		log.Error().Err(err).
			Uint64("target-account", id).
			Msg("Failed to marshal return json!")
		other.HttpErr(w, ErrIdJsonMarshal, "Failed to marshal json", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(retData))
}
