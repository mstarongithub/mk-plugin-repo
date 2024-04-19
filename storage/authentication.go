package storage

import (
	_ "slices"

	_ "github.com/sirupsen/logrus"

	_ "github.com/mstarongithub/mk-plugin-repo/util"
)

// Authenticate a user using a username and password.
// Returns an access token and true on success, empty string and false otherwise
// func (storage *Storage) Authenticate(username, password string) (string, bool) {
// 	user := Account{}
// 	result := storage.db.First(&user, "name = ?", username)
// 	// Check if one account was found
// 	if result.RowsAffected != 1 {
// 		return "", false
// 	}
// 	pwHash, salt, err := util.Hash(password, user.Salt)
// 	if err != nil {
// 		logrus.WithError(err).
// 			WithField("password-raw", password).
// 			WithField("username", username).
// 			Warnln("failed to hash password")
// 		return "", false
// 	}
// 	if !slices.Equal(salt, user.Salt) {
// 		return "", false
// 	}
// 	if slices.Equal(pwHash, user.PasswordHash) {
// 		token, err := util.CreateToken(username)
// 		if err != nil {
// 			return "", false
// 		}
// 		return token, true
// 	}
// 	return "", false
// }
