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
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps   map[int]Chirp     `json:"chirps"`
	Metadata map[string]string `json:"metadata"`
}

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

// NB: Only exported functions are ensured to be thread safe
func (db *DB) writeDB(dbStructure DBStructure) error {
	dat, marshalErr := json.Marshal(dbStructure)
	if marshalErr != nil {
		return marshalErr
	}
	return os.WriteFile(db.path, dat, 0666)
}

func (db *DB) loadDB() (DBStructure, error) {
	dbData, readErr := os.ReadFile(db.path)
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
	_, statErr := os.Stat(db.path)
	if errors.Is(statErr, os.ErrNotExist) {
		dbStructure := DBStructure{make(map[int]Chirp), map[string]string{"nextId": "1"}}
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

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return Chirp{}, loadErr
	}
	nextId, atoiErr := strconv.Atoi(dbStructure.Metadata["nextId"])
	if atoiErr != nil {
		return Chirp{}, atoiErr
	}
	newChirp := Chirp{Id: nextId, Body: body}
	dbStructure.Chirps[nextId] = newChirp
	dbStructure.Metadata["nextId"] = fmt.Sprintf("%d", nextId+1)
	writeErr := db.writeDB(dbStructure)
	return newChirp, writeErr
}

func NewDB(path string) (*DB, error) {
	db := DB{path: path, mu: &sync.RWMutex{}}
	return &db, db.ensureDB()
}
