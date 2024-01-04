package main

import (
	"encoding/json"
	"fsdb"
	"net/http"
	"strings"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	reqBody := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	decoderErr := decoder.Decode(&reqBody)
	if decoderErr != nil {
		respondWithError(w, 500, decoderErr.Error())
		return
	}
	user, validateErr := cfg.db.AuthenticateUser(reqBody.Email, reqBody.Password)
	if validateErr == nil {
		signedAccessToken, accessErr := generateAccessToken(user.Id, cfg.jwtSecret)
		if accessErr != nil {
			respondWithError(w, 500, accessErr.Error())
			return
		}
		signedRefreshToken, refreshErr := generateRefreshToken(user.Id, cfg.jwtSecret)
		if refreshErr != nil {
			respondWithError(w, 500, accessErr.Error())
			return
		}
		type UserWithTokens struct {
			fsdb.User
			AccessToken  string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}
		respondWithJSON(w, 200, UserWithTokens{user, signedAccessToken, signedRefreshToken})
	} else if validateErr.Error() == string(fsdb.IncorrectPassword) {
		respondWithError(w, 401, "Password didn't match")
	} else if validateErr.Error() == string(fsdb.UserNotExist) {
		respondWithError(w, 401, "No user with that email")
	} else {
		respondWithError(w, 500, validateErr.Error())
	}
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) < 2 || authHeader[0] != "Bearer" {
		respondWithError(w, 401, "Missing authorization")
		return
	}
	refreshToken := authHeader[1]
	token, validationErr := cfg.validateToken(refreshToken, string(TokenTypeRefresh))
	if validationErr != nil || !token.Valid {
		respondWithError(w, 401, validationErr.Error())
		return
	}
	userId, idErr := getUserId(token)
	if idErr != nil {
		respondWithError(w, 500, idErr.Error())
	}
	accesToken, tokenErr := generateAccessToken(userId, cfg.jwtSecret)
	if tokenErr != nil {
		respondWithJSON(w, 500, tokenErr.Error())
	}
	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{accesToken})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) < 2 || authHeader[0] != "Bearer" {
		respondWithError(w, 401, "Missing authorization")
		return
	}
	refreshToken := authHeader[1]
	token, validationErr := cfg.validateToken(refreshToken, string(TokenTypeRefresh))
	if validationErr != nil || !token.Valid {
		respondWithError(w, 401, validationErr.Error())
		return
	}
	revokeErr := cfg.db.RevokeToken(refreshToken)
	if revokeErr != nil {
		respondWithError(w, 500, revokeErr.Error())
	}
	w.WriteHeader(200)
}
