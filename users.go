package main

import (
	"encoding/json"
	"fsdb"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
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
	passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		respondWithError(w, 500, hashErr.Error())
	}
	user, createErr := cfg.db.CreateUser(reqBody.Email, string(passwordHash))
	if createErr != nil {
		if createErr.Error() == "Unique email constraint" {
			respondWithError(w, 409, "Email already in use")
			return
		}
		respondWithError(w, 500, createErr.Error())
		return
	}
	respondWithJSON(w, 201, user)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	requestToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, validationErr := validateToken(requestToken, cfg.jwtSecret)
	if validationErr != nil || !token.Valid {
		respondWithError(w, 401, validationErr.Error())
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
	passwordHash, hashErr := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
	if hashErr != nil {
		respondWithError(w, 500, hashErr.Error())
		return
	}
	user, updateErr := cfg.db.UpdateUser(userId, reqBody.Email, string(passwordHash))
	if updateErr != nil {
		respondWithError(w, 500, updateErr.Error())
		return
	}
	respondWithJSON(w, 200, user)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	reqBody := struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}{}
	decoderErr := decoder.Decode(&reqBody)
	if decoderErr != nil {
		respondWithError(w, 500, decoderErr.Error())
		return
	}
	user, validateErr := cfg.db.ValidateUser(reqBody.Email, reqBody.Password)
	if validateErr == nil {
		var expirationTime int
		if reqBody.ExpiresInSeconds <= 0 || reqBody.ExpiresInSeconds > 24*60*60 {
			expirationTime = 24 * 60 * 60
		} else {
			expirationTime = reqBody.ExpiresInSeconds
		}
		signedToken, tokenErr := generateToken(user.Id, expirationTime, cfg.jwtSecret)
		if tokenErr != nil {
			respondWithError(w, 500, tokenErr.Error())
			return
		}
		type UserWithToken struct {
			fsdb.User
			Token string `json:"token"`
		}
		respondWithJSON(w, 200, UserWithToken{user, signedToken})
	} else if validateErr.Error() == "Incorrect password" {
		respondWithError(w, 401, "Password didn't match")
	} else if validateErr.Error() == "User doesn't exist" {
		respondWithError(w, 401, "No user with that email")
	} else {
		respondWithError(w, 500, validateErr.Error())
	}
}
