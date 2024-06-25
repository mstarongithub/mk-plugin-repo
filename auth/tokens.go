package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var ErrExpiredInPast = errors.New("expired date in the past")

// Generate a new token for an account. Should be unique and secure enough for the required purposes
// Token is bas64 encoded version of: hashed username + separator + timestamp in seconds + separator + a randomly selected plugin ID the account owns
func (a *Auth) generateToken(accID uint, expiresAt *time.Time) (string, error) {
	token := uuid.NewString()

	err := a.store.InsertNewToken(accID, token, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to insert new token into db: %w", err)
	}
	return token, nil
}

func (a *Auth) CheckToken(rawToken string) bool {
	dbToken, err := a.store.FindToken(rawToken)
	if err != nil {
		logrus.WithError(err).WithField("rawToken", rawToken).Infoln("token not found in db")
		return false
	}
	if dbToken.ExpiresAt != nil && dbToken.ExpiresAt.After(time.Now()) {
		return false
	}
	return true
}
