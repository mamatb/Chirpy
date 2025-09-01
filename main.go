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

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) middlewareMetricsReset(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Store(0)
		next.ServeHTTP(w, r)
	}
}

func handlerOK(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	mux, config := http.NewServeMux(), apiConfig{}
	mux.HandleFunc("GET /api/healthz", handlerOK)
	mux.Handle("/app/", config.middlewareMetricsInc(http.StripPrefix(
		"/app/",
		http.FileServer(http.Dir("."))),
	))
	mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Hits: "))
		w.Write([]byte(strconv.Itoa((int)(config.fileserverHits.Load()))))
	})
	mux.HandleFunc("POST /api/reset", config.middlewareMetricsReset(handlerOK))
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
