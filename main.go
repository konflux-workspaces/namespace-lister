package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/go-logr/logr"

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

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create controller
	l.Info("creating cache")
	ctrl, err := NewCache(ctx, l)
	if err != nil {
		return err
	}

	// build http server
	l.Info("building server")
	userHeader := getHeaderUsername()
	s := buildServer(l, ctrl, userHeader)

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
