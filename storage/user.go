package storage

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/mstarongithub/passkey"
	"github.com/rs/zerolog/log"
	"gitlab.com/mstarongitlab/goutils/sliceutils"
	"gorm.io/gorm"

	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

// A user account. Profile images are matched by ID
type Account struct {
	gorm.Model

	// ---- Section User data
	Name        string                           // Name of the account, NOT THE ID
	Mail        *string                          // Email linked to the account
	Links       customtypes.GenericSlice[string] `gorm:"serializer:json"`
	Description string                           // A description of the account, added by the user. Not necessary

	// ---- Section access control
	CanApprovePlugins bool // Can this account approve new plugin requests?
	CanApproveUsers   bool // Can this account approve account creation requests?
	// Custom links the user has added to the account
	// This can be things like Fedi profile, personal webpage, etc
	PluginsOwned customtypes.GenericSlice[uint] // IDs of plugins this account owns (has created)
	Approved     bool                           // Is this account approved for performing any actions

	// ---- Section Authentication
	// AuthMethods  customtypes.AuthMethods
	// PasswordHash []byte // The hash of the user's password, if they have one
	// FidoToken    *string
	// TotpToken    *string

	// ---- Section Passkeys
	PasskeyId   []byte
	Credentials []webauthn.Credential `gorm:"serializer:json"`
}

var ErrAccountNotFound = errors.New("account not found")
var ErrAccountNotApproved = errors.New("account not approved for this action")

func (s *Storage) FindAccountByName(name string) (*Account, error) {
	logger := log.With().Str("account-name", name).Logger()
	logger.Debug().Msg("Looking for account")
	acc := Account{}

	res := s.db.Where("name = ?", name).First(&acc)
	if res.RowsAffected == 0 || errors.Is(res.Error, gorm.ErrRecordNotFound) {
		logger.Debug().Msg("No account found")
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		logger.Error().Err(res.Error).Msg("Problem while looking for account")
		return nil, fmt.Errorf("error while searching for account %s: %w", name, res.Error)
	}
	logger.Info().Msg("Found account")
	return &acc, nil
}

func (s *Storage) FindAccountByID(id uint) (*Account, error) {
	acc := Account{}
	logger := log.With().Uint("account-id", id).Logger()
	logger.Debug().Msg("Looking for account")
	res := s.db.First(&acc, id)
	if res.RowsAffected == 0 || errors.Is(res.Error, gorm.ErrRecordNotFound) {
		logger.Debug().Msg("No account found")
		return nil, ErrAccountNotFound
	} else if res.Error != nil {
		logger.Error().Err(res.Error).Msg("Problem while looking for account")
		return nil, fmt.Errorf("problem while finding account id %d: %w", id, res.Error)
	}
	logger.Info().Msg("Found account")
	return &acc, nil
}

func (s *Storage) FindAccountByPasskeyId(pkeyId []byte) (*Account, error) {
	acc := Account{}
	logger := log.With().Bytes("account-passkey-id", pkeyId).Logger()
	logger.Debug().Msg("Looking for account")
	err := s.db.Where(Account{PasskeyId: pkeyId}).First(&acc).Error
	switch err {
	case nil:
		logger.Info().Msg("Found account")
		return &acc, nil
	case gorm.ErrRecordNotFound:
		logger.Info().Msg("No account found")
		return nil, ErrAccountNotFound
	default:
		logger.Error().Err(err).Msg("Problem while looking for account")
		return nil, err
	}
}

func (s *Storage) AddNewAccount(acc Account) (uint, error) {
	log.Debug().Any("account-full", &acc).Msg("Adding new account")
	res := s.db.Create(&acc)
	if res.Error != nil {
		log.Error().Err(res.Error).Any("account-full", &acc).Msg("Failed to add new account")
		return 0, res.Error
	}
	log.Info().Uint("account-id", acc.ID).Msg("New account added")
	return acc.ID, nil
}

