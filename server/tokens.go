package server

import "net/http"

func GetAllTokens(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}
}
