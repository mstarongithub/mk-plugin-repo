package auth

import (
	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mstarongithub/mk-plugin-repo/storage"
)

type AuthManager struct {
	storage     *storage.Storage
	validTokens [][]byte
}

func NewAuthManager(store *storage.Storage) (*AuthManager, error) {}

func (am *AuthManager) Login(username, password string) (bool, string, error) {
	acc, err := am.storage.FindAccountByName(username)
	if err != nil {
		if errors.Is(err, storage.ErrAccountNotFound) {
			return false, "", nil
		} else {
			return false, "", err
		}
	}
	match, _ := argon2id.ComparePasswordAndHash(password, string(acc.PasswordHash))
	if match {

	}
	return match, "", nil
}
