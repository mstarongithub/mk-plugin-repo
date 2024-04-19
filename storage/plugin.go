package storage

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
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

func (storage *Storage) GetAllPlugins() []Plugin {
	logrus.Debugln("Attempting to get all plugins")
	plugins := []Plugin{}
	storage.db.Find(&plugins)
	logrus.WithField("plugins", plugins).Debugln("Plugins found")
	return plugins
}

func (storage *Storage) GetPluginByID(pluginID uint) (*Plugin, error) {
	// TODO: Add logging
	plugin := Plugin{}
	result := storage.db.First(&plugin, pluginID)

	if result.RowsAffected == 0 {
		// TODO: Add logging
		return nil, ErrPluginNotFound
	}
	if result.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf(
			"error while getting entry for id %d from db: %w",
			pluginID,
			result.Error,
		)
	}

	return &plugin, nil
}

// TODO: Complete this
func (storage *Storage) PushNewPluginVersion(
	pluginID uint,
	versionName string,
	code string,
) (*Plugin, error) {
	plugin, err := storage.GetPluginByID(pluginID)
	if err != nil {
		return nil, err
	}
	_, err = storage.TryFindVersion(pluginID, versionName)
	if !errors.Is(err, ErrVersionNotFound) {
	}
	_ = plugin // FIXME: Only done so it compiles for now
	return nil, nil
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
	// Check that account exists
	acc, err := storage.FindAccountByID(authorID)
	if err != nil {
		// TODO: Add logging
		return nil, err
	}
	if !acc.Approved {
		// TODO: Add logging
		return nil, ErrAccountNotApproved
	}

	logrus.WithField("plugin", plugin).Debugln("Inserting new plugin")
	res := storage.db.Debug().Create(&plugin)
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("rows affected during plugin creation were 0: %w", err)
	} else if res.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf("error while creating new plugin (data: %#v) in db: %w", plugin, err)
	}

	err = storage.NewVersion(plugin.ID, firstVersion, code)
	if err != nil {
		return nil, fmt.Errorf("error while creating first plugin version: %w", err)
	}

	return &plugin, nil
}

func (storage *Storage) UpdatePlugin(newPlugin *Plugin) error {
	// TODO: Add logging
	res := storage.db.Save(newPlugin)
	return res.Error
}

func (storage *Storage) DeletePlugin(pluginID, authorID uint) error {
	// TODO: Add logging
	plugin, err := storage.GetPluginByID(pluginID)
	if err != nil && !errors.Is(err, ErrPluginNotFound) {
		// TODO: Add logging
		return err
	}
	if plugin.AuthorID != authorID {
		// TODO: Add logging
		return ErrUnauthorised
	}
	storage.db.Delete(plugin)
	return nil
}
