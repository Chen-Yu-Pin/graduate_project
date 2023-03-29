package util

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type MyClaims struct {
	Account string `json:"user-account"`
	jwt.StandardClaims
}

var SecretKey = []byte("********")

func GenToken(account string) (string, error) {
	c := MyClaims{
		account,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * 60 * time.Minute).Unix(),
			Issuer:    "LoginToken",
		},
	}
	// Choose specific algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// Choose specific Signature
	return token.SignedString(SecretKey)
}

func ParseToken(tokenString string) error {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return SecretKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return errors.New("token expired")
			}
		}

		return err
	}
	// Valid token
	if _, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return nil
	}

	return err
}
