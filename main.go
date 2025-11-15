package main

import (
	"log"
	"net/http"
	"fmt"
	"sync/atomic"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"database/sql"
	"github.com/x6Nenko/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries  *database.Queries
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	queries := database.New(dbConn)
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries: 			queries,
	}

	// Creating a new ServeMux
	ServeMux := http.NewServeMux()

	// Creating a server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: ServeMux,
	}

	fs := http.FileServer(http.Dir("."))
	ServeMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))
	ServeMux.HandleFunc("GET /api/healthz", handlerReadiness)
	ServeMux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidateChirp)
	ServeMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	ServeMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	count := cfg.fileserverHits.Load()
	template := `<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`
	fmt.Fprintf(w, template, count)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}