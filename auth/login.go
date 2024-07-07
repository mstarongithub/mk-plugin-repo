package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/mstarongithub/mk-plugin-repo/config"
	customtypes "github.com/mstarongithub/mk-plugin-repo/storage/customTypes"
)

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
// If it only needs username-password and is ok, returns AUTH_SUCCESS and an access token valid for 24h
func (a *Auth) LoginWithPassword(username, password string) (NextAuthState, string) {
	if a.authMode == AUTH_MODE_NONE {
		return AUTH_SUCCESS, _TOKEN_AUTH_MODE_NONE
	}
	time.Sleep(
		(time.Millisecond * time.Duration(rand.Uint32())) % 250,
	) // Sleep a random amount of time between 0 and 250ms. Fuck those timing attacks

	if username == config.GlobalConfig.Superuser.Username {
		if config.GlobalConfig.Superuser.PasswordIsRaw != nil &&
			*config.GlobalConfig.Superuser.PasswordIsRaw {
			if password != config.GlobalConfig.Superuser.Password {
				return AUTH_FAIL, ""
			}
		} else {
			if a.hasher.Compare(
				[]byte(config.GlobalConfig.Superuser.Password),
				[]byte(password),
			) != nil {
				return AUTH_FAIL, ""
			}
		}
		token, err := a.generateToken(0, nil)
		if err != nil {
			return AUTH_FAIL, ""
		}
		return AUTH_SUCCESS, token
	}

	acc, err := a.store.FindAccountByName(username)
	if err != nil {
		a.log.WithError(err).
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
		expireTime := time.Now().Add(time.Hour * 24)
		token, err := a.generateToken(acc.ID, &expireTime)
		if err != nil {
			return AUTH_FAIL, ""
		}
		return AUTH_SUCCESS, token
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
// If next state is ok, returns AUTH_SUCCESS and token expiring after 24h
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

	if process.NextState == AUTH_SUCCESS {
		delete(a.activeAuthRequests, processID)
		expires := time.Now().Add(time.Hour * 24)
		token, err := a.generateToken(acc.ID, &expires)
		if err != nil {
			// Failed to generate token, undo auth action for retry
			return process.NextState & mfaType, processID
		}
		return AUTH_SUCCESS, token
	}
	a.activeAuthRequests[processID] = process

	return process.NextState, processID
}
