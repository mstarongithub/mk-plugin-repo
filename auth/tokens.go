package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var ErrExpiredInPast = errors.New("expired date in the past")

// Generate a new token for an account. Should be unique and secure enough for the required purposes
// Token is bas64 encoded version of: hashed username + separator + timestamp in seconds + separator + a randomly selected plugin ID the account owns
func (a *Auth) generateToken(accID uint, expiresAt *time.Time) (string, error) {
	token := uuid.NewString()
	if expiresAt == nil {
		tmp := time.Now().Add(time.Hour * 24)
		expiresAt = &tmp
	}

	err := a.store.InsertNewToken(accID, token, *expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to insert new token into db: %w", err)
	}
	return token, nil
}

func (a *Auth) CheckToken(rawToken string) (uint, bool) {
	rawToken = strings.TrimPrefix(rawToken, "Bearer ")
	a.log.WithField("token", rawToken).Debugln("Checking token for validity")
	switch a.authMode {
	case AUTH_MODE_DEFAULT:
		a.log.Debugln("Running in prod mode, full token checks enabled")
	case AUTH_MODE_DEV:
		a.log.Debugln("Runing in dev mode, checking for dev token")
		if rawToken == _TOKEN_DEV_ACCOUNT {
			return 1, true
		}
	case AUTH_MODE_NONE:
		a.log.Debugln("Auth mode disabled, accepting all")
		return 1, true
	}
	dbToken, err := a.store.FindToken(rawToken)
	if err != nil {
		logrus.WithError(err).WithField("rawToken", rawToken).Infoln("token not found in db")
		return 0, false
	}
	if dbToken.ExpiresAt.Before(time.Now()) {
		logrus.WithFields(logrus.Fields{
			"token": rawToken,
		}).Debugln("Token expired")
		return 0, false
	}
	return dbToken.UserID, true
}

func (a *Auth) ExtendToken(tokenToExtend, confirmationToken string, extendTo time.Time) bool {
	if extendTo.Before(time.Now()) {
		a.log.Infoln("Can't extend token to a point in the past")
		return false
	}
	dbTokenToExtend, err := a.store.FindToken(tokenToExtend)
	if err != nil {
		a.log.WithField("token to extend", tokenToExtend).
			WithError(err).
			Warnln("Couldn't get token to extend from db")
		return false
	}

	if dbTokenToExtend.ExpiresAt.Before(time.Now()) {
		a.log.Infoln("Can't extend an already expired token")
		return false
	}

	dbCheckToken, err := a.store.FindToken(confirmationToken)
	if err != nil {
		a.log.WithField("check token", confirmationToken).
			WithError(err).
			Warnln("Couldn't get confirmation token for extension from db")
	}

	if dbCheckToken.ExpiresAt.Before(time.Now()) {
		a.log.Infoln("Confirmation token for extension already expired")
		return false
	}

	if dbTokenToExtend.UserID != dbCheckToken.UserID {
		a.log.Infoln("Can't extend token with token from different account")
	}

	if dbCheckToken.CreatedAt.Sub(time.Now()).Minutes() > 10.0 {
		a.log.Infoln("Check token is too old")
		return false
	}
	dbTokenToExtend.ExpiresAt = extendTo
	if err = a.store.ExtendToken(dbTokenToExtend); err != nil {
		a.log.WithError(err).Warnln("Couldn't update token in db")
		return false
	}

	return true
}
