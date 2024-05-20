package auth

import (
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/ermites-io/passwd"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"

	"github.com/mstarongithub/mk-plugin-repo/config"
	"github.com/mstarongithub/mk-plugin-repo/storage"
	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

type NextAuthState uint

const (
	AUTH_SUCCESS = NextAuthState(0)
	AUTH_FAIL    = NextAuthState(1 << iota)
	AUTH_NEEDS_FIDO
	AUTH_NEEDS_TOTP
	AUTH_NEEDS_MAIL
)

type Auth struct {
	store              *storage.Storage
	webAuth            *webauthn.WebAuthn
	hasher             *passwd.Profile
	activeAuthRequests map[string]TempAuthRequest
}

type TempAuthRequest struct {
	AuthID    string
	AccountID uint // NOTE: Could replace this with a reference to the actual account struct later if db access times become a problem
	NextState NextAuthState
}

// Create a new authentication manager
// Requires a reference to a storage implementation
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
		store:              store,
		webAuth:            webAuth,
		hasher:             hasher,
		activeAuthRequests: map[string]TempAuthRequest{},
	}, nil
}

// TODO: Implement these two

// Start a login attempt using a passkey
// NOTE: Not implemented yet signature will also change
func (a *Auth) LoginWithPasskeyStart() {}

// Complete a login attempt with a passkey
// Requires that attempt to be started by a call to LoginWithPasskeyStart
// NOTE: Not implemented yet, signature will also change
func (a *Auth) LoginWithPasskeyComplete() {}

// Attempt a login using a username and password
// Tries to prevent timing attacks at least a little
// Returns the next state (a set of flags, see the AUTH_ constants, 0 == ok) and a string containing the process ID if mfa is required
func (a *Auth) LoginWithPassword(username, password string) (NextAuthState, string) {
	time.Sleep(
		(time.Millisecond * time.Duration(rand.Uint32())) % 250,
	) // Sleep a random amount of time between 0 and 250ms. Fuck those timing attacks

	acc, err := a.store.FindAccountByName(username)
	if err != nil {
		logrus.WithError(err).
			WithField("username", username).
			Infoln("Couldn't find account for login request")
		return AUTH_FAIL, ""
	}

	if acc.AuthMethods == customtypes.AUTH_METHOD_NONE {
		return AUTH_SUCCESS, ""
	}
	if !customtypes.AuthIsFlagSet(acc.AuthMethods, customtypes.AUTH_METHOD_PASSWORD) {
		return AUTH_FAIL, ""
	}

	if a.hasher.Compare(acc.PasswordHash, []byte(password)) != nil {
		return AUTH_FAIL, ""
	}

	retFlag := AUTH_SUCCESS // Because this is the 0 state
	if customtypes.AuthIsFlagSet(acc.AuthMethods, customtypes.AUTH_METHOD_FIDO) {
		retFlag = retFlag | AUTH_NEEDS_FIDO
	}
	if customtypes.AuthIsFlagSet(acc.AuthMethods, customtypes.AUTH_METHOD_TOTP) {
		retFlag = retFlag | AUTH_NEEDS_TOTP
	}
	if retFlag == 0 && customtypes.AuthIsFlagSet(acc.AuthMethods, customtypes.AUTH_METHOD_MAIL) {
		// TODO: Send mail with code here
		retFlag = retFlag | AUTH_NEEDS_MAIL
	}

	if retFlag == AUTH_SUCCESS {
		return AUTH_SUCCESS, ""
	}

	requestID := username + fmt.Sprint(time.Now().Unix())
	a.activeAuthRequests[requestID] = TempAuthRequest{
		AuthID:    requestID,
		AccountID: acc.ID,
		NextState: retFlag,
	}
	return retFlag, requestID
}

// Continue a login process started via a username + password combo
// Takes the type of mfa as well as a token to check
// Returns the next state (a set of flags, see the AUTH_ constants, 0 == ok) and a string containing the process ID if the process is not complete yet
func (a *Auth) LoginWithMFA(
	processID string,
	token string,
	mfaType NextAuthState,
) (NextAuthState, string) {
	process, ok := a.activeAuthRequests[processID]
	if !ok {
		return AUTH_FAIL, ""
	}
	if !customtypes.AuthIsFlagSet(
		customtypes.AuthMethods(process.NextState),
		customtypes.AuthMethods(mfaType),
	) {
		return AUTH_FAIL, ""
	}
	acc, _ := a.store.FindAccountByID(process.AccountID)

	switch mfaType {
	case AUTH_NEEDS_FIDO:
		// TODO: Implement this
		panic("MFA Fido not implemented yet")
	case AUTH_NEEDS_TOTP:
		if !totp.Validate(token, *acc.TotpToken) {
			return AUTH_FAIL, ""
		}
	case AUTH_NEEDS_MAIL:
		// TODO: Implement this, this'll be pain
		panic("MFA Mail not implemented yet")
	}

	process.NextState = process.NextState &^ mfaType // Disable completed mfa flag. Since 0 is the ok, all is ok
	a.activeAuthRequests[processID] = process

	return process.NextState, processID
}