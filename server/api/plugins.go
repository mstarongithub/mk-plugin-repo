package api

import "net/http"

// Data expected for making a new plugin via POST /api/v1/plugins
type NewPluginData struct {
	Name           string   `json:"name"`            // Name of the plugin
	SummaryShort   string   `json:"summary_short"`   // A short summary
	SummaryLong    string   `json:"summary_long"`    // A full description of the plugin
	InitialVersion string   `json:"initial_version"` // The version of this new plugin
	Code           string   `json:"code"`            // The code associated with this new plugin and version
	Tags           []string `json:"tags"`            // The tags this plugin falls under
	Type           string   `json:"type"`            // What type the plugin is. Valid values are "plugin" and "widget"
}

// Data a request to read a plugin returns (GET /api/v1/plugins -> Array of this, GET /api/v1/plugins/{plugin-id} -> One instance)
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
	ID           uint     `json:"id"`            // The unique ID of the plugin
	Name         string   `json:"name"`          // Name of the plugin
	SummaryShort string   `json:"summary_short"` // A short summary
	SummaryLong  string   `json:"summary_long"`  // A full description of the plugin
	Tags         []string `json:"tags"`          // The tags this plugin falls under
	Type         string   `json:"type"`          // What type the plugin is. Valid values are "plugin" and "widget"
}

// GET /api/v1/plugins
// Get a list of plugins. May be non-exhaustive and uses paging
// Optional GET parameters:
// - name: search for plugins containing the value in their name
// - content: search for plugins containing the value in their description
// - page: which "page" to select of the list of plugins
// - tags: semicolon separated list of tags that must be included
func GetPluginList(w http.ResponseWriter, r *http.Request) {}

// POST /api/v1/plugins
// RESTRICTED
// Add a new plugin to the repo. Requires user authentication
// New plugins will only be available after approval from an admin
// Body must be a json version of NewPluginData
func AddNewPlugin(w http.ResponseWriter, r *http.Request) {}

// GET /api/v1/plugins/{plugin-id}
// Get a specific plugin, specified by {plugin-id}
func GetSpecificPlugin(w http.ResponseWriter, r *http.Request) {}

// PUT /api/v1/plugins/{plugin-id}
// RESTRICTED
// Update a specific plugin
func UpdateSpecificPlugin(w http.ResponseWriter, r *http.Request) {}
