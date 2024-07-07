package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"gitlab.com/mstarongitlab/goutils/sliceutils"

	"github.com/mstarongithub/mk-plugin-repo/storage"
	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// Data expected for making a new plugin via POST /api/v1/plugins
type NewPluginData struct {
	Name            string   `json:"name"`             // Name of the plugin
	SummaryShort    string   `json:"summary_short"`    // A short summary
	SummaryLong     string   `json:"summary_long"`     // A full description of the plugin
	InitialVersion  string   `json:"initial_version"`  // The version of this new plugin
	Code            string   `json:"code"`             // The code associated with this new plugin and version
	Tags            []string `json:"tags"`             // The tags this plugin falls under
	Type            string   `json:"type"`             // What type the plugin is. Valid values are "plugin" and "widget"
	AIScriptVersion string   `json:"aiscript_version"` // The AI Script version this plugin is intended for
}

// Data a request to read a Plugin returns (GET /api/v1/plugins -> Array of this, GET /api/v1/plugins/{Plugin-id} -> One instance)
type Plugin struct {
	ID             uint     `json:"id"`              // The unique ID of the plugin
	Name           string   `json:"name"`            // The name of the plugin
	SummaryShort   string   `json:"summary_short"`   // A short summary of the plugin
	SummaryLong    string   `json:"summary_long"`    // A full description of the plugin
	CurrentVersion string   `json:"current_version"` // The latest version uploaded
	AllVersions    []string `json:"all_versions"`    // All versions of this plugin that have been uploaded
	Tags           []string `json:"tags"`            // All tags this plugin falls under
	AuthorID       uint     `json:"author_id"`       // The ID of the author
	Type           string   `json:"type"`            // Type of the plugin. Valid values are "plugin" and "widget"
}

// Data returned from GET /api/v1/plugins
type PluginList struct {
	Plugins []Plugin `json:"plugins"` // A list of plugins
	Page    *int     `json:"page"`    // The current page you've received. Not set if only one page
	Pages   *int     `json:"pages"`   // Total number of pages. Not set if only one page
}

// Data expected for updating a plugin via PUT /api/v1/plugins/{plugin-id}
type UpdatePluginData struct {
	Name         *string   `json:"name,omitempty"`          // Name of the plugin
	SummaryShort *string   `json:"summary_short,omitempty"` // A short summary
	SummaryLong  *string   `json:"summary_long,omitempty"`  // A full description of the plugin
	Tags         *[]string `json:"tags,omitempty"`          // The tags this plugin falls under
	Type         *string   `json:"type,omitempty"`          // What type the plugin is. Valid values are "plugin" and "widget"
}

// GET /api/v1/plugins
// Get a list of plugins. May be non-exhaustive and uses paging
// Optional GET parameters:
// - name: search for plugins containing the value in their name
// - content: search for plugins containing the value in their description
// - page: which "page" to select of the list of plugins
// - tags: semicolon separated list of tags that must be included
// TODO: Change return value to paginated version using PluginList
func getPluginList(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	dbPlugins := store.GetAllPlugins()
	apiPlugins := sliceutils.Map(dbPlugins, func(p storage.Plugin) Plugin {
		return dbPluginToApiPlugin(&p)
	})
	logrus.WithFields(logrus.Fields{
		"db-plugins":  dbPlugins,
		"api-plugins": apiPlugins,
	}).Debugln("Found plugins with conversion")
	if len(dbPlugins) == 0 {
		return
	}

	r.Header.Add("Content-Type", "application/json")
	data, err := json.Marshal(apiPlugins)
	if err != nil {
		logrus.WithError(err).
			WithField("plugins", apiPlugins).
			Errorln("Failed to convert plugins to json")
		http.Error(w, "json conversion failed", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(data))
}

// POST /api/v1/plugins
// RESTRICTED
// Add a new plugin to the repo
// New plugins will only be available after approval from an admin
// Body must be a json version of NewPluginData
func addNewPlugin(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	actorId := AccIdFromRequestContext(w, r)
	if actorId == nil {
		return
	}

	body, _ := io.ReadAll(r.Body)

	newPlugin := NewPluginData{}
	err := json.Unmarshal(body, &newPlugin)
	if err != nil {
		logrus.WithError(err).
			WithField("body", string(body)).
			Errorln("Failed to parse json from body")
		http.Error(
			w,
			"body must be a json-encoded representation of NewPluginData",
			http.StatusBadRequest,
		)
		return
	}
	// And now parse the plugin type
	var pluginType customtypes.PluginType
	switch newPlugin.Type {
	case "plugin":
		pluginType = customtypes.PLUGIN_TYPE_PLUGIN
	case "widget":
		pluginType = customtypes.PLUGIN_TYPE_WIDGET
	}
	// Then try throwing it into the db
	logrus.WithFields(logrus.Fields{
		"plugin": newPlugin,
		"uid":    *actorId,
	}).Debugln("Attempting to add plugin to db")
	_, err = store.NewPlugin(
		newPlugin.Name,
		*actorId,
		newPlugin.InitialVersion,
		newPlugin.SummaryLong,
		newPlugin.SummaryShort,
		newPlugin.Tags,
		pluginType,
		newPlugin.Code,
		newPlugin.AIScriptVersion,
	)
	if err != nil {
		logrus.WithError(err).WithField("plugin", newPlugin).Errorln("Failed to add plugin to db")
		http.Error(
			w,
			fmt.Sprintf("failed to insert new plugin. Error: %s", err.Error()),
			http.StatusInternalServerError,
		)
	}
	w.WriteHeader(http.StatusCreated)
}

// GET /api/v1/plugins/{pluginId}
// Get a specific plugin, specified by {plugin-id}
func getSpecificPlugin(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}

	pluginID := r.PathValue("pluginId")
	if pluginID == "" {
		// TODO: Add stat collection
		// Not necessary to log this case
		http.Error(
			w,
			"missing plugin id. Endpoint usage: GET /api/v1/plugins/{plugin-id}",
			http.StatusBadRequest,
		)
		return
	}
	pID, err := strconv.ParseUint(pluginID, 10, 0)
	if err != nil {
		// TODO: Add stat collection
		// Not necessary to log this case
		http.Error(w, "bad plugin ID", http.StatusBadRequest)
		return
	}
	log.WithField("plugin-id", pID).Infoln("Requested public plugin data")

	storagePlugin, err := store.GetPluginByID(uint(pID))
	if err != nil {
		// TODO: Add stat collection
		if errors.Is(err, storage.ErrPluginNotFound) {
			// Not necessary to log this case
			http.Error(w, "plugin not found", http.StatusNotFound)
		} else {
			log.WithError(err).WithField("plugin-id", pID).Warningln("Failed to get plugin from storage layer")
			http.Error(w, "error getting plugin from storage layer", http.StatusInternalServerError)
		}
		return
	}
	apiPlugin := dbPluginToApiPlugin(storagePlugin)
	jbody, err := json.Marshal(&apiPlugin)
	if err != nil {
		log.WithError(err).WithField("plugin-id", pID).Warning("Failed to encode result to json")
		http.Error(w, "json encoding failed", http.StatusInternalServerError)
		return
	}
	w.Write(jbody)
}

