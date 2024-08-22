package storage

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/sliceutils"
	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// Also used for widgets
type Plugin struct {
	gorm.Model
	CurrentVersion   string                 // The current version string
	PreviousVersions []string               `gorm:"serializer:json"` // List of all previous versions
	Name             string                 // The name of the plugin
	SummaryShort     string                 // A short description for this plugin
	SummaryLong      string                 // Full summary for this plugin
	AuthorID         uint                   // ID of the author
	Tags             []string               `gorm:"serializer:json"` // Tags for this plugin
	Type             customtypes.PluginType // What type of plugin this is. Normal plugin or widget are the only options currently
	Approved         bool                   // Got this plugin approved for publishing?
}

var ErrPluginMustHaveAtLeastOneVersion = errors.New("plugins must have at least one version")

func (storage *Storage) GetAllPlugins() []Plugin {
	log.Debug().Msg("Collecting all plugins")
	plugins := []Plugin{}
	storage.db.Where("approved = ?", true).Find(&plugins)
	log.Debug().
		Strs("plugin-ids", sliceutils.Map(plugins, func(t Plugin) string { return t.Name })).
		Msg("Plugins collected")
	return plugins
}

func (storage *Storage) GetPluginByID(pluginID uint) (*Plugin, error) {
	logger := log.With().Uint("plugin-id", pluginID).Logger()
	logger.Debug().Msg("searching for plugin")
	plugin := Plugin{}
	result := storage.db.First(&plugin, pluginID)

	if result.RowsAffected == 0 {
		logger.Debug().Msg("Didn't find one")
		return nil, ErrPluginNotFound
	}
	if result.Error != nil {
		logger.Error().Err(result.Error).Msg("Database error")
		return nil, fmt.Errorf(
			"error while getting entry for id %d from db: %w",
			pluginID,
			result.Error,
		)
	}

	logger.Debug().Msg("Found it")
	return &plugin, nil
}

// Tell a plugin that a new version has been added
func (storage *Storage) PushNewPluginVersion(
	pluginID uint,
	versionName string,
) (*Plugin, error) {
	logger := log.With().Str("version-name", versionName).Uint("plugin-id", pluginID).Logger()
	logger.Debug().Msg("Searching for plugin to update the current version")
	plugin, err := storage.GetPluginByID(pluginID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to find plugin for updating the current version")
		return nil, err
	}
	_, err = storage.TryFindVersion(pluginID, versionName)
	if err != nil {
		logger.Error().Err(err).Msg("Target version not found for setting new current one")
		return nil, err
	}
	plugin.PreviousVersions = append(plugin.PreviousVersions, versionName)
	plugin.CurrentVersion = versionName
	res := storage.db.Save(plugin)
	logger.Debug().Err(res.Error).Msg("Set new current version for plugin")
	return nil, res.Error
}

func (storage *Storage) NewPlugin(
	name string,
	authorID uint,
	firstVersion,
	SummaryLong,
	SummaryShort string,
	tags []string,
	pluginType customtypes.PluginType,
	code string,
	aiscriptVersion string,
) (*Plugin, error) {
	plugin := Plugin{
		CurrentVersion:   firstVersion,
		PreviousVersions: make(customtypes.GenericSlice[string], 0),
		Name:             name,
		SummaryShort:     SummaryShort,
		SummaryLong:      SummaryLong,
		AuthorID:         authorID,
		Tags:             customtypes.GenericSlice[string](tags),
		Type:             pluginType,
		Approved:         false,
	}
	log.Debug().Uint("author-id", authorID).Msg("Looking if account for new plugin exists")
	// Check that account exists
	acc, err := storage.FindAccountByID(authorID)
	if err != nil {
		log.Error().
			Uint("author-id", authorID).
			Err(err).
			Msg("Problem finding account for new plugin")
		return nil, err
	}
	if !acc.Approved {
		log.Warn().Uint("author-id", authorID).Msg("Account for new plugin not approved yet")
		return nil, ErrAccountNotApproved
	}

	// Check if version already exists
	placeholder := Plugin{}
	storage.db.First(&placeholder, "name = ?", name)
	if placeholder.ID != 0 {
		log.Warn().Str("plugin-name", name).Msg("A plugin with that name already exists")
		return nil, ErrAlreadyExists
	}

	log.Debug().Any("plugin-full", &plugin).Msg("Attempting to insert new plugin into db")
	res := storage.db.Create(&plugin)
	if res.RowsAffected == 0 {
		log.Error().
			Any("plugin-full", &plugin).
			Err(res.Error).
			Msg("Zero rows were affected while inserting new plugin")
		return nil, fmt.Errorf("rows affected during plugin creation were 0: %w", err)
	} else if res.Error != nil {
		log.Error().Err(res.Error).Any("plugin-full", &plugin).Msg("Failed to insert new plugin into db")
		return nil, fmt.Errorf("error while creating new plugin (data: %#v) in db: %w", plugin, err)
	}

	// No extra logging, new version does all of that
	err = storage.NewVersion(plugin.ID, firstVersion, code, aiscriptVersion)
	if err != nil {
		log.Error().
			Err(err).
			Uint("plugin-id", plugin.ID).
			Str("version-name", firstVersion).
			Msg("Failed to insert first version")
		return nil, fmt.Errorf("error while creating first plugin version: %w", err)
	}
	log.Debug().Any("plugin-full", &plugin).Msg("New plugin created")
	return &plugin, nil
}

