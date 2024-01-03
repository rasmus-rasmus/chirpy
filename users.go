package main

import (
	"encoding/json"
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
	accessToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	token, validationErr := cfg.validateToken(accessToken, "chirpy-access")
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
