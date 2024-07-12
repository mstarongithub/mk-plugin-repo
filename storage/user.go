package storage

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// A user account. Profile images are matched by ID
type Account struct {
	gorm.Model

	// ---- Section User data
	Name        string  // Name of the account, NOT THE ID
	Mail        *string // Email linked to the account
	Links       customtypes.GenericSlice[string]
	Description string // A description of the account, added by the user. Not necessary

	// ---- Section access control
	CanApprovePlugins bool // Can this account approve new plugin requests?
	CanApproveUsers   bool // Can this account approve account creation requests?
	// Custom links the user has added to the account
	// This can be things like Fedi profile, personal webpage, etc
	PluginsOwned customtypes.GenericSlice[uint] // IDs of plugins this account owns (has created)
	Approved     bool                           // Is this account approved for performing any actions

	// ---- Section Authentication
	AuthMethods  customtypes.AuthMethods
	PasswordHash []byte // The hash of the user's password, if they have one
	FidoToken    *string
	TotpToken    *string
	Passkeys     map[string]string `gorm:"serializer:json"`
}

var ErrAccountNotFound = errors.New("account not found")
var ErrAccountNotApproved = errors.New("account not approved for this action")

func (s *Storage) FindAccountByName(name string) (*Account, error) {
	// TODO: Add logging
	acc := Account{}

	res := s.db.First(&acc, "name = ?", name)
	if res.RowsAffected == 0 {
		// TODO: Add logging
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf("error while searching for account %s: %w", name, res.Error)
	}
	// TODO: Add logging
	return &acc, nil
}

func (s *Storage) FindAccountByID(id uint) (*Account, error) {
	acc := Account{}
	// TODO: Add logging
	res := s.db.First(&acc, id)
	if res.RowsAffected == 0 {
		// TODO: Add logging
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		// TODO: Add logging
		return nil, fmt.Errorf("problem while finding account id %d: %w", id, res.Error)
	}
	// TODO: Add logging
	return &acc, nil
}

func (s *Storage) AddNewAccount(acc Account) (uint, error) {
	res := s.db.Create(&acc)
	if res.Error != nil {
		return 0, res.Error
	}
	return acc.ID, nil
}

func (s *Storage) UpdateAccount(acc *Account) error {
	res := s.db.Save(acc)
	return res.Error
}

func (s *Storage) GetAllUnapprovedAccounts() ([]Account, error) {
	accs := []Account{}
	res := s.db.Where("approved = ", false).Find(&accs)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, res.Error
	}
	return accs, nil
}

func (s *Storage) DeleteAccount(id uint) {
	s.db.Delete(&Account{}, id)
}
