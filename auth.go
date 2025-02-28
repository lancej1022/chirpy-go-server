package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	// TODO: `parameters` and `returnVals` are duplicated here and in `users.go` -- should probably share a single definition
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type returnVals struct {
		Id        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
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

	respondWithJSON(w, http.StatusOK, returnVals{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
