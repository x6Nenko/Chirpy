package main

import (
	"net/http"
	"encoding/json"
	"time"
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
    ID:          user.ID,
    CreatedAt:   user.CreatedAt,
    UpdatedAt:   user.UpdatedAt,
    Email:       user.Email,
    IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(w, 201, convertedUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Step 1: Define what you expect to receive
	type parameters struct {
		Email 						string `json:"email"`
		Password 					string `json:"password"`
		ExpiresInSeconds  *int 	 `json:"expires_in_seconds"` // pointer = optional param
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	// Step 5: Create JWT token
	jwtToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate JWT token", err)
		return
	}

	// Step 6: Create and Save refresh token
	now := time.Now()
	expiration := now.Add(60 * 24 * time.Hour)  // 60 days from now

	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate Refresh token", err)
		return
	}

	savedRefreshTokenEntry, err := cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:    	refreshTokenString,
		UserID: 		user.ID,
		ExpiresAt:  expiration,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save Refresh token", err)
		return
	}

	// Step 7: Logged in
	convertedUser := response{
    User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
    },
    Token: 				jwtToken,
		RefreshToken: savedRefreshTokenEntry.Token,
	}

	respondWithJSON(w, 200, convertedUser)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
    Token string `json:"token"`
	}

	// 1. Get refresh token from headers
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Couldn't get bearer token", err)
		return
	}

	// 2. Check if there is such token in the DB
	dbRefreshToken, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "Couldn't get Refresh token", err)
		return
	}

	// 3. Check if the token is valid
	now := time.Now()
	if now.After(dbRefreshToken.ExpiresAt) {
    // Token is expired  
		respondWithError(w, 401, "Refresh token is expired", err)
		return
	}

	if dbRefreshToken.RevokedAt.Valid {
    // revoked_at is NOT NULL (token is revoked)
		respondWithError(w, 401, "Refresh token is expired", err)
		return
	}

	// 4. Create new JWT token
	jwtToken, err := auth.MakeJWT(dbRefreshToken.UserID , cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate JWT token", err)
		return
	}

	convertedResponse := response{
    Token: jwtToken,
	}

	respondWithJSON(w, 200, convertedResponse)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	// 1. Get refresh token from headers
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Couldn't get bearer token", err)
		return
	}

	// 2. Revoke refresh token
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)  // 204
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	// Step 1: Define what you expect to receive
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	// Step 2. Get auth token from headers
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Couldn't get bearer token", err)
		return
	}

	// Step 3. Check if token is valid
	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized", err)
		return
	}

	// Step 4: Decode the request body
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Step 5: Hash pass
	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Step 6: Update user
	user, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:    			params.Email,
		HashedPassword: hashedPass,
		ID:					userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	convertedUser := User{
    ID:          user.ID,
    CreatedAt:   user.CreatedAt,
    UpdatedAt:   user.UpdatedAt,
    Email:       user.Email,
    IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(w, 200, convertedUser)
}