// PUT /api/v1/plugins/{pluginId}
// RESTRICTED
// Update a specific plugin
func updateSpecificPlugin(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := LogFromRequestContext(w, r)
	if log == nil {
		return
	}
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}

	// Get and parse user id and plugin id

	pluginString := r.PathValue("pluginId")
	pluginID, err := strconv.ParseUint(pluginString, 10, 0)
	if err != nil {
		http.Error(w, "bad plugin id. Must be a uint", http.StatusBadRequest)
		return
	}

	// Try getting plugin from db
	// TODO: Add logging: What plugin to get
	plugin, err := store.GetPluginByID(uint(pluginID))
	if err != nil {
		if errors.Is(err, storage.ErrPluginNotFound) {
			http.Error(w, "plugin not found", http.StatusNotFound)
		} else {
			log.WithError(err).WithField("plugin-id", pluginID).Warningln("Failed to get plugin from storage layer")
			http.Error(w, "problem getting plugin from storage layer", http.StatusInternalServerError)
		}
		return
	}
	// Check if the user authenticated is actually allowed to edit this plugin (aka is the owner)
	if plugin.AuthorID != *accId {
		http.Error(w, "you're not the owner of the plugin", http.StatusUnauthorized)
		return
	}

	body, _ := io.ReadAll(r.Body) // FIXME: Properly handle error
	updateData := UpdatePluginData{}
	err = json.Unmarshal(body, &updateData)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	if updateData.Name != nil {
		plugin.Name = *updateData.Name
	}
	if updateData.Tags != nil {
		plugin.Tags = *updateData.Tags
	}
	if updateData.Type != nil {
		switch *updateData.Type {
		case PLUGIN_TYPE_PLUGIN:
			plugin.Type = customtypes.PLUGIN_TYPE_PLUGIN
		case PLUGIN_TYPE_WIDGET:
			plugin.Type = customtypes.PLUGIN_TYPE_WIDGET
		}
	}
	if updateData.SummaryShort != nil {
		plugin.SummaryShort = *updateData.SummaryShort
	}
	if updateData.SummaryLong != nil {
		plugin.SummaryLong = *updateData.SummaryLong
	}

	_ = store.UpdatePlugin(plugin)
}

// DELETE /api/v1/plugins/{pluginId}
// RESTRICTED
// Delete a specific plugin
// Note: Won't actually delete, but marked to no longer be displayed
func deleteSpecificPlugin(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}

	// Get and parse user id and plugin id
	pluginString := r.PathValue("pluginId")
	pluginID, err := strconv.ParseUint(pluginString, 10, 0)
	if err != nil {
		http.Error(w, "bad plugin id. Must be a uint", http.StatusBadRequest)
		return
	}

	// TODO: Add logging: About to attempt plugin deletion with plugin and user id
	err = store.DeletePlugin(uint(pluginID), *accId)
	if err != nil {
		// TODO: Add logging
		http.Error(w, "couldn't delete plugin", http.StatusInternalServerError)
	}
}
