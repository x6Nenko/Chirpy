package main

import (
	"log"
	"net/http"
	"fmt"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	// Creating a new ServeMux
	ServeMux := http.NewServeMux()

	// Creating a server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}

	fs := http.FileServer(http.Dir("."))
	apiCfg := &apiConfig{}
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))
	ServeMux.HandleFunc("/healthz", handlerReadiness)
	ServeMux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	ServeMux.HandleFunc("/reset", apiCfg.handlerReset)

	// Start the server
	log.Printf("Serving on port: 8080\n")
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()
	fmt.Fprintf(w, "Hits: %d", count)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}