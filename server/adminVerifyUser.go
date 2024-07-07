package server

import (
	"encoding/json"
	"io"
	"net/http"
)

type VerifyUserData struct {
	Id uint `json:"id"`
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
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}
	actor, err := store.FindAccountByID(*actorId)
	if err != nil {
		log.WithError(err).
			WithField("actor-id", *actorId).
			Warning("Failed to get actor from ID after verification")
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	if !actor.Approved && !actor.CanApproveUsers {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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
	data := VerifyUserData{}
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
