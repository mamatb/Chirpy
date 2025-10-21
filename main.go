package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mamatb/Chirpy/database"
	"github.com/mamatb/Chirpy/web"
)

const (
	driverName  = "postgres"
	envDbUrl    = "DB_URL"
	envPlatform = "PLATFORM"
	envPolkaKey = "POLKA_KEY"
	envSecret   = "SECRET"
	tcpPort     = ":8080"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    tcpPort,
		Handler: mux,
	}
	config := web.ApiConfig{
		Platform: os.Getenv(envPlatform),
		PolkaKey: os.Getenv(envPolkaKey),
		Secret:   os.Getenv(envSecret),
	}
	if db, err := sql.Open(driverName, os.Getenv(envDbUrl)); err != nil {
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
		web.HandlerGetApiHealth(),
	)
	mux.Handle(
		"/app/",
		web.HandlerApp(&config),
	)
	mux.HandleFunc(
		"GET /admin/metrics",
		web.HandlerGetAdminMetrics(&config),
	)
	mux.HandleFunc(
		"POST /admin/reset",
		web.HandlerPostAdminReset(&config),
	)
	mux.HandleFunc(
		"POST /api/users",
		web.HandlerPostApiUsers(&config),
	)
	mux.HandleFunc(
		"PUT /api/users",
		web.HandlerPutApiUsers(&config),
	)
	mux.HandleFunc(
		"POST /api/login",
		web.HandlerPostApiLogin(&config),
	)
	mux.HandleFunc(
		"POST /api/refresh",
		web.HandlerPostApiRefresh(&config),
	)
	mux.HandleFunc(
		"POST /api/revoke",
		web.HandlerPostApiRevoke(&config),
	)
	mux.HandleFunc(
		"GET /api/chirps/{id}",
		web.HandlerGetApiChirpsId(&config),
	)
	mux.HandleFunc(
		"GET /api/chirps",
		web.HandlerGetApiChirps(&config),
	)
	mux.HandleFunc(
		"POST /api/chirps",
		web.HandlerPostApiChirps(&config, profanities),
	)
	mux.HandleFunc(
		"DELETE /api/chirps/{id}",
		web.HandlerDeleteApiChirpsId(&config),
	)
	mux.HandleFunc(
		"POST /api/polka/webhooks",
		web.HandlerPostApiPolkaWebhooks(&config),
	)

	log.Fatal(server.ListenAndServe())
}
