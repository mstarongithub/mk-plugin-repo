package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var ErrExpiredInPast = errors.New("expired date in the past")

// Generate a new token for an account. Should be unique and secure enough for the required purposes
// Token is bas64 encoded version of: hashed username + separator + timestamp in seconds + separator + a randomly selected plugin ID the account owns
func (a *Auth) generateToken(accID uint, expiresAt *time.Time) (string, error) {
	if expiresAt == nil {
		tmp := time.Now().Add(time.Hour * 24)
		expiresAt = &tmp
	}

	token, err := a.store.NewToken(accID, *expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to insert new token into db: %w", err)
	}
	return token, nil
}

func (a *Auth) CheckToken(rawToken string) (uint, bool) {
	rawToken = strings.TrimPrefix(rawToken, "Bearer ")
	a.log.Debug().Str("token", rawToken).Msg("Checking token for validity")
	switch a.authMode {
	case AUTH_MODE_DEFAULT:
		a.log.Debug().Msg("Running in prod mode, full token checks enabled")
	case AUTH_MODE_DEV:
		a.log.Debug().Msg("Runing in dev mode, checking for dev token")
		if rawToken == _TOKEN_DEV_ACCOUNT {
			return 1, true
		}
	case AUTH_MODE_NONE:
		a.log.Debug().Msg("Auth mode disabled, accepting all")
		return 1, true
	}
	dbToken, err := a.store.FindToken(rawToken)
	if err != nil {
		log.Info().Str("token", rawToken).Err(err).Msg("Token not found in db")
		return 0, false
	}
	if dbToken.ExpiresAt.Before(time.Now()) {
		log.Info().Str("token", rawToken).Msg("Token expired")
		return 0, false
	}
	return dbToken.UserID, true
}

func (a *Auth) ExtendToken(tokenToExtend, confirmationToken string, extendTo time.Time) bool {
	if extendTo.Before(time.Now()) {
		log.Warn().Msg("Can't extend token expiration date to a point in the past")
		return false
	}
	dbTokenToExtend, err := a.store.FindToken(tokenToExtend)
	if err != nil {
		log.Error().Err(err).Str("token", tokenToExtend).Msg("Couldn't get token to extend from db")
		return false
	}

	if dbTokenToExtend.ExpiresAt.Before(time.Now()) {
		log.Info().Msg("Cant extend already expired token")
		return false
	}

	dbCheckToken, err := a.store.FindToken(confirmationToken)
	if err != nil {
		log.Error().
			Err(err).
			Str("token-confirmation", confirmationToken).
			Msg("Couldn't get confirmation token from db")
		return false
	}

	if dbCheckToken.ExpiresAt.Before(time.Now()) {
		log.Warn().Msg("Attempt to extend another token using an expired one as confirmation")
		return false
	}

	if dbTokenToExtend.UserID != dbCheckToken.UserID {
		log.Warn().Msg("Mismatched tokens for token extension")
	}

	dbTokenToExtend.ExpiresAt = extendTo
	if err = a.store.ExtendToken(dbTokenToExtend); err != nil {
		log.Error().
			Err(err).
			Str("token", tokenToExtend).
			Str("confirmation-token", confirmationToken).
			Msg("Couldn't update token expiration date in db")
		return false
	}

	log.Info().Msg("Token expiration date extended")
	return true
}
