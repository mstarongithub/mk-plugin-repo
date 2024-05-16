package auth

import (
	"fmt"
	"net/url"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sirupsen/logrus"

	"github.com/ermites-io/passwd"
	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type tempAuthState uint

type Auth struct {
	store               *storage.Storage
	webAuth             *webauthn.WebAuthn
	currentAuthRequests map[string]tempAuthRequest
	hasher              *passwd.Profile
}

type tempAuthRequest struct {
	ID     string
	State  tempAuthState
	UserID uint
}

const (
	_AUTH_STATE_DONE           = tempAuthState(0)
	_AUTH_STATE_NEEDS_MFA_FIDO = tempAuthState(1 << iota)
	_AUTH_STAT_NEEDS_MFA_APP_TOKEN
)

func NewAuth(store *storage.Storage) (*Auth, error) {
	if config.GlobalConfig == nil {
		panic("Global config is nil!")
	}
	webAuthConf := webauthn.Config{}
	webAuthConf.RPDisplayName = config.GlobalConfig.WebAuth.DisplayName
	tmpUrl, err := url.Parse(config.GlobalConfig.General.RootUrl)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse root url %q: %w",
			config.GlobalConfig.General.RootUrl,
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
	hasher.SetKey([]byte(config.GlobalConfig.General.HashingSecret))

	return &Auth{
		store:   store,
		webAuth: webAuth,
		hasher:  hasher,
	}, nil
}

// Returns an ID of a temporary login process and whether the initial attempt could even be found
// So if username and password match, this function returns the ID of a login process and true
func (a *Auth) LoginStartWithPassword(username, password string) (string, bool) {
	acc, err := a.store.FindAccountByName(username)
	if err != nil {
		logrus.WithError(err).
			WithField("username", username).
			Info("Error while trying to start a login request with username and password")
		return "", false
	}

	if a.hasher.Compare(acc.PasswordHash, []byte(password)) != nil {
		logrus.
			WithField("username", username).
			Infoln("Bad password while authenticating user via password")
		return "", false
	}

	return "", false
}

func (a *Auth) LoginContinueWithPassword(tempID string) {}

func (a *Auth) LoginStartWithPasskey()    {}
func (a *Auth) LoginCompleteWithPasskey() {}

func (a *Auth) RegisterStartWithPasskey()    {}
func (a *Auth) RegisterCompleteWithPasskey() {}

func (a *Auth) GetUserToken(uId uint) (string, error) {
	return "", nil
}

// Check if a token is valid for authentication. Returns username and result
func (a *Auth) VerifyToken(token string) (string, bool) {
	return "", false
}
