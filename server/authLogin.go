package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mstarongithub/mk-plugin-repo/auth"
)

type AuthStateReturn struct {
	State            int    `json:"state"`
	ProcessIDorToken string `json:"process_id_or_token"`
}

type AuthMfaKey struct {
	Value     string
	ProcessId string
	Type      int
}

// Mounts at /api/v1/auth/password-start
func AuthLoginPWHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	username, password, basicAuthUsed := r.BasicAuth()
	if !basicAuthUsed {
		http.Error(w, "no auth set", http.StatusBadRequest)
		return
	}
	next, processID := authLayer.LoginWithPassword(username, password)
	returnState := AuthStateReturn{
		State:            int(next),
		ProcessIDorToken: processID,
	}
	data, err := json.Marshal(&returnState)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

// Mounts at /api/v1/auth/mfa-continue
func AuthLoginMfaHandler(w http.ResponseWriter, r *http.Request) {
	authLayer := AuthFromRequestContext(w, r)
	if authLayer == nil {
		return
	}
	rawData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "couldn't read body", http.StatusInternalServerError)
		return
	}
	authData := AuthMfaKey{}
	err = json.Unmarshal(rawData, &authData)
	if err != nil {
		http.Error(w, "body not json", http.StatusBadRequest)
		return
	}
	next, processId := authLayer.LoginWithMFA(
		authData.ProcessId,
		authData.Value,
		auth.NextAuthState(authData.Type),
	)
	nextState := AuthStateReturn{
		State:            int(next),
		ProcessIDorToken: processId,
	}
	returnData, err := json.Marshal(&nextState)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(returnData))
}
