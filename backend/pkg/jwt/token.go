package jwt

import (
	"fmt"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       uint64 `json:"userId"`
	Role         string `json:"role"`
	TokenType    string `json:"tokenType,omitempty"`
	TokenVersion int    `json:"tokenVersion,omitempty"`
	golangjwt.RegisteredClaims
}

const (
	TokenTypeAccess     = "ACCESS"
	TokenTypeReactivate = "REACTIVATE"
)

func Generate(secret string, expireHours int, userID uint64, role string, tokenVersion int) (string, error) {
	if expireHours <= 0 {
		expireHours = 168
	}
	return generate(secret, time.Duration(expireHours)*time.Hour, userID, role, TokenTypeAccess, tokenVersion)
}

func GenerateReactivation(secret string, userID uint64, tokenVersion int) (string, error) {
	return generate(secret, 5*time.Minute, userID, "", TokenTypeReactivate, tokenVersion)
}

func generate(secret string, duration time.Duration, userID uint64, role, tokenType string, tokenVersion int) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:       userID,
		Role:         role,
		TokenType:    tokenType,
		TokenVersion: tokenVersion,
		RegisteredClaims: golangjwt.RegisteredClaims{
			ExpiresAt: golangjwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  golangjwt.NewNumericDate(now),
		},
	}

	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt token: %w", err)
	}
	return signed, nil
}

func Parse(secret string, tokenString string) (*Claims, error) {
	token, err := golangjwt.ParseWithClaims(tokenString, &Claims{}, func(token *golangjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid jwt token")
	}
	return claims, nil
}
