package main

import (
	"encoding/json"
	"fsdb"
	"net/http"
	"strings"
)

type EventType string

const (
	EventTypeUserUpgrade EventType = "user.upgraded"
)

func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := strings.Split(r.Header.Get("Authorization"), " ")
	if len(authHeader) < 2 || authHeader[0] != "ApiKey" {
		respondWithError(w, 401, "Missing authorization")
		return
	}
	apiKey := authHeader[1]
	if apiKey != cfg.polkaApiKey {
		respondWithError(w, 401, "Authorization failed")
		return
	}
	decoder := json.NewDecoder(r.Body)
	reqBody := struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}{}
	decoderErr := decoder.Decode(&reqBody)
	if decoderErr != nil {
		respondWithError(w, 500, decoderErr.Error())
		return
	}
	if reqBody.Event != string(EventTypeUserUpgrade) {
		w.WriteHeader(200)
		return
	}
	upgradeErr := cfg.db.UpgradeUser(reqBody.Data.UserId)
	if upgradeErr != nil {
		if upgradeErr.Error() == string(fsdb.UserNotExist) {
			respondWithError(w, 404, "User doesn't exist")
			return
		}
		respondWithError(w, 500, upgradeErr.Error())
		return
	}
	w.WriteHeader(200)
}
