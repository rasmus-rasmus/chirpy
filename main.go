package main

import (
	"fsdb"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func startServer(port string) {
	cfg := apiConfig{0}

	mainRouter := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mainRouter.Handle("/app/*", cfg.middlewareMetricsIncrementer(fileHandler))
	mainRouter.Handle("/app", cfg.middlewareMetricsIncrementer(fileHandler))

	adminRouter.Get("/metrics", cfg.serveHitCountMetrics)

	db, dbErr := fsdb.NewDB("./database.json")
	if dbErr != nil {
		log.Fatal("Could not open database connection", dbErr.Error())
	}
	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.HandleFunc("/reset", cfg.resetHitCountMetrics)
	apiRouter.Post("/chirps", makeChirpsPostHandler(db))
	apiRouter.Get("/chirps", makeChirpsGetHandler(db))

	mainRouter.Mount("/api/", apiRouter)
	mainRouter.Mount("/admin/", adminRouter)

	corsRouter := middlewareCors(mainRouter)

	srv := &http.Server{Addr: ":" + port, Handler: corsRouter}

	log.Printf("Starting server on port: %s", port)
	log.Fatal(srv.ListenAndServe())
}

func main() {
	startServer("8080")
}
