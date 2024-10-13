package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/other"
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
	log := hlog.FromRequest(r)
	dbPlugins := store.GetAllPlugins()
	apiPlugins := sliceutils.Map(dbPlugins, func(p storage.Plugin) Plugin {
		return dbPluginToApiPlugin(&p)
	})

	log.Debug().
		Uints("db-plugins", sliceutils.Map(dbPlugins, func(t storage.Plugin) uint { return t.ID })).
		Msg("Found plugins")
	if len(dbPlugins) == 0 {
		return
	}

	r.Header.Add("Content-Type", "application/json")
	data, err := json.Marshal(apiPlugins)
	if err != nil {
		log.Error().Err(err).
			Uints("plugins", sliceutils.Map(dbPlugins, func(t storage.Plugin) uint { return t.ID })).
			Msg("Failed to convert plugins to json")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"Failed to marshal response",
			http.StatusInternalServerError,
		)
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
	log := hlog.FromRequest(r)

	body, _ := io.ReadAll(r.Body)

	newPlugin := NewPluginData{}
	err := json.Unmarshal(body, &newPlugin)
	if err != nil {
		log.Error().Err(err).
			Bytes("body", body).
			Msg("Failed to parse json from body")
		other.HttpErr(w, ErrIdBadRequest, "Body must be valid json data", http.StatusBadRequest)
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
	log.Debug().
		Any("plugin", newPlugin).
		Uint("actor", *actorId).
		Msg("Attempting to add plugin to db")
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
		log.Error().Err(err).Any("plugin", newPlugin).Msg("Failed to add plugin to db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Failed to insert new plugin into db",
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
	log := hlog.FromRequest(r)

	pluginID := r.PathValue("pluginId")
	if pluginID == "" {
		// TODO: Add stat collection
		// Not necessary to log this case
		other.HttpErr(w, ErrIdBadRequest, "Missing path parameter plugin-id", ErrIdBadRequest)
		return
	}
	pID, err := strconv.ParseUint(pluginID, 10, 0)
	if err != nil {
		// TODO: Add stat collection
		// Not necessary to log this case
		other.HttpErr(w, ErrIdBadRequest, "Plugin ID must be a uint", http.StatusBadRequest)
		return
	}
	log.Info().Uint64("plugin-id", pID).Msg("Requested public plugin data")

	storagePlugin, err := store.GetPluginByID(uint(pID))
	if err != nil {
		// TODO: Add stat collection
		if errors.Is(err, storage.ErrPluginNotFound) {
			// Not necessary to log this case
			other.HttpErr(w, ErrIdDataNotFound, "Plugin not found", http.StatusNotFound)
		} else {
			log.Error().Err(err).Uint64("plugin-id", pID).Msg("Failed to get plugin from storage layer")
			other.HttpErr(w, ErrIdDbErr, "Failed to get plugin from db", http.StatusInternalServerError)
		}
		return
	}
	apiPlugin := dbPluginToApiPlugin(storagePlugin)
	jbody, err := json.Marshal(&apiPlugin)
	if err != nil {
		log.Warn().Err(err).Uint64("plugin-id", pID).Msg("Failed to encode result to json")
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"Failed to marshal response data",
			http.StatusInternalServerError,
		)
		return
	}
	fmt.Fprint(w, string(jbody))
}

// PUT /api/v1/plugins/{pluginId}
// RESTRICTED
// Update a specific plugin
func updateSpecificPlugin(w http.ResponseWriter, r *http.Request) {
	store := StorageFromRequest(w, r)
	if store == nil {
		return
	}
	log := hlog.FromRequest(r)
	accId := AccIdFromRequestContext(w, r)
	if accId == nil {
		return
	}

	// Get and parse user id and plugin id

	pluginString := r.PathValue("pluginId")
	pluginID, err := strconv.ParseUint(pluginString, 10, 0)
	if err != nil {
		other.HttpErr(w, ErrIdBadRequest, "Plugin id must be a uint", http.StatusBadRequest)
		return
	}

	// Try getting plugin from db
	// TODO: Add logging: What plugin to get
	plugin, err := store.GetPluginByID(uint(pluginID))
	if err != nil {
		if errors.Is(err, storage.ErrPluginNotFound) {
			other.HttpErr(w, ErrIdDataNotFound, "Plugin not found", http.StatusNotFound)
		} else {
			log.Error().Err(err).Uint64("plugin-id", pluginID).Msg("Failed to get plugin from storage layer")
			other.HttpErr(w, ErrIdDbErr, "Failed to get plugin from db", http.StatusInternalServerError)
		}
		return
	}
	actor, err := store.FindAccountByID(*accId)
	if err != nil {
		log.Error().Err(err).Uint("actor-id", *accId).Msg("Failed to get actor")
		other.HttpErr(w, ErrIdDbErr, "Failed to get actor from db", http.StatusInternalServerError)
	}
	// Check if the user authenticated is actually allowed to edit this plugin (aka is the owner)
	if plugin.AuthorID != *accId && !actor.CanApprovePlugins {
		other.HttpErr(
			w,
			ErrIdNotApproved,
			"You're not the owner of the plugin",
			http.StatusUnauthorized,
		)
		return
	}

	body, _ := io.ReadAll(r.Body) // FIXME: Properly handle error
	updateData := UpdatePluginData{}
	err = json.Unmarshal(body, &updateData)
	if err != nil {
		other.HttpErr(
			w,
			ErrIdJsonMarshal,
			"Failed to marshal response data",
			http.StatusInternalServerError,
		)
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

	err = store.UpdatePlugin(plugin)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update plugin in db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Failed to update plugin in db",
			http.StatusInternalServerError,
		)
	}
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
		other.HttpErr(w, ErrIdBadRequest, "Plugin ID must be a uint", ErrIdBadRequest)
		return
	}

	// TODO: Add logging: About to attempt plugin deletion with plugin and user id
	err = store.DeletePlugin(uint(pluginID), *accId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete plugin in db")
		other.HttpErr(
			w,
			ErrIdDbErr,
			"Failed to delete plugin in db",
			http.StatusInternalServerError,
		)
	}
}
