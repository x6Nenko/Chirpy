package main

import (
	"net/http"
	"encoding/json"
	"errors"
	"database/sql"
	"github.com/x6Nenko/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
    Event string `json:"event"`
    Data  struct {
			UserID uuid.UUID `json:"user_id"`
    } `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.dbQueries.UpdateUserChirpyRed(r.Context(), database.UpdateUserChirpyRedParams{
		IsChirpyRed:  true,
		ID:						params.Data.UserID,
	})
	if err != nil {
		// Is this a "not found" error or a real problem?
    if errors.Is(err, sql.ErrNoRows) {
			// This specific error means "nothing found"
			w.WriteHeader(http.StatusNotFound)
			return
    }
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user chirpy red status", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}