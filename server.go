package main

import (
	"log/slog"
	"net/http"
	"time"
)

func addLogMiddleware(l *slog.Logger, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("received request", "request", r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func buildServer(l *slog.Logger, cache *Cache, userHeader string) *http.Server {
	// configure the server
	h := http.NewServeMux()
	h.Handle("GET /api/v1/namespaces", addLogMiddleware(l, newListNamespacesHandler(l, cache, userHeader)))
	return &http.Server{
		Addr:              getAddress(),
		Handler:           h,
		ReadHeaderTimeout: 3 * time.Second,
	}
}
