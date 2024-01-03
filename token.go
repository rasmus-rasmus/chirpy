package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
	// TokenTypeRefresh -
	TokenTypeRefresh TokenType = "chirpy-refresh"
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
	return generateToken(string(TokenTypeAccess), userId, jwtSecret, time.Hour)
}

func generateRefreshToken(userId int, jwtSecret string) (string, error) {
	return generateToken(string(TokenTypeRefresh), userId, jwtSecret, time.Duration(24*60)*time.Hour)
}

func (cfg *apiConfig) validateToken(token, expectedIssuer string) (*jwt.Token, error) {
	parsedToken, error := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(cfg.jwtSecret), nil })
	if error != nil {
		return parsedToken, error
	}
	issuer, _ := parsedToken.Claims.GetIssuer()
	if issuer != expectedIssuer {
		return parsedToken, errors.New("Unauthorized issuer")
	}
	if isIssuerRevokable(issuer) {
		revokeErr := cfg.db.IsTokenRevoked(token)
		return parsedToken, revokeErr
	}
	return parsedToken, nil
}

func isIssuerRevokable(issuer string) bool {
	return issuer == string(TokenTypeRefresh)
}

// func validateAccessToken(accessToken, secret string) (*jwt.Token, error) {
// 	parsedToken, error := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
// 	if error != nil {
// 		return parsedToken, error
// 	}
// 	issuer, _ := parsedToken.Claims.GetIssuer()
// 	if issuer != "chirpy-access" {
// 		return parsedToken, errors.New("Unauthorized issuer")
// 	}
// 	return parsedToken, nil
// }

// func validateRefreshToken(accessToken, secret string) (*jwt.Token, error) {
// 	parsedToken, error := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) { return []byte(secret), nil })
// 	if error != nil {
// 		return parsedToken, error
// 	}
// 	issuer, _ := parsedToken.Claims.GetIssuer()
// 	if issuer != "chirpy-refresh" {
// 		return parsedToken, errors.New("Unauthorized issuer")
// 	}
// 	return parsedToken, nil
// }

func getUserId(token *jwt.Token) (int, error) {
	stringifiedUserId, subjectErr := token.Claims.GetSubject()
	if subjectErr != nil {
		return 0, subjectErr
	}
	userId, atoiErr := strconv.Atoi(stringifiedUserId)
	if atoiErr != nil {
		return 0, atoiErr
	}
	return userId, nil
}
