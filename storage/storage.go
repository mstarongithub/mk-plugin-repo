package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
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
		log.Info().Msg("No gorm config provided, using default")
		customConfig = &gorm.Config{
			Logger: logger.New(&log.Logger, logger.Config{
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
		log.Error().Err(err).Str("db-file", sqliteFile).Msg("Failed to read database file")
		return storage, fmt.Errorf("failed to read database file %s: %w", sqliteFile, err)
	}
	// Migrate stuff
	log.Debug().Msg("Applying auto-migrations")
	err = db.AutoMigrate(
		&Account{},
		&Plugin{},
		&PluginVersion{},
		&Token{},
		&PasskeySession{},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to apply auto-migrations")
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
