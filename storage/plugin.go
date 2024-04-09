package storage

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/mstarongitlab/goutils/sliceutils"
	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// Also used for widgets
type Plugin struct {
	gorm.Model
	CurrentVersion   string                           // The current version string
	PreviousVersions customtypes.GenericSlice[string] // List of all previous versions
	Name             string                           // The name of the plugin
	SummaryShort     string                           // A short description for this plugin
	SummaryLong      string                           // Full summary for this plugin
	AuthorID         uint                             // ID of the author
	Tags             customtypes.GenericSlice[string] // Tags for this plugin
	Type             customtypes.PluginType           // What type of plugin this is. Normal plugin or widget are the only options currently
}

type PluginVersion struct {
	Version  string // The version string
	Code     string // Raw code for this version
	PluginID uint   // The plugin ID this version belongs to
}

func (storage *Storage) GetAllPlugins() []Plugin {
	logrus.Debugln("Attempting to get all plugins")
	plugins := []Plugin{}
	storage.db.Find(plugins)
	logrus.WithField("plugins", plugins).Debugln("Plugins found")
	return plugins
}

func (storage *Storage) GetVersionsFor(plugin *Plugin) []PluginVersion {
	logrus.WithFields(logrus.Fields{
		"plugin.name": plugin.Name,
		"plugin.id":   plugin.ID,
	}).Debugln("Attempting to get versions for plugin")
	plugins := []PluginVersion{}
	result := storage.db.Find(plugins, "plugin_id = ?", plugin.ID)
	if result.Error != nil {
		logrus.WithFields(logrus.Fields{
			"plugin.name": plugin.Name,
			"plugin.id":   plugin.ID,
		}).WithError(result.Error).Warnln("Error while getting versions for plugin")
	}
	logrus.WithFields(logrus.Fields{
		"plugins.len": len(plugins),
		"plugins.versions": sliceutils.Map(
			plugins,
			func(plugin PluginVersion) string { return plugin.Version },
		),
		"plugin.name": plugin.Name,
		"plugin.id":   plugin.ID,
	}).Debugln("found versions")
	return plugins
}
