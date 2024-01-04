package fsdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	Path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp        `json:"chirps"`
	Users         map[int]DBUser       `json:"users"`
	RevokedTokens map[string]time.Time `json:"revoked-tokens"`
	Metadata      map[string]string    `json:"metadata"`
}

type Chirp struct {
	AuthorId int    `json:"author_id"`
	Id       int    `json:"id"`
	Body     string `json:"body"`
}

type User struct {
	Id          int    `json:"id"`
	Email       string `json:"email"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

type DBUser struct {
	User
	Password string `json:"password"`
}

type ErrorMessage string

const (
	InvalidUserId     ErrorMessage = "Invalid user id"
	TokenRevoked      ErrorMessage = "Token revoked"
	IncorrectPassword ErrorMessage = "Incorrect password"
	UserNotExist      ErrorMessage = "User doesn't exist"
	Unauthorized      ErrorMessage = "Unauthorized"
	ResourceNotExist  ErrorMessage = "Resource doesn't exist"
)

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
		dbStructure := DBStructure{
			make(map[int]Chirp),
			make(map[int]DBUser),
			make(map[string]time.Time),
			map[string]string{"nextChirpId": "1", "nextUserId": "1"},
		}
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

func (db *DB) CreateChirp(body string, createdById int) (Chirp, error) {
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
	newChirp := Chirp{AuthorId: createdById, Id: nextChirpId, Body: body}
	dbStructure.Chirps[nextChirpId] = newChirp
	dbStructure.Metadata["nextChirpId"] = fmt.Sprintf("%d", nextChirpId+1)
	writeErr := db.writeDB(dbStructure)
	return newChirp, writeErr
}

func (db *DB) DeleteChirp(chirpId, userId int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return loadErr
	}
	chirp, ok := dbStructure.Chirps[chirpId]
	if !ok {
		return errors.New(string(ResourceNotExist))
	}
	if chirp.AuthorId != userId {
		return errors.New(string(Unauthorized))
	}
	delete(dbStructure.Chirps, chirpId)
	return db.writeDB(dbStructure)
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return User{}, loadErr
	}
	for _, val := range dbStructure.Users {
		if val.Email == email {
			return User{}, errors.New("Unique email constraint")
		}
	}
	nextUserId, atoiErr := strconv.Atoi(dbStructure.Metadata["nextUserId"])
	if atoiErr != nil {
		return User{}, atoiErr
	}
	newUser := DBUser{User: User{nextUserId, email, false}, Password: password}
	dbStructure.Users[nextUserId] = newUser
	dbStructure.Metadata["nextUserId"] = fmt.Sprintf("%d", nextUserId+1)
	writeErr := db.writeDB(dbStructure)
	return newUser.User, writeErr
}

func (db *DB) AuthenticateUser(email string, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return User{}, loadErr
	}
	for _, val := range dbStructure.Users {
		if val.Email == email {
			if bcrypt.CompareHashAndPassword([]byte(val.Password), []byte(password)) == nil {
				return val.User, nil
			}
			return User{}, errors.New(string(IncorrectPassword))
		}
	}
	return User{}, errors.New(string(UserNotExist))
}

func (db *DB) GetUser(userId int) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return User{}, loadErr
	}
	user, ok := dbStructure.Users[userId]
	if !ok {
		return User{}, errors.New(string(InvalidUserId))
	}
	return user.User, nil
}

func (db *DB) UpdateUser(userId int, newEmail, newPassword string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return User{}, loadErr
	}
	user, ok := dbStructure.Users[userId]
	if !ok {
		return User{}, errors.New(string(InvalidUserId))
	}
	user.Email = newEmail
	user.Password = newPassword
	dbStructure.Users[userId] = user
	writeErr := db.writeDB(dbStructure)
	return user.User, writeErr
}

func (db *DB) UpgradeUser(userId int) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return loadErr
	}
	user, ok := dbStructure.Users[userId]
	if !ok {
		return errors.New(string(InvalidUserId))
	}
	user.IsChirpyRed = true
	dbStructure.Users[userId] = user
	return db.writeDB(dbStructure)
}

func (db *DB) RevokeToken(token string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return loadErr
	}
	dbStructure.RevokedTokens[token] = time.Now()
	return db.writeDB(dbStructure)
}

func (db *DB) IsTokenRevoked(token string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	dbStructure, loadErr := db.loadDB()
	if loadErr != nil {
		return loadErr
	}
	_, ok := dbStructure.RevokedTokens[token]
	if ok {
		return errors.New(string(TokenRevoked))
	}
	return nil
}

func NewDB(path string) (*DB, error) {
	db := DB{Path: path, mu: &sync.RWMutex{}}
	return &db, db.ensureDB()
}
