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

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func respondOk(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func respondForbidden(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(403)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("FORBIDDEN"))
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

func respondJsonClean(w http.ResponseWriter, _ *http.Request, cleanedBody string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := json.Marshal(struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanedBody,
	})
	w.Write(response)
}

func (c *apiConfig) respondJsonCreated(w http.ResponseWriter, r *http.Request,
	email string) {
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	user, _ := c.dbQueries.CreateUser(r.Context(), email)
	response, _ := json.Marshal(User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
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
		respondOk,
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
				respondForbidden(w, r)
			} else {
				config.dbQueries.DeleteUsers(r.Context())
				config.middleMetricsReset(respondOk)
			}
		},
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

	mux.HandleFunc(
		"POST /api/users",
		func(w http.ResponseWriter, r *http.Request) {
			request := struct {
				Email string `json:"email"`
			}{}
			if json.NewDecoder(r.Body).Decode(&request) != nil {
				respondJsonError(w, r, "Something went wrong")
			} else {
				config.respondJsonCreated(w, r, request.Email)
			}
		},
	)

	log.Fatal(server.ListenAndServe())
}
