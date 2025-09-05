package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mamatb/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func respondOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func respondJsonClean(w http.ResponseWriter, _ *http.Request, cleanedBody string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanedBody,
	})
	w.Write(response)
}

func respondJsonError(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(400)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: message,
	})
	w.Write(response)
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

func cleanProfanities(body string, profanities map[string]bool) string {
	bodySlice := strings.Split(body, " ")
	for wordIdx, word := range bodySlice {
		if profanities[strings.ToLower(word)] {
			bodySlice[wordIdx] = "****"
		}
	}
	return strings.Join(bodySlice, " ")
}

func main() {
	godotenv.Load()
	mux := http.NewServeMux()
	db, _ := sql.Open("postgres", os.Getenv("DB_URL"))
	defer db.Close()
	config := apiConfig{
		dbQueries: database.New(db),
	}
	profanities := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc(
		"GET /api/healthz",
		respondOk,
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
		config.middlewareMetricsReset(respondOk),
	)

	mux.HandleFunc(
		"POST /api/validate_chirp",
		func(w http.ResponseWriter, r *http.Request) {
			request := struct {
				Body string `json:"body"`
			}{}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				respondJsonError(w, r, "Something went wrong")
			} else if len(request.Body) > 140 {
				respondJsonError(w, r, "Chirp is too long")
			} else {
				respondJsonClean(w, r, cleanProfanities(request.Body, profanities))
			}
		},
	)

	log.Fatal(server.ListenAndServe())
}
