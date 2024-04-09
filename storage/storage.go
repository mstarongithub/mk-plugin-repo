package storage

import (
	"context"
	"fmt"

	"github.com/volatiletech/authboss/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(sqliteFile string, customConfig *gorm.Config) (storage Storage, err error) {
	db, err := gorm.Open(sqlite.Open(sqliteFile), customConfig)
	if err != nil {
		return storage, fmt.Errorf("failed to read database file %s: %w", sqliteFile, err)
	}
	// Migrate stuf
	// Migrate stufff
	err = db.AutoMigrate(
		&Account{},
		&Plugin{},
		&PluginVersion{},
	)
	if err != nil {
		return storage, fmt.Errorf("migration failed: %w", err)
	}
	storage.db = db
	return storage, nil
}

// Authboss ServerStorer interface implementation
func (storage *Storage) Load(ctx context.Context, key string) (authboss.User, error) {
	return nil, fmt.Errorf("unimplemented") // FIX: Implement me!
}

// Authboss ServerStorer interface implementation
func (storage *Storage) Save(ctx context.Context, user authboss.User) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}
