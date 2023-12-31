package fsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
)

type DB struct {
	Path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps   map[int]Chirp     `json:"chirps"`
	Users    map[int]User      `json:"users"`
	Metadata map[string]string `json:"metadata"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

// NB: Only exported functions are ensured to be thread safe
func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, marshalErr := json.Marshal(dbStructure)
	if marshalErr != nil {
		return marshalErr
	}
	return os.WriteFile(db.Path, dat, 0666)
}

func (db *DB) loadDB() (DBStructure, error) {
	dbData, readErr := os.ReadFile(db.Path)
	if readErr != nil {
		return DBStructure{}, readErr
	}
	dbStructure := DBStructure{}
	unMarshalErr := json.Unmarshal(dbData, &dbStructure)
	if unMarshalErr != nil {
		return DBStructure{}, unMarshalErr
	}
	return dbStructure, nil
}

func (db *DB) ensureDB() error {
	_, statErr := os.Stat(db.Path)
	if errors.Is(statErr, os.ErrNotExist) {
		dbStructure := DBStructure{make(map[int]Chirp), make(map[int]User), map[string]string{"nextChirpId": "1", "nextUserId": "1"}}
		return db.writeDB(dbStructure)
	}
	return statErr
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return []Chirp{}, loadErr
	}
	allChirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		allChirps = append(allChirps, chirp)
	}
	sort.Slice(allChirps, func(i, j int) bool { return allChirps[i].Id < allChirps[j].Id })
	return allChirps, nil
}

func (db *DB) GetUniqueChirp(chirpId int) (Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return Chirp{}, loadErr
	}
	chirp, ok := dbStructure.Chirps[chirpId]
	if !ok {
		return Chirp{}, errors.New("Invalid chirp id")
	}
	return chirp, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return Chirp{}, loadErr
	}
	nextChirpId, atoiErr := strconv.Atoi(dbStructure.Metadata["nextChirpId"])
	if atoiErr != nil {
		return Chirp{}, atoiErr
	}
	newChirp := Chirp{Id: nextChirpId, Body: body}
	dbStructure.Chirps[nextChirpId] = newChirp
	dbStructure.Metadata["nextChirpId"] = fmt.Sprintf("%d", nextChirpId+1)
	writeErr := db.writeDB(dbStructure)
	return newChirp, writeErr
}

func (db *DB) CreateUser(email string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return User{}, loadErr
	}
	nextUserId, atoiErr := strconv.Atoi(dbStructure.Metadata["nextUserId"])
	if atoiErr != nil {
		return User{}, atoiErr
	}
	newUser := User{Id: nextUserId, Email: email}
	dbStructure.Users[nextUserId] = newUser
	dbStructure.Metadata["nextUserId"] = fmt.Sprintf("%d", nextUserId+1)
	writeErr := db.writeDB(dbStructure)
	return newUser, writeErr
}

func NewDB(path string) (*DB, error) {
	db := DB{Path: path, mu: &sync.RWMutex{}}
	return &db, db.ensureDB()
}
