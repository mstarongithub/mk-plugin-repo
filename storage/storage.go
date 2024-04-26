package storage

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/volatiletech/authboss/v3"
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
		logrus.Infoln("No config provided, using default")
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
		Approved:  true,
		Confirmed: true,
	})
	storage.db = db
	storage.tokens = map[uint][]string{}
	return storage, nil
}

// ----- AUTHBOSS interface stuff

// Authboss ServerStorer interface implementation
func (storage *Storage) Load(_ context.Context, key string) (authboss.User, error) {
	uid, err := strconv.ParseUint(key, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("key not a uint: %w", err)
	}
	user, err := storage.FindAccountByID(uint(uid))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// Authboss ServerStorer interface implementation
func (storage *Storage) Save(ctx context.Context, user authboss.User) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) New(_ context.Context) authboss.User {
	return &Account{}
}

func (storage *Storage) Create(_ context.Context, abUser authboss.User) error {
	user, ok := abUser.(*Account)
	if !ok {
		return errors.New("failed to cast ab user to account")
	}
	if _, err := storage.FindAccountByID(user.Model.ID); err == nil ||
		!errors.Is(err, ErrAccountNotFound) {
		return authboss.ErrUserFound
	}
	logrus.WithField("user", user).Infoln("Saving new user")
	res := storage.db.Create(user)
	if res.Error != nil {
		return fmt.Errorf("failed to insert new user: %w", res.Error)
	}

	return nil
}

func (storage *Storage) LoadByConfirmSelector(
	_ context.Context,
	selector string,
) (user authboss.ConfirmableUser, err error) {
	return nil, fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) LoadByRecoverSelector(
	_ context.Context,
	selector string,
) (user authboss.RecoverableUser, err error) {
	return nil, fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) AddRememberToken(_ context.Context, pid, token string) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) DelRememberTokens(_ context.Context, pid string) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) UseRememberToken(_ context.Context, pid, token string) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) NewFromOAuth2(
	_ context.Context,
	provider string,
	details map[string]string,
) (authboss.OAuth2User, error) {
	return nil, fmt.Errorf("unimplemented") // FIX: Implement me!
}

func (storage *Storage) SaveOAuth2(_ context.Context, user authboss.OAuth2User) error {
	return fmt.Errorf("unimplemented") // FIX: Implement me!
}
