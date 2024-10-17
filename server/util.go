package server

import (
	"net/http"

	"github.com/mstarongithub/mk-plugin-repo/storage"
	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

const (
	PLUGIN_TYPE_PLUGIN  = "plugin"
	PLUGIN_TYPE_WIDGET  = "widget"
	PLUGIN_TYPE_INVALID = "invalid"
)

func StorageFromRequest(w http.ResponseWriter, r *http.Request) *storage.Storage {
	store, ok := r.Context().Value(CONTEXT_KEY_STORAGE).(*storage.Storage)
	if !ok {
		http.Error(w, "no storage in request context", http.StatusInternalServerError)
		return nil
	}
	return store
}

func ServerFromRequest(w http.ResponseWriter, r *http.Request) *Server {
	store, ok := r.Context().Value(CONTEXT_KEY_SERVER).(*Server)
	if !ok {
		http.Error(w, "no server in request context", http.StatusInternalServerError)
		return nil
	}
	return store
}

// func AuthFromRequestContext(w http.ResponseWriter, r *http.Request) *auth.Auth {
// 	a, ok := r.Context().Value(CONTEXT_KEY_AUTH_LAYER).(*auth.Auth)
// 	if !ok {
// 		http.Error(w, "no auth in request context", http.StatusInternalServerError)
// 		return nil
// 	}
// 	return a
// }

func AccIdFromRequestContext(w http.ResponseWriter, r *http.Request) *uint {
	a, ok := r.Context().Value(CONTEXT_KEY_ACTOR_ID).(uint)
	if !ok {
		http.Error(w, "no account ID in request context", http.StatusInternalServerError)
		return nil
	}
	return &a
}

func dbPluginToApiPlugin(plugin *storage.Plugin) Plugin {
	newPlugin := Plugin{
		ID:             plugin.Model.ID,
		Name:           plugin.Name,
		SummaryShort:   plugin.SummaryShort,
		SummaryLong:    plugin.SummaryLong,
		CurrentVersion: plugin.CurrentVersion,
		AllVersions:    plugin.PreviousVersions,
		Tags:           plugin.Tags,
		AuthorID:       plugin.AuthorID,
	}
	switch plugin.Type {
	case customtypes.PLUGIN_TYPE_PLUGIN:
		newPlugin.Type = PLUGIN_TYPE_PLUGIN
	case customtypes.PLUGIN_TYPE_WIDGET:
		newPlugin.Type = PLUGIN_TYPE_WIDGET
	default:
		newPlugin.Type = PLUGIN_TYPE_INVALID
	}
	return newPlugin
}
