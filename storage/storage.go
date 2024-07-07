package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/mstarongithub/mk-plugin-repo/util"
)

type Storage struct {
	db                   *gorm.DB
	tokens               util.MutexMap[uint, []string]
	serviceWorkersActive util.MutexMap[string, bool]
}

var ErrVersionNotFound = errors.New("version not found")
var ErrPluginNotFound = errors.New("plugin not found")
var ErrAlreadyExists = errors.New("entry already exists")
var ErrUnknown = errors.New("unknown problem occured")
var ErrUnauthorised = errors.New("action is unauthorised")

func NewStorage(sqliteFile string, customConfig *gorm.Config) (storage *Storage, err error) {
	if customConfig == nil {
		logrus.Infoln("No gorm config provided, using default")
		customConfig = &gorm.Config{
			Logger: logger.New(logrus.StandardLogger(), logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Error,
				Colorful:                  false,
				IgnoreRecordNotFoundError: true,
			}),
		}
	}
	storage = &Storage{
		serviceWorkersActive: util.NewMutexMap[string, bool](),
		tokens:               util.NewMutexMap[uint, []string](),
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
		&Token{},
	)
	if err != nil {
		// TODO: Add logging
		return storage, fmt.Errorf("migration failed: %w", err)
	}
	storage.db = db
	return storage, nil
}

func (storage *Storage) LaunchMiniServices() func() {
	exitChanOldTokens := make(chan any)
	exitChanOldData := make(chan any)
	go storage.serviceCleanOldTokens(exitChanOldTokens)
	go storage.serviceGdprCleanOldDeletedData(exitChanOldData)
	return func() {
		exitChanOldData <- 1
		exitChanOldTokens <- 1
		close(exitChanOldData)
		close(exitChanOldTokens)
	}
}
