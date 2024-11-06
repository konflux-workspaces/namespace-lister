package main

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	l := buildLogger()
	if err := run(l); err != nil {
		l.Error("error running the server", "error", err)
		os.Exit(1)
	}
}

func run(l *slog.Logger) error {
	log.SetLogger(logr.FromSlogHandler(l.Handler()))

	l.Info("configuring k8s clients")
	// get k8s rest config
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create controller
	l.Info("creating controller")
	ctrl, err := NewController(ctx, l)
	if err != nil {
		return err
	}

	// build http server
	l.Info("building server")
	userHeader := cmp.Or(os.Getenv("HEADER_USERNAME"), "X-Email")
	s := buildServer(cfg, l, ctrl, userHeader)

	// HTTP Server graceful shutdown
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.Shutdown(sctx); err != nil {
			l.Error("error gracefully shutting down the HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// start server
	l.Info("serving...")
	return s.ListenAndServe()
}

func addLogMiddleware(l *slog.Logger, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("received request", "request", r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

func buildServer(cfg *rest.Config, l *slog.Logger, ctrl *Controller, userHeader string) *http.Server {
	// configure the server
	h := http.NewServeMux()
	h.Handle("GET /api/v1/namespaces", addLogMiddleware(l, newListNamespacesHandler(rest.CopyConfig(cfg), l, ctrl, userHeader)))
	return &http.Server{
		Addr:              cmp.Or(os.Getenv("ADDRESS"), DefaultAddr),
		Handler:           h,
		ReadHeaderTimeout: 3 * time.Second,
	}
}
