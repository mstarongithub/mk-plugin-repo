package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type VersionData struct {
	Code                    string `json:"code"`
	IntendedAiScriptVersion string `json:"aiscript_version"`
}

type NewVersion struct {
	Code                    string `json:"code"`
	IntendedAiScriptVersion string `json:"aiscript_version"`
	VersionName             string `json:"version_name"`
}

// GET /api/v1/plugins/{pluginId}/{versionName}
// Get the details for a specific version
// Returns a json formatted VersionData on success
func getVersion(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	pluginIDString := r.PathValue("pluginId")
	versionName := r.PathValue("versionName")
	logger := log.With().Str("plugin-id", pluginIDString).Str("version-name", versionName).Logger()
	if pluginIDString == "" || versionName == "" {
		logger.Info().Msg("Bad path request parameters")
		http.Error(w, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logger.Info().Err(err).Msg("Plugin ID is not a uint")
	}
	version, err := store.TryFindVersion(uint(pluginID), versionName)
	if err != nil {
		if errors.Is(err, storage.ErrVersionNotFound) {
			logger.Info().Msg("Plugin version not found")
		} else {
			logger.Error().Err(err).Msg("Problem getting version for plugin")
		}
	}
	logger.Info().Msg("Found version, marshalling and sending off")
	binaryData, err := json.Marshal(&VersionData{
		Code:                    version.Code,
		IntendedAiScriptVersion: version.AiScriptVersion,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal version")
		http.Error(w, "json marshalling failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(binaryData))
}

// POST /api/v1/plugins/{pluginId}
// RESTRICTED
// Create a new version
// Expects json formatted NewVersion
// Returns 4xx (whatever the bad request status is) if the version already exists
func newVersion(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	pluginIDString := r.PathValue("pluginId")
	logger := log.With().Str("plugin-id", pluginIDString).Logger()
	if pluginIDString == "" {
		logger.Warn().Msg("Bad path request parameters")
		http.Error(w, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logger.Warn().Err(err).Msg("Plugin Id is not a uint")
		http.Error(w, "plugin id not a uint", http.StatusBadRequest)
		return
	}

	// Ignore error. Should never fail I think
	body, _ := io.ReadAll(r.Body)
	newVersion := NewVersion{}
	err = json.Unmarshal(body, &newVersion)
	if err != nil {
		logger.Warn().Bytes("body", body).Msg("Failed to unmarshal body to Version")
		http.Error(w, "body is not a json-encoded NewVersion", http.StatusBadRequest)
		return
	}

	logger.Debug().Msg("Inserting new plugin version")
	err = store.NewVersion(
		uint(pluginID),
		newVersion.VersionName,
		newVersion.Code,
		newVersion.IntendedAiScriptVersion,
	)
	if err != nil {
		if !errors.Is(err, storage.ErrVersionAlreadyExists) {
			logger.Error().Err(err).Msg("Failed to create new version")
			http.Error(w, "version creation failed", http.StatusInternalServerError)
		} else {
			logger.Warn().Msg("Version already exists")
			http.Error(w, "version already exists", http.StatusNotAcceptable)
		}
		return
	}
}

// DELETE /api/v1/plugins/{pluginId}/{versionName}
// RESTRICTED
// Hide a version. Doesn't delete, just hides it from the API
func hideVersion(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}

	pluginIDString := r.PathValue("pluginId")
	versionName := r.PathValue("versionName")
	logger := log.With().Str("plugin-id", pluginIDString).Str("version-name", versionName).Logger()
	if pluginIDString == "" || versionName == "" {
		logger.Info().Msg("Bad path parameters")
		http.Error(w, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logger.Warn().Err(err).Msg("Plugin ID is not a uint")
		http.Error(w, "plugin ID not a uint", http.StatusBadRequest)
		return
	}
	if err = store.DeleteVersion(uint(pluginID), versionName); err != nil {
		logger.Error().Err(err).Msg("Couldn't delete version")
		http.Error(w, "problem trying to delete version", http.StatusInternalServerError)
		return
	}
}
