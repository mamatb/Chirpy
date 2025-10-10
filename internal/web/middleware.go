package web

import "net/http"

func (c *ApiConfig) middleMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *ApiConfig) middleMetricsReset(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.FileserverHits.Store(0)
		next.ServeHTTP(w, r)
	}
}
