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
	CanApprovePlugins bool   // Can this account approve new plugin requests?
	CanApproveUsers   bool   // Can this account approve account creation requests?
	Name              string // Name of the account, NOT THE ID
	// Custom links the user has added to the account
	// This can be things like Fedi profile, personal webpage, etc
	Links        customtypes.GenericSlice[string]
	Description  string                         // A description of the account, added by the user. Not necessary
	PluginsOwned customtypes.GenericSlice[uint] // IDs of plugins this account owns (has created)
	Approved     bool                           // Is this account approved for performing any actions
	PasswordHash []byte                         // The hash of the user's password, if they have one
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
