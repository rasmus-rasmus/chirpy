package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) chirpsPostHandler(w http.ResponseWriter, r *http.Request) {
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
	chirp, createErr := cfg.db.CreateChirp(cleanBody)
	if createErr != nil {
		respondWithError(w, 500, createErr.Error())
	}
	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) chirpsGetHandler(w http.ResponseWriter, r *http.Request) {
	chirps, getErr := cfg.db.GetChirps()
	if getErr != nil {
		respondWithError(w, 500, getErr.Error())
		return
	}
	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) chirpsGetUniqueHandler(w http.ResponseWriter, r *http.Request) {
	chirpId, atoiErr := strconv.Atoi(chi.URLParam(r, "chirpId"))
	if atoiErr != nil {
		respondWithError(w, 500, atoiErr.Error())
	}
	chirp, getErr := cfg.db.GetUniqueChirp(chirpId)
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

func validateChirp(chirpBody string) (string, error) {
	if len(chirpBody) > 140 {
		return "", errors.New("Chirp is too long")
	}
	return cleanUpChirp(chirpBody), nil
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
