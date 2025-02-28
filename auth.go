package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}
	type returnVals struct {
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
		Id        uuid.UUID `json:"id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong decoding the response", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}

	if ok := auth.CheckPasswordHash(params.Password, user.HashedPassword); ok != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password.", err)
		return
	}

	// Set default expiration time to 1 hour
	expirationInSeconds := 60 * 60 // 1 hour in seconds

	// If expires_in_seconds is provided, use it (with a cap at 1 hour)
	if params.ExpiresInSeconds > 0 {
		if params.ExpiresInSeconds > expirationInSeconds {
			// Cap at 1 hour
			params.ExpiresInSeconds = expirationInSeconds
		}
		expirationInSeconds = params.ExpiresInSeconds
	}

	// TODO: do we actually need `time.Duration(expirationInSeconds)*time.Second`
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(expirationInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token:     token,
	})
}
