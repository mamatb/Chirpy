package main

import (
	"fmt"
	"log"
	"net/http"
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
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, _ *http.Request) {
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
	})
	mux.HandleFunc("POST /admin/reset", config.middlewareMetricsReset(handlerOK))
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