func (storage *Storage) UpdatePlugin(newPlugin *Plugin) error {
	log.Debug().Any("plugin-full", newPlugin).Msg("Attempting to update plugin")
	res := storage.db.Save(newPlugin)
	if res.Error != nil {
		log.Error().Err(res.Error).Any("plugin-full", newPlugin).Msg("Failed to update plugin")
	} else {
		log.Debug().Uint("plugin-id", newPlugin.ID).Msg("Updated plugin")
	}

	return res.Error
}

func (storage *Storage) DeletePlugin(pluginID, authorID uint) error {
	log.Debug().Uint("plugin-id", pluginID).Msg("Attempting to delete plugin")
	plugin, err := storage.GetPluginByID(pluginID)
	if err != nil {
		if errors.Is(err, ErrPluginNotFound) {
			log.Debug().Uint("plugin-id", pluginID).Msg("Can't delete a plugin that doesn't exist")
			return nil
		}
		log.Error().Uint("plugin-id", pluginID).Err(err).Msg("Failed to get plugin for deletion")
		return err
	}
	log.Debug().Uint("authod-id", authorID).Msg("Fetching authorising account for plugin deletion")
	acc, err := storage.FindAccountByID(authorID)
	if err != nil {
		log.Error().
			Err(err).
			Uint("author-id", authorID).
			Msg("Error while getting authorizing account for plugin deletion")
		return err
	}
	if plugin.AuthorID != authorID && !acc.CanApprovePlugins {
		log.Error().
			Uint("plugin-id", pluginID).
			Uint("author-id", authorID).
			Msg("Authorising account for plugin deletion is neither plugin author nor a plugin admin")
		return ErrUnauthorised
	}
	log.Debug().Uint("plugin-id", pluginID).Msg("Removing plugin from db")
	storage.db.Delete(plugin)
	return nil
}

func (storage *Storage) HidePluginVersionFromPlugin(pluginID uint, versionName string) error {
	log.Debug().
		Uint("plugin-id", pluginID).
		Str("version-name", versionName).
		Msg("Starting process to hide a version from a plugin")
	plugin := Plugin{}
	res := storage.db.First(plugin, pluginID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			log.Debug().
				Uint("plugin-id", pluginID).
				Msg("Can't hide version from a plugin that doesn't exist")
			return ErrPluginNotFound
		}
	}
	// TODO: Add authorization check

	// Don't remove a version if said version is the only one this plugin has
	if len(plugin.PreviousVersions) == 1 {
		log.Warn().
			Uint("plugin-id", pluginID).
			Msg("Can't hide version of a plugin with only one version")
		return ErrPluginMustHaveAtLeastOneVersion
	}

	// Then first remove it from the slice of versions
	plugin.PreviousVersions = sliceutils.Filter(plugin.PreviousVersions, func(t string) bool {
		return t != versionName
	})
	// And finally downgrade the current version if it matches the version to delete
	if plugin.CurrentVersion == versionName {
		plugin.CurrentVersion = plugin.PreviousVersions[len(plugin.PreviousVersions)-1]
	}
	log.Debug().
		Uint("plugin-id", pluginID).
		Strs("all-versions", plugin.PreviousVersions).
		Str("current-version", plugin.CurrentVersion).
		Msg("Updating plugin with \"new\" versions")
	storage.db.Save(&plugin)

	return nil
}

func (storage *Storage) GetUnapprovedPlugins() ([]Plugin, error) {
	plugins := []Plugin{}
	log.Debug().Msg("Getting all plugins that haven't been approved yet")
	res := storage.db.Where("approved = ?", false).Find(&plugins)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, res.Error
	}
	return plugins, nil
}
