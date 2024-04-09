package util

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const HASH_TIME uint32 = 1
const HASH_MEMORY uint32 = 64 * 1024
const HASH_THREADS uint8 = 4
const HASH_KEY_LENGTH uint32 = 32
const HASH_SALT_LENGTH uint32 = 24

func Hash(toHash string, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		var err error
		salt, err = generateSalt(HASH_SALT_LENGTH)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"salt (length %d) generation failed: %w",
				HASH_SALT_LENGTH,
				err,
			)
		}
	}

	key := argon2.IDKey([]byte(toHash), salt, HASH_TIME, HASH_MEMORY, HASH_THREADS, HASH_KEY_LENGTH)
	return key, salt, nil
}

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to read random salt: %w", err)
	}
	return salt, nil
}
