package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/other"

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
	type OutData struct {
		Code                    string `json:"code"`
		IntendedAiScriptVersion string `json:"aiscript_version"`
	}
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	pluginIDString := r.PathValue("pluginId")
	versionName := r.PathValue("versionName")
	logger := log.With().Str("plugin-id", pluginIDString).Str("version-name", versionName).Logger()
	if pluginIDString == "" || versionName == "" {
		logger.Info().Msg("Bad path request parameters")
		other.HttpErr(w, ErrIdBadRequest, "Bad path request parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "Plugin ID is not a uint", ErrIdBadRequest)
	}
	version, err := store.TryFindVersion(uint(pluginID), versionName)
	if err != nil {
		if errors.Is(err, storage.ErrVersionNotFound) {
			logger.Info().Msg("Plugin version not found")
			other.HttpErr(w, ErrIdDataNotFound, "Version not found", http.StatusNotFound)
		} else {
			logger.Error().Err(err).Msg("Problem getting version for plugin")
			other.HttpErr(w, ErrIdDbErr, "Failed to get version from db", http.StatusInternalServerError)
		}
		return
	}
	logger.Info().Msg("Found version, marshalling and sending off")
	binaryData, err := json.Marshal(&OutData{
		Code:                    version.Code,
		IntendedAiScriptVersion: version.AiScriptVersion,
	})
	if err != nil {
		logger.Error().Err(err).Msg("Failed to marshal version")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"Failed to marshal response",
			http.StatusInternalServerError,
		)
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
	type InData struct {
		Code                    string `json:"code"`
		IntendedAiScriptVersion string `json:"aiscript_version"`
		VersionName             string `json:"version_name"`
	}
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
		other.HttpErr(w, ErrIdBadRequest, "Plugin ID is not a uint", http.StatusBadRequest)
		return
	}

	// Ignore error. Should never fail I think
	body, _ := io.ReadAll(r.Body)
	newVersion := InData{}
	err = json.Unmarshal(body, &newVersion)
	if err != nil {
		logger.Warn().Bytes("body", body).Msg("Failed to unmarshal body to Version")
		other.HttpErr(w, ErrIdBadRequest, "Body is not valid json", http.StatusBadRequest)
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
			other.HttpErr(
				w,
				ErrIdDbErr,
				"Failed to insert new version into db",
				http.StatusInternalServerError,
			)
		} else {
			logger.Warn().Msg("Version already exists")
			other.HttpErr(w, ErrIdAlreadyExists, "Version of that name already exists", http.StatusNotAcceptable)
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
		other.HttpErr(w, ErrIdBadRequest, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logger.Warn().Err(err).Msg("Plugin ID is not a uint")
		other.HttpErr(w, ErrIdBadRequest, "Plugin ID is not a uint", http.StatusBadRequest)
		return
	}
	if err = store.DeleteVersion(uint(pluginID), versionName); err != nil {
		logger.Error().
			Err(err).
			Uint("plugin-id", uint(pluginID)).
			Str("plugin-version", versionName).
			Msg("Couldn't delete version")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Db failure while deleting version",
			http.StatusInternalServerError,
		)
		return
	}
}
