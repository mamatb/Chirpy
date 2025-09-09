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
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mamatb/Chirpy/internal/database"
)

type apiConfig struct {
	platform       string
	dbQueries      *database.Queries
	fileserverHits atomic.Int32
}

type errorJson struct {
	Error string `json:"error"`
}

type userJson struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type chirpJson struct {
	ID        uuid.UUID     `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Body      string        `json:"body"`
	UserID    uuid.NullUUID `json:"user_id"`
}

func respPlainOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func respPlainForbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("FORBIDDEN"))
}

func respJsonError(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(400)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(errorJson{
		Error: message,
	})
	w.Write(response)
}

func respJsonCreatedUser(w http.ResponseWriter, _ *http.Request, user database.User) {
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(userJson{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
	w.Write(response)
}

func respJsonCreatedChirp(w http.ResponseWriter, _ *http.Request, chirp database.Chirp) {
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(chirpJson{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	w.Write(response)
}

func (c *apiConfig) middleMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) middleMetricsReset(next http.HandlerFunc) http.HandlerFunc {
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
	db, _ := sql.Open("postgres", os.Getenv("DB_URL"))
	defer db.Close()
	mux, config := http.NewServeMux(), apiConfig{
		platform:  os.Getenv("PLATFORM"),
		dbQueries: database.New(db),
	}
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	profanities := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	mux.HandleFunc(
		"GET /api/healthz",
		respPlainOk,
	)

	mux.Handle(
		"/app/",
		config.middleMetricsInc(http.StripPrefix(
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
		func(w http.ResponseWriter, r *http.Request) {
			if config.platform != "dev" {
				respPlainForbidden(w, r)
			} else {
				config.dbQueries.DeleteUsers(r.Context())
				config.middleMetricsReset(respPlainOk)
			}
		},
	)

	mux.HandleFunc(
		"POST /api/users",
		func(w http.ResponseWriter, r *http.Request) {
			request := struct {
				Email string `json:"email"`
			}{}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				respJsonError(w, r, "Something went wrong")
			} else {
				user, _ := config.dbQueries.CreateUser(r.Context(), request.Email)
				respJsonCreatedUser(w, r, user)
			}
		},
	)

	mux.HandleFunc(
		"POST /api/chirps",
		func(w http.ResponseWriter, r *http.Request) {
			request := struct {
				Body   string        `json:"body"`
				UserID uuid.NullUUID `json:"user_id"`
			}{}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				respJsonError(w, r, "Something went wrong")
			} else if len(request.Body) > 140 {
				respJsonError(w, r, "Chirp is too long")
			} else {
				chirp, _ := config.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
					Body:   cleanProfanities(request.Body, profanities),
					UserID: request.UserID,
				})
				respJsonCreatedChirp(w, r, chirp)
			}
		},
	)

	log.Fatal(server.ListenAndServe())
}
