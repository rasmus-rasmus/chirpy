package main

import (
	"encoding/json"
	"errors"
	"fsdb"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (cfg *apiConfig) chirpsPostHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) < 2 || authHeader[0] != "Bearer" {
		respondWithError(w, 401, "Missing authorization")
		return
	}
	accessToken := authHeader[1]
	parsedToken, validationErr := cfg.validateToken(accessToken, string(TokenTypeAccess))
	if validationErr != nil {
		respondWithError(w, 401, validationErr.Error())
		return
	}
	userId, idErr := getUserId(parsedToken)
	if idErr != nil {
		respondWithError(w, 500, idErr.Error())
		return
	}
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
	chirp, createErr := cfg.db.CreateChirp(cleanBody, userId)
	if createErr != nil {
		respondWithError(w, 500, createErr.Error())
	}
	respondWithJSON(w, 201, chirp)
}

func (cfg *apiConfig) chirpsGetHandler(w http.ResponseWriter, r *http.Request) {
	authorIdParam := r.URL.Query().Get("author_id")
	var chirps []fsdb.Chirp
	var getErr error
	if len(authorIdParam) == 0 {
		chirps, getErr = cfg.db.GetChirps()
	} else {
		authorId, convErr := strconv.Atoi(authorIdParam)
		if convErr != nil {
			respondWithError(w, 500, convErr.Error())
			return
		}
		chirps, getErr = cfg.db.GetChirpsFromAuthor(authorId)
	}
	if getErr != nil {
		respondWithError(w, 500, getErr.Error())
		return
	}
	sort := r.URL.Query().Get("sort")
	if sort == "desc" {
		slices.Reverse(chirps)
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

func (cfg *apiConfig) chirpsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) < 2 || authHeader[0] != "Bearer" {
		respondWithError(w, 401, "Missing authorization")
		return
	}
	accessToken := authHeader[1]
	parsedToken, validationErr := cfg.validateToken(accessToken, string(TokenTypeAccess))
	if validationErr != nil {
		respondWithError(w, 401, validationErr.Error())
		return
	}
	userId, idErr := getUserId(parsedToken)
	if idErr != nil {
		respondWithError(w, 500, idErr.Error())
		return
	}
	chirpId, atoiErr := strconv.Atoi(chi.URLParam(r, "chirpId"))
	if atoiErr != nil {
		respondWithError(w, 500, atoiErr.Error())
	}
	deleteErr := cfg.db.DeleteChirp(chirpId, userId)
	if deleteErr != nil {
		if deleteErr.Error() == string(fsdb.Unauthorized) {
			respondWithError(w, 403, deleteErr.Error())
			return
		}
		respondWithError(w, 500, deleteErr.Error())
		return
	}
	w.WriteHeader(200)
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
