package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Storage struct {
	db     *gorm.DB
	tokens map[uint][]string
}

var ErrVersionNotFound = errors.New("version not found")
var ErrPluginNotFound = errors.New("plugin not found")
var ErrAlreadyExists = errors.New("entry already exists")
var ErrUnknown = errors.New("unknown problem occured")
var ErrUnauthorised = errors.New("action is unauthorised")

func NewStorage(sqliteFile string, customConfig *gorm.Config) (storage Storage, err error) {
	if customConfig == nil {
		logrus.Infoln("No gorm config provided, using default")
		customConfig = &gorm.Config{
			Logger: logger.New(logrus.StandardLogger(), logger.Config{
				SlowThreshold: time.Second,
				LogLevel:      logger.Error,
				Colorful:      false,
			}),
		}
	}
	db, err := gorm.Open(sqlite.Open(sqliteFile), customConfig)
	if err != nil {
		// TODO: Add logging
		return storage, fmt.Errorf("failed to read database file %s: %w", sqliteFile, err)
	}
	// Migrate stuff
	// TODO: Add logging
	err = db.AutoMigrate(
		&Account{},
		&Plugin{},
		&PluginVersion{},
	)
	if err != nil {
		// TODO: Add logging
		return storage, fmt.Errorf("migration failed: %w", err)
	}
	// TODO: Add logging
	db.FirstOrCreate(&Account{
		Model: gorm.Model{
			ID: 12345,
		},
		Approved: true,
	})
	storage.db = db
	storage.tokens = map[uint][]string{}
	return storage, nil
}
