package api

import (
	"net/http"

	"golang.org/x/crypto/argon2"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

func BuildPublicApi() (http.Handler, error) {
	router := http.NewServeMux()

	return router, nil
}

func handleLoginRequest(w http.ResponseWriter, r *http.Request) {
	storage, ok := r.Context().Value(storage.STORAGE_CONTEXT_KEY).(*storage.Storage)
	if !ok {
		http.Error(w, "failed to get a storage reference", http.StatusInternalServerError)
		return
	}

}
