package main

import (
	"encoding/json"
	"fsdb"
	"net/http"
)

func makeCreateUserHandler(db *fsdb.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		reqBody := struct {
			Email string `json:"email"`
		}{}
		decoderErr := decoder.Decode(&reqBody)
		if decoderErr != nil {
			respondWithError(w, 500, decoderErr.Error())
			return
		}
		user, createErr := db.CreateUser(reqBody.Email)
		if createErr != nil {
			respondWithError(w, 500, createErr.Error())
			return
		}
		respondWithJSON(w, 201, user)
	}
}
