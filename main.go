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

const (
	HeaderContentType = "Content-Type"
	ContentTypePlain  = "text/plain; charset=utf-8"
	ContentTypeHtml   = "text/html; charset=utf-8"
	ContentTypeJson   = "application/json; charset=utf-8"
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
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("OK")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainForbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("FORBIDDEN")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respPlainNotFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(404)
	w.Header().Set(HeaderContentType, ContentTypePlain)
	body := []byte("NOT FOUND")
	if _, err := w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonError(w http.ResponseWriter, _ *http.Request, message string) {
	w.WriteHeader(400)
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(errorJson{
		Error: message,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonCreatedUser(w http.ResponseWriter, _ *http.Request, user database.User) {
	w.WriteHeader(201)
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(userJson{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonCreatedChirp(w http.ResponseWriter, r *http.Request, chirp database.Chirp) {
	w.WriteHeader(201)
	respJsonChirp(w, r, chirp)
}

func respJsonChirps(w http.ResponseWriter, _ *http.Request, chirps []database.Chirp) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	var chirpsJson []chirpJson
	for _, chirp := range chirps {
		chirpsJson = append(chirpsJson, chirpJson{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	if body, err = json.Marshal(chirpsJson); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
}

func respJsonChirp(w http.ResponseWriter, _ *http.Request, chirp database.Chirp) {
	w.Header().Set(HeaderContentType, ContentTypeJson)
	var err error
	var body []byte
	if body, err = json.Marshal(chirpJson{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}); err != nil {
		log.Fatal(err)
	}
	if _, err = w.Write(body); err != nil {
		log.Fatal(err)
	}
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
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	mux, config := http.NewServeMux(), apiConfig{platform: os.Getenv("PLATFORM")}
	if db, err := sql.Open("postgres", os.Getenv("DB_URL")); err != nil {
		log.Fatal(err)
	} else {
		defer db.Close()
		config.dbQueries = database.New(db)
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
		"GET /api/health",
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
			w.Header().Set(HeaderContentType, ContentTypeHtml)
			if _, err := w.Write([]byte(fmt.Sprintf(""+
				"<html>\n"+
				"  <body>\n"+
				"    <h1>Welcome, Chirpy Admin</h1>\n"+
				"    <p>Chirpy has been visited %d times!</p>\n"+
				"  </body>\n"+
				"</html>\n",
				config.fileserverHits.Load(),
			))); err != nil {
				log.Fatal(err)
			}
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

	mux.HandleFunc(
		"GET /api/chirps",
		func(w http.ResponseWriter, r *http.Request) {
			chirps, _ := config.dbQueries.GetChirps(r.Context())
			respJsonChirps(w, r, chirps)
		},
	)

	mux.HandleFunc(
		"GET /api/chirps/{chirpID}",
		func(w http.ResponseWriter, r *http.Request) {
			var chirp database.Chirp
			if chirpID, err := uuid.Parse(r.PathValue("chirpID")); err != nil {
				log.Fatal(err)
			} else {
				chirp, _ = config.dbQueries.GetChirp(r.Context(), chirpID)
			}
			if chirp.ID == uuid.Nil {
				respPlainNotFound(w, r)
			} else {
				respJsonChirp(w, r, chirp)
			}
		},
	)

	log.Fatal(server.ListenAndServe())
}
