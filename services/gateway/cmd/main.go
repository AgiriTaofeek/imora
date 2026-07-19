// Command gateway is the entry point for the gateway service — the single routing
// chokepoint per docs/research/02-domain/README.md#bounded-contexts. Per
// docs/design-system.md §6, this is the only place adapters and the domain layer meet.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := ":8080"
	if v := os.Getenv("GATEWAY_ADDR"); v != "" {
		addr = v
	}

	// 1-4: config (above), adapters/domain/delivery (none yet — no handlers exist to route
	// to besides the health check below; AuthMiddleware/RateLimitMiddleware land here once
	// their underlying services exist, per docs/coding-standards.md §9).
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 5: start, then block on the shutdown signal — per docs/coding-standards.md §8.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("gateway listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown did not complete cleanly", "error", err)
	}
}
