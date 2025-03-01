package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
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
		UserID: userId, // Use the userId from the JWT
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
	authorId := r.URL.Query().Get("author_id")

	var getChirpsErr error
	var chirps []database.Chirp

	if authorId != "" {
		authorId, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
		}

		chirps, getChirpsErr = cfg.db.GetChirpsByUserId(r.Context(), authorId)
	} else {
		chirps, getChirpsErr = cfg.db.GetChirps(r.Context())
	}

	if getChirpsErr != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when getting chirps", getChirpsErr)
		return
	}

	if len(chirps) == 0 {
		respondWithJSON(w, http.StatusOK, []Chirp{})
		return
	}

	chirpsResponse := make([]Chirp, len(chirps))
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
func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "Missing chirp ID", nil)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token", err)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Could not find chirp with id: %s", chirpId), err)
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when deleting chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
