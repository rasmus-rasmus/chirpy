package main

import (
	"encoding/json"
	"errors"
	"fsdb"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func makeChirpsPostHandler(db *fsdb.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		reqBody := struct {
			Body string `json:"body"`
		}{}
		decoderErr := decoder.Decode(&reqBody)
		if decoderErr != nil {
			respondWithError(w, 500, decoderErr.Error())
			return
		}

		cleanBody, validationErr := validateChirp(reqBody.Body)
		if validationErr != nil {
			respondWithError(w, 400, validationErr.Error())
			return
		}
		chirp, createErr := db.CreateChirp(cleanBody)
		if createErr != nil {
			respondWithError(w, 500, createErr.Error())
		}
		respondWithJSON(w, 201, chirp)
	}
}

func makeChirpsGetHandler(db *fsdb.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		chirps, getErr := db.GetChirps()
		if getErr != nil {
			respondWithError(w, 500, getErr.Error())
			return
		}
		respondWithJSON(w, 200, chirps)
	}
}

func makeChirpsGetUniqueHandler(db *fsdb.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		chirpId, atoiErr := strconv.Atoi(chi.URLParam(r, "chirpId"))
		if atoiErr != nil {
			respondWithError(w, 500, atoiErr.Error())
		}
		chirp, getErr := db.GetUniqueChirp(chirpId)
		if getErr != nil {
			if getErr.Error() == "Invalid chirp id" {
				respondWithError(w, 404, "Chirp does not exist")
				return
			}
			respondWithError(w, 500, getErr.Error())
			return
		}
		respondWithJSON(w, 200, chirp)
	}
}

func validateChirp(chirpBody string) (string, error) {
	if len(chirpBody) > 140 {
		return "", errors.New("Chirp is too long")
	}
	return cleanUpChirp(chirpBody), nil
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := struct {
		Msg string `json:"error"`
	}{message}
	respondWithJSON(w, statusCode, errorResponse)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func cleanUpChirp(profaneChirp string) string {
	theProfane := []string{"kerfuffle", "sharbert", "fornax"}
	wordList := strings.Split(profaneChirp, " ")
	for i, word := range wordList {
		if slices.Contains(theProfane, strings.ToLower(word)) {
			wordList[i] = "****"
		}
	}
	return strings.Join(wordList, " ")
}
