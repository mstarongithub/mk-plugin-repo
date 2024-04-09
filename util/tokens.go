package util

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const TOKEN_ISSUER = "mk-plugin-repo-api"

var secretKey = []byte("some super secret key that no one will ever guess")

var ErrInvalidToken = fmt.Errorf("invalid token")

func CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS512,
		jwt.MapClaims{
			"iss": TOKEN_ISSUER,
			"sub": username,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		},
	)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("token signing failed: %w", err)
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to parse string as token: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"token.issuer":   resToString(token.Claims.GetIssuer()),
		"token.subject":  resToString(token.Claims.GetSubject()),
		"token.audience": resToString(token.Claims.GetAudience()),
	}).Debugln("Verifying token")

	if !token.Valid {
		return false, ErrInvalidToken
	}
	return true, nil
}

func resToString(x any, e error) string {
	if e != nil {
		return fmt.Sprintf("%v", e)
	} else {
		return fmt.Sprintf("%v", x)
	}
}
