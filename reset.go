package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Reset is only allowed in dev mode", nil)
		return
	}

	cfg.fileserverHits.Store(0)
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong when trying to delete the users", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
