package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handleChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when decoding request", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned := removeProfanity(params.Body)
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.UserId, // TODO: is this right...?
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when creating chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
	})
}

func removeProfanity(input string) string {
	bannedWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	split := strings.Fields(input)
	for i, word := range split {
		if bannedWords[strings.ToLower(word)] {
			split[i] = "****"
		}
	}
	return strings.Join(split, " ")
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when getting chirps", err)
		return
	}

	chirpsResponse := []Chirp{}
	for i, chirp := range chirps {
		chirpsResponse[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
		}
	}

	respondWithJSON(w, http.StatusOK, chirpsResponse)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID", nil)
		return
	}
	chirp, err := cfg.db.GetChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Could not find chirp with id: %s", chirpId), err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
	})
}
