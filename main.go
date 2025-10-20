package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mamatb/Chirpy/database"
	web2 "github.com/mamatb/Chirpy/web"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	config := web2.ApiConfig{
		Platform: os.Getenv("PLATFORM"),
		PolkaKey: os.Getenv("POLKA_KEY"),
		Secret:   os.Getenv("SECRET"),
	}
	if db, err := sql.Open("postgres", os.Getenv("DB_URL")); err != nil {
		log.Fatal(err)
	} else {
		defer db.Close()
		config.DBQueries = database.New(db)
	}
	profanities := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	mux.HandleFunc(
		"GET /api/health",
		web2.HandlerGetApiHealth(),
	)
	mux.Handle(
		"/app/",
		web2.HandlerApp(&config),
	)
	mux.HandleFunc(
		"GET /admin/metrics",
		web2.HandlerGetAdminMetrics(&config),
	)
	mux.HandleFunc(
		"POST /admin/reset",
		web2.HandlerPostAdminReset(&config),
	)
	mux.HandleFunc(
		"POST /api/users",
		web2.HandlerPostApiUsers(&config),
	)
	mux.HandleFunc(
		"PUT /api/users",
		web2.HandlerPutApiUsers(&config),
	)
	mux.HandleFunc(
		"POST /api/login",
		web2.HandlerPostApiLogin(&config),
	)
	mux.HandleFunc(
		"POST /api/refresh",
		web2.HandlerPostApiRefresh(&config),
	)
	mux.HandleFunc(
		"POST /api/revoke",
		web2.HandlerPostApiRevoke(&config),
	)
	mux.HandleFunc(
		"GET /api/chirps/{id}",
		web2.HandlerGetApiChirpsId(&config),
	)
	mux.HandleFunc(
		"GET /api/chirps",
		web2.HandlerGetApiChirps(&config),
	)
	mux.HandleFunc(
		"POST /api/chirps",
		web2.HandlerPostApiChirps(&config, profanities),
	)
	mux.HandleFunc(
		"DELETE /api/chirps/{id}",
		web2.HandlerDeleteApiChirpsId(&config),
	)
	mux.HandleFunc(
		"POST /api/polka/webhooks",
		web2.HandlerPostApiPolkaWebhooks(&config),
	)

	log.Fatal(server.ListenAndServe())
}
