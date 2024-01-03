package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateToken(issuer string, userId int, jwtSecret string, expirationDuration time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
		Subject:   fmt.Sprintf("%d", userId),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func generateAccessToken(userId int, jwtSecret string) (string, error) {
	return generateToken("chirpy-access", userId, jwtSecret, time.Hour)
}

func generateRefreshToken(userId int, jwtSecret string) (string, error) {
	return generateToken("chirpy-refresh", userId, jwtSecret, time.Duration(24*60)*time.Hour)
}

func validateAccessToken(accessToken, secret string) (*jwt.Token, error) {
	parsedToken, error := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if error != nil {
		return parsedToken, error
	}
	issuer, _ := parsedToken.Claims.GetIssuer()
	if issuer != "chirpy-access" {
		return parsedToken, errors.New("Unautorized: cannot use refresh token for access")
	}
	return parsedToken, nil
}

func validateRefreshToken(accessToken, secret string) (*jwt.Token, error) {
	parsedToken, error := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if error != nil {
		return parsedToken, error
	}
	issuer, _ := parsedToken.Claims.GetIssuer()
	if issuer != "chirpy-refresh" {
		return parsedToken, errors.New("Unautorized: cannot use access token for refresh")
	}
	return parsedToken, nil
}