func (s *Storage) UpdateAccount(acc *Account) error {
	log.Debug().Uint("account-id", acc.ID).Msg("Updating account")
	res := s.db.Save(acc)
	if res.Error != nil {
		log.Error().Err(res.Error).Uint("account-id", acc.ID).Msg("Failed to update account")
	} else {
		log.Info().Uint("account-id", acc.ID).Msg("Updated account")
	}
	return res.Error
}

func (s *Storage) GetAllUnapprovedAccounts() ([]Account, error) {
	accs := []Account{}
	log.Debug().Msg("Looking for all unapproved accounts")
	res := s.db.Where("approved = ?", false).Find(&accs)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Error().Err(res.Error).Msg("Failed to get unapproved accounts from db")
		return nil, res.Error
	}
	log.Info().
		Uints("account-ids", sliceutils.Map(accs, func(t Account) uint { return t.ID })).
		Msg("Found unapproved accounts")
	return accs, nil
}

func (s *Storage) DeleteAccount(id uint) {
	log.Debug().Uint("account-id", id).Msg("Deleting account")
	s.db.Delete(&Account{}, id)
	log.Info().Uint("account-id", id).Msg("Account deleted")
}

// ---- Section webauthn.User

func (u *Account) WebAuthnID() []byte {
	return u.PasskeyId
}

func (u *Account) WebAuthnName() string {
	return u.Name
}

func (u *Account) WebAuthnDisplayName() string {
	return u.Name
}

func (u *Account) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *Account) WebAuthnIcon() string {
	return ""
}

// ---- Section passkey.User

func (u *Account) PutCredential(new webauthn.Credential) {
	u.Credentials = append(u.Credentials, new)
}

// Section passkey.UserStore

func (s *Storage) GetOrCreateUser(userID string) passkey.User {
	log.Debug().Str("account-name", userID).Msg("Searching or creating user for passkey stuff")
	acc := &Account{}
	s.db.Model(&Account{}).Where("name = ?", userID).FirstOrCreate(acc)
	if len(acc.PasskeyId) == 0 {
		log.Debug().
			Str("account-name", userID).
			Msg("Account doesn't have a passkey id yet, creating one")
		data := make([]byte, 64)
		c, err := rand.Read(data)
		for err != nil || c != len(data) || c < 64 {
			data = make([]byte, 64)
			c, err = rand.Read(data)
		}
		acc.PasskeyId = data
	}
	acc.Name = userID
	s.db.Save(acc)
	log.Info().Uint("account-id", acc.ID).Msg("Found or created account for passkey stuff")
	return acc
}

func (s *Storage) GetUserByWebAuthnId(id []byte) passkey.User {
	acc := &Account{}
	log.Debug().Msg("looking for account with passkey id")
	res := s.db.Model(acc).Where("passkey_id = ?", id).First(acc)
	if res.Error != nil {
		log.Error().
			Err(res.Error).
			Bytes("passkey-id", id).
			Msg("Failed to find account with passkey id")
		return nil
	}
	log.Info().Uint("account-id", acc.ID).Msg("Found account by passkey id")
	return acc
}

func (s *Storage) SaveUser(rawUser passkey.User) {
	user, ok := rawUser.(*Account)
	if !ok {
		log.Error().
			Any("raw-account", rawUser).
			Msg("Given passkey user couldn't be cast into a db user")
		return
	}
	// NOTE: This is probably not a good idea. Not at all I think. But it should work
	// If the user being saved has id of 1, give that user superuser permissions.
	// I don't particularly want to make an entirely custom version of the passkey lib
	// just to provide password protection for the initial setup of su
	if user.ID == 1 {
		user.Approved = true
		user.CanApprovePlugins = true
		user.CanApproveUsers = true

	}
	s.db.Save(user)
	log.Info().Uint("account-id", user.ID).Msg("Updated account from passkey data")
}
