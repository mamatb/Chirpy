package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

type validateChirpRequest struct {
	Body string `json:"body"`
}

type validateChirpResponseOk struct {
	Valid bool `json:"valid"`
}

type validateChirpResponseError struct {
	Error string `json:"error"`
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

func handlerOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func main() {
	mux, config := http.NewServeMux(), apiConfig{}
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc(
		"GET /api/healthz",
		handlerOk,
	)

	mux.Handle(
		"/app/",
		config.middlewareMetricsInc(http.StripPrefix(
			"/app/",
			http.FileServer(http.Dir(".")),
		)),
	)

	mux.HandleFunc(
		"GET /admin/metrics",
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(fmt.Sprintf(""+
				"<html>\n"+
				"  <body>\n"+
				"    <h1>Welcome, Chirpy Admin</h1>\n"+
				"    <p>Chirpy has been visited %d times!</p>\n"+
				"  </body>\n"+
				"</html>\n",
				config.fileserverHits.Load(),
			)))
		},
	)

	mux.HandleFunc(
		"POST /admin/reset",
		config.middlewareMetricsReset(handlerOk),
	)

	mux.HandleFunc(
		"POST /api/validate_chirp",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			chirpReq := validateChirpRequest{}
			if json.NewDecoder(r.Body).Decode(&chirpReq) != nil {
				w.WriteHeader(400)
				chirpRespFail, _ := json.Marshal(validateChirpResponseError{
					Error: "Something went wrong",
				})
				w.Write(chirpRespFail)
				return
			}
			if len(chirpReq.Body) > 140 {
				w.WriteHeader(400)
				chirpRespFail, _ := json.Marshal(validateChirpResponseError{
					Error: "Chirp is too long",
				})
				w.Write(chirpRespFail)
				return
			}
			chirpRespSuccess, _ := json.Marshal(validateChirpResponseOk{
				Valid: true,
			})
			w.Write(chirpRespSuccess)
		},
	)

	log.Fatal(server.ListenAndServe())
}
