package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleAddUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
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
