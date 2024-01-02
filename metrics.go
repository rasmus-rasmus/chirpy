package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) serveHitCountMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	responseBody := fmt.Sprintf(
		`<html>

			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>

		</html>`,
		cfg.fileServerHits,
	)
	w.Write([]byte(responseBody))
}

func (cfg *apiConfig) resetHitCountMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
