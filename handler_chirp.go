package main

import (
	"net/http"
	"encoding/json"
	"github.com/x6Nenko/Chirpy/internal/database"
	"github.com/x6Nenko/Chirpy/internal/auth"
	"github.com/google/uuid"
	"time"
	"sort"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body 	 string 	 `json:"body"`
		// UserId uuid.UUID `json:"user_id"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Couldn't get bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	validatedChirp := replaceBadWords(params.Body)

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
    Body:   validatedChirp,
    UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	convertedChirp := Chirp{
    ID:        chirp.ID,
    CreatedAt: chirp.CreatedAt,
    UpdatedAt: chirp.UpdatedAt,
    Body:      chirp.Body,
		UserID:		 userID,
	}

	respondWithJSON(w, 201, convertedChirp)
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	authorIDString := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")

	var dbChirps []database.Chirp
	var err error

	if authorIDString != "" {
		// authorID was provided as query parameter
		// Parse a UUID string
		authorID, parseErr := uuid.Parse(authorIDString)
		if parseErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't parse UUID string", parseErr)
			return
		}

		dbChirps, err = cfg.dbQueries.GetAllChirpsByAuthor(r.Context(), authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get all chirps by author", err)
			return
		}
	} else {
		dbChirps, err = cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't get all chirps", err)
			return
		}
	}

	// Convert database chirps to API chirps
	convertedChirps := []Chirp{}
	for _, chirp := range dbChirps {
		convertedChirps = append(convertedChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	// Apply sorting if requested
	if sortOrder == "desc" {
		sort.Slice(convertedChirps, func(i, j int) bool {
			return convertedChirps[i].CreatedAt.After(convertedChirps[j].CreatedAt)
		})
	}

	respondWithJSON(w, 200, convertedChirps)
}

func (cfg *apiConfig) handlerChirpsGetOne(w http.ResponseWriter, r *http.Request) {
	chirpIdString := r.PathValue("chirpID") // String literal matches {chirpID} from route

	// Parse a UUID string
	chirpID, err := uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse UUID string", err)
		return
	}

	chirp, err := cfg.dbQueries.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Couldn't get chirp", err)
		return
	}

	convertedChirp := Chirp{
    ID:        chirp.ID,
    CreatedAt: chirp.CreatedAt,
    UpdatedAt: chirp.UpdatedAt,
    Body:      chirp.Body,
		UserID:		 chirp.UserID,
	}

	respondWithJSON(w, 200, convertedChirp)
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	chirpIdString := r.PathValue("chirpID") // String literal matches {chirpID} from route

	// Parse a UUID string
	chirpID, err := uuid.Parse(chirpIdString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse UUID string", err)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Couldn't get bearer token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	chirp, err := cfg.dbQueries.GetOneChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "Couldn't get chirp", err)
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, 403, "Unauthorized", err)
		return
	}

	err = cfg.dbQueries.DeleteOneChirp(r.Context(), database.DeleteOneChirpParams{
		ID:    			chirpID,
		UserID: 		userID,
	})
	if err != nil {
		respondWithError(w, 403, "Unauthorized", err)
		return
	}

	w.WriteHeader(204)
	return
}