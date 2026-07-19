// Command query-api is the entry point for the query-api service — the REST surface
// per docs/research/06-api/README.md#rest-api. Per docs/design-system.md §6, this is
// the only place adapters and the domain layer meet.
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

	addr := ":8081"
	if v := os.Getenv("QUERY_API_ADDR"); v != "" {
		addr = v
	}

	// 1-4: config (above), adapters/domain/delivery (none yet — the /v1 route group and its
	// AuditedQueryHandler-wrapped handlers, per docs/design-system.md §1 and
	// docs/research/03-architecture/diagrams.md#component-diagrams, land here once
	// query-api's store adapters exist).
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
		logger.Info("query-api listening", "addr", addr)
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
