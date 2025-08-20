package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/justanamir/medappoint/internal/api"
	"github.com/justanamir/medappoint/internal/config"
	dbconn "github.com/justanamir/medappoint/internal/db"
	"github.com/justanamir/medappoint/internal/db/gen"
)

var Version = "0.2.0-day2"

func main() {
	_ = godotenv.Load() // dev convenience
	cfg := config.FromEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx := context.Background()
	pg, err := dbconn.Connect(ctx, cfg.PGConnString(""))
	if err != nil {
		logger.Error("db connect fail", "err", err)
		os.Exit(1)
	}
	defer pg.Close()

	queries := gen.New(pg.Pool)
	r := api.NewRouter()

	// v1 routes
	r.Route("/v1", func(r chi.Router) {
		ad := api.AuthDeps{Cfg: cfg, Q: queries}
		r.Post("/auth/register", ad.RegisterHandler)
		r.Post("/auth/login", ad.LoginHandler)

		cd := api.ClinicDeps{Q: queries}
		r.Get("/clinics", cd.ListClinicsHandler)

		pd := api.ProviderDeps{Q: queries}
		r.Get("/providers", pd.ListProvidersHandler)

		sd := api.ServiceDeps{Q: queries}
		r.Get("/services", sd.ListServicesHandler)

		avd := api.AvailabilityDeps{Q: queries}
		r.Get("/availabilities", avd.ListByProviderHandler)
	})

	hs := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", cfg.Port, "env", cfg.Env, "version", Version)
		errCh <- hs.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal", "signal", sig.String())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}

	ctxShutdown, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = hs.Shutdown(ctxShutdown)
}
