module main

go 1.21.4

replace fsdb v0.0.0 => ./fsdb

require (
	fsdb v0.0.0
	github.com/go-chi/chi/v5 v5.0.11
)

require golang.org/x/crypto v0.17.0

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/joho/godotenv v1.5.1
)
