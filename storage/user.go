package storage

import (
	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// A user account. Profile images are matched by ID
type Account struct {
	gorm.Model               // Default data for stuff
	CanApprovePlugins bool   // Can this account approve new plugin requests?
	CanApproveUsers   bool   // Can this account approve account creation requests?
	Name              string // Name of the account, NOT THE ID
	Mail              string // Mail address of the account
	PasswordHash      []byte // Hash of the password
	Salt              []byte // Salt added to the password hash
	// Custom links the user has added to the account
	// This can be things like Fedi profile, personal webpage, etc
	Links        customtypes.GenericSlice[string]
	Description  string                         // A description of the account, added by the user. Not necessary
	PluginsOwned customtypes.GenericSlice[uint] // IDs of plugins this account owns (has created)
	Approved     bool                           // Is this account approved for performing any actions
}
