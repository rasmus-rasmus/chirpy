package main

import (
	"bufio"
	"flag"
	"fmt"
	"fsdb"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileServerHits int
	jwtSecret      string
	polkaApiKey    string
	db             *fsdb.DB
}

func startServer(port string, debug bool, dbPathChan chan string) {
	// Connect to database, load environment variables from .env-file
	// and set up api config
	var db *fsdb.DB
	var dbErr error
	if debug {
		db, dbErr = fsdb.NewDB("./database.debug.json")
	} else {
		db, dbErr = fsdb.NewDB("./database.json")
	}
	if dbErr != nil {
		log.Fatal("Could not open database connection", dbErr.Error())
	}
	godotenv.Load()
	cfg := apiConfig{
		fileServerHits: 0,
		jwtSecret:      os.Getenv("JWT_SECRET"),
		polkaApiKey:    os.Getenv("POLKA_API_KEY"),
		db:             db,
	}

	mainRouter := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mainRouter.Handle("/app/*", cfg.middlewareMetricsIncrementer(fileHandler))
	mainRouter.Handle("/app", cfg.middlewareMetricsIncrementer(fileHandler))

	adminRouter.Get("/metrics", cfg.serveHitCountMetrics)

	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.HandleFunc("/reset", cfg.resetHitCountMetrics)

	apiRouter.Post("/chirps", cfg.chirpsPostHandler)
	apiRouter.Get("/chirps", cfg.chirpsGetHandler)
	apiRouter.Get("/chirps/{chirpId}", cfg.chirpsGetUniqueHandler)
	apiRouter.Delete("/chirps/{chirpId}", cfg.chirpsDeleteHandler)

	apiRouter.Post("/users", cfg.createUserHandler)
	apiRouter.Put("/users", cfg.updateUserHandler)
	apiRouter.Post("/login", cfg.loginHandler)

	apiRouter.Post("/refresh", cfg.refreshHandler)
	apiRouter.Post("/revoke", cfg.revokeHandler)

	apiRouter.Post("/polka/webhooks", cfg.polkaWebhookHandler)

	mainRouter.Mount("/api/", apiRouter)
	mainRouter.Mount("/admin/", adminRouter)

	corsRouter := middlewareCors(mainRouter)

	srv := &http.Server{Addr: ":" + port, Handler: corsRouter}

	log.Printf("Starting server on port: %s", port)
	dbPathChan <- db.Path
	log.Fatal(srv.ListenAndServe())
}

func main() {
	debugFlg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	dbPathChan := make(chan string)
	go startServer("8080", *debugFlg, dbPathChan)

	dbPath := <-dbPathChan
	time.Sleep(20 * time.Millisecond)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Type 'exit' to shut down the server > ")
		scanner.Scan()
		command := scanner.Text()
		if command == "exit" {
			if *debugFlg {
				fmt.Printf("Deleting test database: %s\n", dbPath)
				os.Remove(dbPath)
			} else {
				fmt.Printf("Database: %s persists\n", dbPath)
			}
			break
		}
	}
}
