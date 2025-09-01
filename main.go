package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux, apiCfg := http.NewServeMux(), apiConfig{fileserverHits: atomic.Int32{}}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK"))
	})
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix(
		"/app/",
		http.FileServer(http.Dir(".")),
	)))
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Hits: "))
		w.Write([]byte(strconv.Itoa((int)(apiCfg.fileserverHits.Load()))))
	})
	mux.Handle("/reset", apiCfg.middlewareMetricsReset(http.StripPrefix(
		"/reset",
		http.FileServer(http.Dir(".")),
	)))
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
