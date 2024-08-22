package auth

import (
	"fmt"
	"net/url"

	"github.com/ermites-io/passwd"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type NextAuthState uint
type AuthProviderMode uint8

const AUTH_TOKEN_HEADER = "Authorization"
const _TOKEN_AUTH_MODE_NONE = "token-for-mode-no-auth"
const _TOKEN_DEV_ACCOUNT = "dev-account-token"
const _DEV_ACCOUNT_USERNAME = "developer"
const _DEV_ACCOUNT_PASSWORD = "developer"

type Auth struct {
	store              *storage.Storage
	webAuth            *webauthn.WebAuthn
	hasher             *passwd.Profile
	activeAuthRequests map[string]TempAuthRequest
	registerRequests   RegisterRequestHolder
	authMode           AuthProviderMode
	log                zerolog.Logger
}

type TempAuthRequest struct {
	AuthID    string
	AccountID uint // NOTE: Could replace this with a reference to the actual account struct later if db access times become a problem
	NextState NextAuthState
}

const (
	AUTH_SUCCESS = NextAuthState(0)
	AUTH_FAIL    = NextAuthState(1 << (iota - 1))
	AUTH_NEEDS_FIDO
	AUTH_NEEDS_TOTP
	AUTH_NEEDS_MAIL
)

const (
	AUTH_MODE_DEFAULT = AuthProviderMode(0)
	AUTH_MODE_DEV     = AuthProviderMode(1 << (iota - 1))
	AUTH_MODE_NONE
)

// Create a new authentication manager
// Requires a reference to a storage implementation
func NewAuth(store *storage.Storage, mode AuthProviderMode) (*Auth, error) {
	if config.GlobalConfig == nil {
		panic("Global config is nil!")
	}
	webAuthConf := webauthn.Config{}
	webAuthConf.RPDisplayName = config.GlobalConfig.WebAuth.DisplayName

	tmpUrl, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse root url %q: %w",
			"http://localhost:8080",
			err,
		)
	}
	webAuthConf.RPID = tmpUrl.Hostname()
	webAuthConf.RPOrigins = []string{tmpUrl.Scheme + tmpUrl.Host}
	webAuth, err := webauthn.New(&webAuthConf)
	if err != nil {
		return nil, fmt.Errorf("webAuth creation failed with config %#v: %w", webAuthConf, err)
	}
	hasher, err := passwd.New(passwd.Argon2idDefault)
	if err != nil {
		return nil, fmt.Errorf("failed to create password hasher: %w", err)
	}
	_ = hasher.SetKey([]byte(config.GlobalConfig.General.HashingSecret))
	a := Auth{
		store:              store,
		webAuth:            webAuth,
		hasher:             hasher,
		activeAuthRequests: map[string]TempAuthRequest{},
		log:                log.Logger,
		authMode:           mode,
		registerRequests: RegisterRequestHolder{
			Requests: map[string]RegisterProcess{},
		},
	}
	a.insertSuAccount()
	if mode == AUTH_MODE_DEV {
		a.insertDevAccount()
	}

	return &a, nil
}

func (a *Auth) insertDevAccount() {
	a.store.InsertDevAccount()
}

func (a *Auth) insertSuAccount() {
	// if !config.GlobalConfig.Superuser.Enabled {
	// 	return
	// }
	// acc := storage.Account{
	// 	Name: config.GlobalConfig.Superuser.Username,
	// }
	// if config.GlobalConfig.Superuser.PasswordIsRaw != nil &&
	// 	*config.GlobalConfig.Superuser.PasswordIsRaw {
	// 	hashed, err := a.hasher.Hash([]byte(config.GlobalConfig.Superuser.Password))
	// 	if err != nil {
	// 		panic("Failed to hash superuser password!")
	// 	}
	// 	acc.PasswordHash = hashed
	// }
	// a.store.InsertSudoAccount(acc.Name, acc.PasswordHash)
}
