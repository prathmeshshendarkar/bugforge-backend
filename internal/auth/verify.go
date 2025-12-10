package auth

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// Returns (userID, customerID, error)
func VerifyTokenAndGetUserID(tokenString string) (string, string, error) {
	if tokenString == "" {
		return "", "", errors.New("missing token")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("jwt secret not configured")
	}

	claims := &JWTClaims{}

	// Parse + verify
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return "", "", err
	}

	if !token.Valid {
		return "", "", errors.New("invalid token")
	}

	if claims.UserID == "" {
		return "", "", errors.New("missing user id in token")
	}

	return claims.UserID, claims.CustomerID, nil
}
