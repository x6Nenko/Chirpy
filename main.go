package main

import (
	"log"
	"net/http"
	"fmt"
	"sync/atomic"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"time"
	"os"
	"database/sql"
	"github.com/x6Nenko/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries  		 *database.Queries
	platform 			 string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platformEnv := os.Getenv("PLATFORM")
	if platformEnv == "" {
		log.Fatal("PLATFORM must be set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	queries := database.New(dbConn)
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries: 			queries,
		platform:				platformEnv,
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
	ServeMux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	ServeMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGetOne)
	ServeMux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsGetAll)
	ServeMux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	ServeMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
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
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.fileserverHits.Store(0)

	err := cfg.dbQueries.Reset(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}