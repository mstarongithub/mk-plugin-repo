package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

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
	fmt.Fprint(w, string(binaryData))
}

// POST /api/v1/plugins/{pluginId}
// RESTRICTED
// Create a new version
// Expects json formatted NewVersion
// Returns 4xx (whatever the bad request status is) if the version already exists
func newVersion(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(r)
	if store == nil {
		logrus.Errorln("newVersion: Failed to get storage from request context")
		http.Error(
			w,
			"failed to get storage layer from request context",
			http.StatusInternalServerError,
		)
		return
	}
	// ab := AuthbossFromRequest(r)
	// if ab == nil {
	// 	logrus.Errorln("newVersion: Failed to get authboss instance from request context")
	// 	http.Error(
	// 		w,
	// 		"failed to get auth layer from request context",
	// 		http.StatusInternalServerError,
	// 	)
	// 	return
	// }
	pluginIDString := r.PathValue("pluginId")
	if pluginIDString == "" {
		logrus.WithFields(logrus.Fields{
			"pluginId": pluginIDString,
		}).Infoln("Bad path request parameters")
		http.Error(w, "bad path parameters", http.StatusBadRequest)
		return
	}
	pluginID, err := strconv.ParseUint(pluginIDString, 10, 0)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"pluginId": pluginIDString,
		}).Infoln("Plugin ID is parsable as uint")
	}

	// Ignore error. Should never fail I think
	body, _ := io.ReadAll(r.Body)
	newVersion := NewVersion{}
	err = json.Unmarshal(body, &newVersion)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"body":           body,
			"body-as-string": string(body),
			"pluginId":       pluginID,
		}).Debugln("Failed to extract new version from body")
		http.Error(w, "body is not a json-encoded NewVersion", http.StatusBadRequest)
	}

	err = store.NewVersion(
		uint(pluginID),
		newVersion.VersionName,
		newVersion.Code,
		newVersion.IntendedAiScriptVersion,
	)
	if err != nil {
		if !errors.Is(err, storage.ErrVersionAlreadyExists) {
			logrus.WithError(err).WithFields(logrus.Fields{
				"new-version": newVersion,
				"pluginId":    pluginID,
			}).Errorln("failed to create new version")
			http.Error(w, "version creation failed", http.StatusInternalServerError)
		} else {
			logrus.WithFields(logrus.Fields{
				"new-version": newVersion,
				"pluginId":    pluginID,
			}).Debugln("version with that name already exists, ignoring")
			http.Error(w, "version already exists", http.StatusNotAcceptable)
		}
		return
	}
}

// DELETE /api/v1/plugins/{pluginId}/{versionName}
// RESTRICTED
// Hide a version. Doesn't delete, just hides it from the API
func hideVersion(w http.ResponseWriter, r *http.Request) {
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
	if err = store.HideVersion(uint(pluginID), versionName); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"pluginID":    pluginID,
			"versionName": versionName,
		}).Errorln("Error trying to \"delete\" a version")
		http.Error(w, "problem trying to delete version", http.StatusInternalServerError)
		return
	}
}
