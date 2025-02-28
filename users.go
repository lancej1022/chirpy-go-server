package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleAddUser(w http.ResponseWriter, r *http.Request) {
	// TODO: should this be shared with `auth.go`?
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type returnVals struct {
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Id        uuid.UUID `json:"id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong decoding the response", err)
		return
	}

	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to generate password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when creating the user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		Id:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
