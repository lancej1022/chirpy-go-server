package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	filepathRoot := "."
	port := "8080"
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}
	// ServeMux is an HTTP request multiplexer.
	// It matches the URL of each incoming request against a list of registered patterns
	// and calls the handler for the pattern that most closely matches the URL.
	mux := http.NewServeMux()
	// If a pattern ends with a slash /, it matches all URL paths that have the same prefix.
	// For example, a pattern /images/ matches /images/, /images/logo.png, and /images/css/style.css.
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/validate_chirp", apiCfg.handleChirp)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
