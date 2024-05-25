package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

// NOTE: If the separator ever gets changed, the db check will fail
const _TOKEN_VAL_SEPARATOR = ";+;"
const _EMPTY_VAL = "nada"

var ErrExpiredInPast = errors.New("expired date in the past")

// Generate a new token for an account. Should be unique and secure enough for the required purposes
// Token is bas64 encoded version of: hashed username + separator + timestamp in seconds + separator + a randomly selected plugin ID the account owns
func (a *Auth) generateToken(accID uint, expiresAt *time.Time) (string, error) {
	acc, err := a.store.FindAccountByID(accID)
	if err != nil {
		return "", fmt.Errorf("failed to get account for %d: %w", accID, err)
	}
	usernameHash, err := a.hasher.Hash([]byte(acc.Name))
	if err != nil {
		return "", fmt.Errorf("failed to hash username: %w", err)
	}
	timeString := fmt.Sprint(time.Now().Unix())
	randomPluginID := ""
	if len(acc.PluginsOwned) == 0 {
		randomPluginID = _EMPTY_VAL
	} else {
		elemNr := rand.Intn(len(acc.PluginsOwned))
		randomPluginID = fmt.Sprint(acc.PluginsOwned[elemNr])
	}
	tokenString := string(
		base64.StdEncoding.EncodeToString(
			[]byte(
				string(usernameHash) +
					_TOKEN_VAL_SEPARATOR +
					timeString +
					_TOKEN_VAL_SEPARATOR +
					randomPluginID,
			),
		),
	)

	if expiresAt == nil {
		stamp := time.Now().Add(time.Hour * 24)
		expiresAt = &stamp
	}

	if expiresAt.After(time.Now()) {
		return "", ErrExpiredInPast
	}

	token := storage.Token{
		Token:     tokenString,
		UserID:    acc.ID,
		ExpiresAt: *expiresAt,
	}
	err = a.store.InsertNewToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to insert new token into db: %w", err)
	}
	return tokenString, nil
}

func (a *Auth) CheckToken(rawToken string) bool {
	dbToken, err := a.store.FindToken(rawToken)
	if err != nil {
		logrus.WithError(err).WithField("rawToken", rawToken).Infoln("token not found in db")
		return false
	}
	// First has to decode from base 64
	base64Res, err := base64.StdEncoding.DecodeString(rawToken)
	if err != nil {
		logrus.WithError(err).
			WithField("rawToken", rawToken).
			Infoln("Failed to decode token from base64")
		return false
	}
	from64 := string(base64Res)

	// Then check that there are exactly 3 parts after splitting on the separator
	elements := strings.Split(from64, _TOKEN_VAL_SEPARATOR)
	if len(elements) != 3 {
		logrus.WithField("decodedBase64", from64).
			Infoln("Bad amount of parts after splitting on separator")
		return false
	}

	return false
}
