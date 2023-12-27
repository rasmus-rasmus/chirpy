package main

import (
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

	apiRouter.Get("/healthz/", readinessHandler)
	apiRouter.Get("/healthz", readinessHandler)

	adminRouter.Get("/metrics/", cfg.serveHitCountMetrics)
	adminRouter.Get("/metrics", cfg.serveHitCountMetrics)

	apiRouter.HandleFunc("/reset/", cfg.resetHitCountMetrics)
	apiRouter.HandleFunc("/reset", cfg.resetHitCountMetrics)

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
