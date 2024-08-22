package storage

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
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
	log.Debug().Uint("plugin-id", pluginID).Msg("Grabbing plugin versions")
	plugins := []PluginVersion{}
	result := storage.db.Find(plugins, "plugin_id = ?", pluginID)
	if result.Error != nil {
		log.Warn().
			Err(result.Error).
			Uint("plugin-id", pluginID).
			Msg("No versions for a non-existing plugin")
	}
	log.Debug().Uint("plugin-id", pluginID).Int("version-count", len(plugins)).Msg("Found versions")
	return plugins
}

// Try and find a plugin version for the given plugin ID and version name
func (storage *Storage) TryFindVersion(pluginID uint, versionName string) (*PluginVersion, error) {
	version := PluginVersion{
		Version:  versionName,
		PluginID: pluginID,
	}
	logger := log.With().
		Uint("plugin-id", pluginID).
		Str("version-name", versionName).
		Logger()
	logger.Debug().Msg("Looking for version of plugin")
	result := storage.db.Where("version = ?", versionName).
		Where("plugin_id = ?", pluginID).
		First(&version)
	if result.RowsAffected < 1 || errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logger.Debug().Msg("Version not found")
		return nil, ErrVersionNotFound
	}
	if result.Error != nil {
		logger.Warn().Err(result.Error).Msg("Problem getting version for plugin")
		return nil, fmt.Errorf(
			"problem getting the first matching entry for version %q of plugin with ID %d: %w",
			versionName,
			pluginID,
			result.Error,
		)
	}
	logger.Debug().Msg("Found version")
	return &version, nil
}

// TODO: Rename to DeleteVersion
func (storage *Storage) DeleteVersion(pluginID uint, versionName string) error {
	logger := log.With().Uint("plugin-id", pluginID).Str("version-name", versionName).Logger()
	logger.Debug().Msg("Looking for version to delete")
	version, err := storage.TryFindVersion(pluginID, versionName)
	if err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			logger.Debug().Msg("Version doesn't exist in the first place")
			return nil
		} else {
			logger.Error().Err(err).Msg("Problem getting version for deletion")
			return err
		}
	}

	logger.Debug().Msg("Updating parent plugin while deleting version")
	if err = storage.HidePluginVersionFromPlugin(pluginID, versionName); err != nil {
		logger.Error().Err(err).Msg("Can't update parent plugin during version deletion")
		return fmt.Errorf("can't update parent plugin: %w", err)
	}

	storage.db.Delete(version)
	logger.Debug().Msg("Version deleted")
	return nil
}

func (storage *Storage) NewVersion(
	forPluginID uint,
	versionName, code, aiscript_version string,
) error {
	logger := log.With().Uint("plugin-id", forPluginID).Str("version-name", versionName).Logger()
	logger.Debug().Msg("Attempting to insert new version for plugin")
	// First check if a version already exists
	_, err := storage.TryFindVersion(forPluginID, versionName)
	if err == nil {
		logger.Warn().
			Msg("No error received while checking for existence of version. Assuming it already exists and aborting")
		return ErrAlreadyExists
	} else if !errors.Is(err, ErrVersionNotFound) {
		logger.Error().Err(err).Msg("Got error while trying to verify that version doesn't exist already")
		return err
	}

	// Then check if there actually is a plugin with the given ID
	_, err = storage.GetPluginByID(forPluginID)
	if err != nil {
		logger.Error().Err(err).Msg("Error while getting parent plugin for new version")
		return err
	}

	// Now make the new version, push it to the db
	newVersion := PluginVersion{
		PluginID:        forPluginID,
		Version:         versionName,
		Code:            code,
		AiScriptVersion: aiscript_version,
	}
	logger.Debug().Msg("Inserting new plugin version into db")
	result := storage.db.Create(&newVersion)
	if result.Error != nil {
		logger.Error().Err(result.Error).Msg("Problem while creating new version")
		return fmt.Errorf("error trying to create new version: %w", result.Error)
	}

	// And update the parent plugin
	logger.Debug().Msg("Updating parent plugin to point at new version")
	_, err = storage.PushNewPluginVersion(forPluginID, versionName)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to update parent plugin for new version")
		return fmt.Errorf("failed to update plugin info: %w", err)
	}

	logger.Debug().Msg("New version applied")
	return nil
}
