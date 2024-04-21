package storage

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.com/mstarongitlab/goutils/sliceutils"
	"gorm.io/gorm"
)

type PluginVersion struct {
	gorm.Model
	Version         string `gorm:"version;<-create"`           // The version string
	Code            string `gorm:"code;<-:create"`             // Raw code for this version
	PluginID        uint   `gorm:"plugin_id;<-:create"`        // The plugin ID this version belongs to
	AiScriptVersion string `gorm:"aiscript_version;<-:create"` // The targeted AIScript version this plugin version was made for
}

var ErrVersionAlreadyExists = errors.New("version already exists")

// Get all versions for a plugin
// Will return empty list if that plugin doesn't exist
// Will return empty list if that plugin doesn't exist
func (storage *Storage) GetVersionsFor(pluginID uint) []PluginVersion {
	logrus.WithFields(logrus.Fields{
		"pluginID": pluginID,
		"source":   "storage.GetVersionsFor",
	}).
		Debugln("storage: Attempting to get versions for plugin")
	plugins := []PluginVersion{}
	result := storage.db.Find(plugins, "plugin_id = ?", pluginID)
	if result.Error != nil {
		logrus.WithFields(logrus.Fields{
			"pluginID": pluginID,
			"source":   "storage.GetVersionsFor",
		}).
			WithError(result.Error).
			Warnln("storage: Error while getting versions for plugin")
	}
	logrus.WithFields(logrus.Fields{
		"plugins.len": len(plugins),
		"plugins.versions": sliceutils.Map(
			plugins,
			func(plugin PluginVersion) string { return plugin.Version },
		),
		"pluginID": pluginID,
	}).Debugln("found versions")
	return plugins
}

// Try and find a plugin version for the given plugin ID and version name
func (storage *Storage) TryFindVersion(pluginID uint, versionName string) (*PluginVersion, error) {
	version := PluginVersion{
		Version:  versionName,
		PluginID: pluginID,
	}
	// TODO: Add logging
	result := storage.db.First(&version)
	if result.RowsAffected < 1 {
		// TODO: Add logging
		return nil, ErrVersionNotFound
	}
	if result.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf(
			"problem getting the first matching entry for version %q of plugin with ID %d: %w",
			versionName,
			pluginID,
			result.Error,
		)
	}
	// TODO: Add logging
	return &version, nil
}

// Hide/Disable a specific version of a plugin. This doesn't delete it, but makes it unavailable
func (storage *Storage) HideVersion(pluginID uint, versionName string) error {
	// TODO: Add logging
	version, err := storage.TryFindVersion(pluginID, versionName)
	if err != nil {
		// TODO: Add logging
		if errors.Is(err, ErrVersionNotFound) {
			return nil
		} else {
			return err
		}
	}
	storage.db.Delete(version)
	// TODO: Add logging
	return nil
}

func (storage *Storage) NewVersion(
	forPluginID uint,
	versionName, code, aiscript_version string,
) error {
	// First check if a version already exists
	_, err := storage.TryFindVersion(forPluginID, versionName)
	if !errors.Is(err, ErrVersionNotFound) {
		// TODO: Add logging
		return ErrAlreadyExists
	}

	// Then check if there actually is a plugin with the given ID
	_, err = storage.GetPluginByID(forPluginID)
	if err != nil {
		// TODO: Add logging
		return err
	}

	// Now make the new version, push it to the db
	newVersion := PluginVersion{
		PluginID: forPluginID,
		Version:  versionName,
		Code:     code,
	}
	// TODO: Add logging
	result := storage.db.Create(&newVersion)
	if result.Error != nil {
		// TODO: Add logging
		return fmt.Errorf("error trying to create new version: %w", result.Error)
	}

	// And update the parent plugin
	_, err = storage.PushNewPluginVersion(forPluginID, versionName, code)
	if err != nil {
		// TODO: Add logging
		return fmt.Errorf("failed to update plugin info: %w", err)
	}

	// TODO: Add logging
	return nil
}
