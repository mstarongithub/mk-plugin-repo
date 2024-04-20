package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type VersionData struct {
	Code                    string `json:"code"`
	IntendedAiScriptVersion string `json:"aiscript_version"`
}

// GET /api/v1/plugins/{pluginId}/{versionName}
// Get the details for a specific version
// Returns a json formatted VersionData on success
func getVersion(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(r)
	if store == nil {
		logrus.Errorln("getVersion: Failed to get storage from request context")
		http.Error(
			w,
			"failed to get storage layer from request context",
			http.StatusInternalServerError,
		)
		return
	}
	pluginIDString := r.PathValue("pluginId")
	versionName := r.PathValue("versionName")
	if pluginIDString == "" || versionName == "" {
		logrus.WithFields(logrus.Fields{
			"pluginId":    pluginIDString,
			"versionName": versionName,
		}).Infoln("Bad path request parameters")
		http.Error(w, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"pluginId":    pluginIDString,
			"versionName": versionName,
		}).Infoln("Plugin ID is parsable as uint")
	}
	version, err := store.TryFindVersion(uint(pluginID), versionName)
	if err != nil {
		if errors.Is(err, storage.ErrVersionNotFound) {
			logrus.WithFields(logrus.Fields{
				"pluginId":    pluginID,
				"versionName": versionName,
			}).Infoln("Plugin version not found")
		} else {
			logrus.WithError(err).WithFields(logrus.Fields{
				"pluginId":    pluginID,
				"versionName": versionName,
			}).Error("Problem getting version for plugin")
		}
	}
	logrus.WithFields(logrus.Fields{
		"pluginId":    pluginID,
		"versionName": versionName,
		"version":     version,
	}).Debugln("Found plugin version")
	binaryData, err := json.Marshal(&VersionData{
		Code:                    version.Code,
		IntendedAiScriptVersion: version.AiScriptVersion,
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"pluginId":    pluginID,
			"versionName": versionName,
			"version":     version,
		}).Errorln("Failed to marshal version")
		http.Error(w, "json marshalling failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, binaryData)
}

// POST /api/v1/plugins/{pluginId}/{versionName}
// RESTRICTED
// Create a new version
// Expects json formatted VersionData
// {version-name} will be the name of the new version
// Returns 4xx (whatever the bad request status is) if the version already exists
// TODO: Implement me!
func newVersion(w http.ResponseWriter, r *http.Request) {}

// DELETE /api/v1/plugins/{pluginId}/{versionName}
// RESTRICTED
// Hide a version. Doesn't delete, just hides it from the API
// TODO: Implement me!
func hideVersion(w http.ResponseWriter, r *http.Request) {}
