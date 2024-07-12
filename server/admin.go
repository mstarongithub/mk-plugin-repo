package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/mstarongitlab/goutils/sliceutils"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type VerifyThing struct {
	Id uint `json:"id"`
}

type IdList struct {
	Ids []uint `json:"ids"`
}

func VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}
	target, err := store.FindAccountByID(data.Id)
	if err != nil {
		http.Error(w, "bad account id", http.StatusBadRequest)
		return
	}
	if target.Approved {
		return
	}
	target.Approved = true
	err = store.UpdateAccount(target)
	if err != nil {
		log.WithError(err).
			WithField("target-id", target.ID).
			Warningln("Failed to approve user in db")
		http.Error(w, "failed to update account in db", http.StatusInternalServerError)
		return
	}
}

func VerifyNewPluginHandler(w http.ResponseWriter, r *http.Request) {
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}
	plugin, err := store.GetPluginByID(data.Id)
	if err != nil {
		http.Error(w, "plugin not found", http.StatusBadRequest)
		return
	}
	plugin.Approved = true
	err = store.UpdatePlugin(plugin)
	if err != nil {
		log.WithError(err).WithField("plugin-id", plugin.ID).Warningln("Failed to approve plugin")
		http.Error(w, "failed to update", http.StatusInternalServerError)
	}
}

func GetAllUnverifiedPluginshandler(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	plugins, err := store.GetUnapprovedPlugins()
	if err != nil {
		log.WithError(err).Warningln("Failed to get list of unapproved plugins from db")
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	ids := sliceutils.Map(plugins, func(p storage.Plugin) uint {
		return p.ID
	})
	data, err := json.Marshal(&ids)
	if err != nil {
		log.WithError(err).WithField("ids", ids).Warningln("Failed to marshal ids to json")
		http.Error(w, "failed to marshal ids to json", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

func GetAllUnverifiedAccountsHandler(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	plugins, err := store.GetAllUnapprovedAccounts()
	if err != nil {
		log.WithError(err).Warningln("Failed to get list of unapproved accounts from db")
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	ids := sliceutils.Map(plugins, func(p storage.Account) uint {
		return p.ID
	})
	data, err := json.Marshal(&ids)
	if err != nil {
		log.WithError(err).WithField("ids", ids).Warningln("Failed to marshal ids to json")
		http.Error(w, "failed to marshal ids to json", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

func PromotePluginAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	if err != nil {
		log.WithError(err).Warningln("Failed to get account to promote to plugin admin")
		http.Error(w, "problem finding account", http.StatusBadRequest)
		return
	}
	if !acc.Approved {
		http.Error(w, "target account not Approved", http.StatusBadRequest)
		return
	}
	acc.CanApprovePlugins = true
	store.UpdateAccount(acc)
}

func PromoteAccountAdminHandler(w http.ResponseWriter, r *http.Request) {
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}
	acc, err := store.FindAccountByID(data.Id)
	if err != nil {
		log.WithError(err).Warningln("Failed to get account to promote to plugin admin")
		http.Error(w, "problem finding account", http.StatusBadRequest)
		return
	}
	if !acc.Approved {
		http.Error(w, "target account not Approved", http.StatusBadRequest)
		return
	}
	acc.CanApproveUsers = true
	store.UpdateAccount(acc)
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	actor, err := store.FindAccountByID(*actorId)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
		return
	}
	data := VerifyThing{}
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		http.Error(w, "body not json data", http.StatusBadRequest)
		return
	}
	if data.Id != *actorId && !actor.CanApproveUsers {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	store.DeleteAccount(data.Id)
}
