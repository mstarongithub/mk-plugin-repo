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

func StorageFromRequest(r *http.Request) *storage.Storage {
	store, ok := r.Context().Value(CONTEXT_KEY_STORAGE).(*storage.Storage)
	if !ok {
		store = nil
	}
	return store
}

func ServerFromRequest(r *http.Request) *Server {
	store, ok := r.Context().Value(CONTEXT_KEY_SERVER).(*Server)
	if !ok {
		store = nil
	}
	return store
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
