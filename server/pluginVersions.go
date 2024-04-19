package server

import "net/http"

type VersionData struct {
	Code string `json:"code"`
}

// GET /api/v1/plugins/{pluginId}/{version-name}
// Get the details for a specific version
// Returns a json formatted VersionData on success
// TODO: Implement me!
func getVersion(w http.ResponseWriter, r *http.Request) {}

// POST /api/v1/plugins/{pluginId}/{version-name}
// RESTRICTED
// Create a new version
// Expects json formatted VersionData
// {version-name} will be the name of the new version
// Returns 4xx (whatever the bad request status is) if the version already exists
// TODO: Implement me!
func newVersion(w http.ResponseWriter, r *http.Request) {}

// DELETE /api/v1/plugins/{pluginId}/{version-name}
// RESTRICTED
// Hide a version. Doesn't delete, just hides it from the API
// TODO: Implement me!
func hideVersion(w http.ResponseWriter, r *http.Request) {}
