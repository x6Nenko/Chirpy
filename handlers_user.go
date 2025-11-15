package main

import (
	"net/http"
	"encoding/json"
	"github.com/x6Nenko/Chirpy/internal/auth"
	"github.com/x6Nenko/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	// Step 1: Define what you expect to receive
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	// Step 2: Decode the request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Step 3: Hash pass
	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Step 4: Create user
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:    params.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	convertedUser := User{
    ID:        user.ID,
    CreatedAt: user.CreatedAt,
    UpdatedAt: user.UpdatedAt,
    Email:     user.Email,
	}

	respondWithJSON(w, 201, convertedUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Step 1: Define what you expect to receive
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	// Step 2: Decode the request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Step 3: Get user by email
	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password", err)
		return
	}

	// Step 4: Compare passwords
	ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !ok {
		respondWithError(w, 401, "Incorrect email or password", err)
		return
	}

	// Step 5: Logged in
	convertedUser := User{
    ID:        user.ID,
    CreatedAt: user.CreatedAt,
    UpdatedAt: user.UpdatedAt,
    Email:     user.Email,
	}

	respondWithJSON(w, 200, convertedUser)
}