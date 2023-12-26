package main

import (
	"log"
	"net/http"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func startServer(port string) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	corsMux := middlewareCors(mux)

	srv := http.Server{Addr: "localhost:" + port, Handler: corsMux}

	log.Printf("Starting server on port: %s", port)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	startServer("8080")
}
