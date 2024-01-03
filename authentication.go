package main

import (
	"encoding/json"
	"fsdb"
	"net/http"
	"strconv"
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
	} else if validateErr.Error() == "Incorrect password" {
		respondWithError(w, 401, "Password didn't match")
	} else if validateErr.Error() == "User doesn't exist" {
		respondWithError(w, 401, "No user with that email")
	} else {
		respondWithError(w, 500, validateErr.Error())
	}
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, validationErr := validateRefreshToken(refreshToken, cfg.jwtSecret)
	if validationErr != nil || !token.Valid {
		respondWithError(w, 401, validationErr.Error())
		return
	}
	isRevoked, revokeErr := cfg.db.IsTokenRevoked(refreshToken)
	if revokeErr != nil {
		respondWithError(w, 500, revokeErr.Error())
		return
	}
	if isRevoked {
		respondWithError(w, 401, "Token revoked")
		return
	}
	stringifiedUserId, subjectErr := token.Claims.GetSubject()
	if subjectErr != nil {
		respondWithJSON(w, 500, subjectErr.Error())
		return
	}
	userId, atoiErr := strconv.Atoi(stringifiedUserId)
	if atoiErr != nil {
		respondWithJSON(w, 500, atoiErr.Error())
		return
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
	refreshToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, validationErr := validateRefreshToken(refreshToken, cfg.jwtSecret)
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
