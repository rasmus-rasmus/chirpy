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
)

func startServer(port string, debug bool, dbPathChan chan string) {
	cfg := apiConfig{0}

	mainRouter := chi.NewRouter()
	apiRouter := chi.NewRouter()
	adminRouter := chi.NewRouter()

	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mainRouter.Handle("/app/*", cfg.middlewareMetricsIncrementer(fileHandler))
	mainRouter.Handle("/app", cfg.middlewareMetricsIncrementer(fileHandler))

	adminRouter.Get("/metrics", cfg.serveHitCountMetrics)

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

	apiRouter.Get("/healthz", readinessHandler)
	apiRouter.HandleFunc("/reset", cfg.resetHitCountMetrics)
	apiRouter.Post("/chirps", makeChirpsPostHandler(db))
	apiRouter.Get("/chirps", makeChirpsGetHandler(db))
	apiRouter.Get("/chirps/{chirpId}", makeChirpsGetUniqueHandler(db))
	apiRouter.Post("/users", makeCreateUserHandler(db))

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
