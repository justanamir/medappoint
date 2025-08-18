// cmd/server/main.go
// Go Medical Appointment App — Day 1 Skeleton
// ------------------------------------------------------------
// What you have here:
// - Minimal HTTP server using chi router
// - Health check endpoint
// - Basic project scaffolding & domain types
// - Clear TODOs for the next 4 days
// ------------------------------------------------------------
// How to run (after you create a module):
//   $ mkdir -p cmd/server
//   $ cd cmd/server
//   $ go mod init github.com/yourname/medappoint
//   $ go get github.com/go-chi/chi/v5
//   $ go run .
// Then visit: http://localhost:8080/healthz
// ------------------------------------------------------------

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
	"github.com/go-chi/chi/v5/middleware"
)

// Version is the app version; update during releases.
var Version = "0.1.0-day1"

// Config holds environment-driven settings.
// In later days we will expand this with DB_URL, JWT secrets, etc.
type Config struct {
	Port int    // HTTP port, e.g., 8080
	Env  string // e.g., "dev", "staging", "prod"
}

// loadConfig reads environment variables with sensible defaults.
func loadConfig() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		fmt.Sscanf(v, "%d", &port)
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	return Config{Port: port, Env: env}
}

// Server wires up the HTTP router, logger, and configuration.
type Server struct {
	cfg    Config
	log    *slog.Logger
	router *chi.Mux
	http   *http.Server
}

// newServer constructs the Server with middlewares and routes configured.
func newServer(cfg Config) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	r := chi.NewRouter()
	// Middlewares: request ID, real IP, logging, recovery, timeouts
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Basic routes
	r.Get("/healthz", healthHandler)
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s\n", Version)
	})

	// API v1 placeholder (we'll fill these in on Day 2-3)
	r.Route("/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","message":"v1 root"}`))
		})
		// TODO: Day 2 - Auth endpoints: POST /auth/register, POST /auth/login
		// TODO: Day 3 - Providers/services/slots/appointments endpoints
		// TODO: Day 4 - Reschedule/cancel, admin availability mgmt, validation
	})

	// Configure http.Server with sane timeouts
	hs := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Server{cfg: cfg, log: logger, router: r, http: hs}
}

// Start runs the HTTP server and handles graceful shutdown on SIGINT/SIGTERM.
func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("server starting", "port", s.cfg.Port, "env", s.cfg.Env, "version", Version)
		errCh <- s.http.ListenAndServe()
	}()

	// Wait for a termination signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.log.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}
	}

	ctxShutdown, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.http.Shutdown(ctxShutdown)
}

// healthHandler responds with a simple 200 OK to indicate liveness.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// --------------------------- Domain Types ---------------------------
// These are light sketches so we can reason about the data model.
// On Day 2, we will turn this into a proper PostgreSQL schema & migrations.

// Role enumerates basic roles in the system.
type Role string

const (
	RolePatient  Role = "patient"
	RoleProvider Role = "provider"
	RoleAdmin    Role = "admin"
)

// User is a system account; it can belong to a Patient or Provider profile.
// Passwords will be stored as bcrypt hashes (Day 2).
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Patient captures patient-specific details.
type Patient struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	DOB       time.Time `json:"dob"` // optional
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Provider represents a clinician (doctor, dentist, physiotherapist, etc.).
type Provider struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	FullName   string    `json:"full_name"`
	Speciality string    `json:"speciality"` // e.g. "Cardiology"
	ClinicID   int64     `json:"clinic_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Clinic is the physical or virtual location offering services.
type Clinic struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Timezone  string    `json:"timezone"` // IANA tz, e.g., "Asia/Kuala_Lumpur"
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Service is a bookable item (e.g., "General Consultation, 30min").
type Service struct {
	ID          int64     `json:"id"`
	ClinicID    int64     `json:"clinic_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DurationMin int       `json:"duration_min"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Availability expresses a Provider's working hours (weekly pattern).
type Availability struct {
	ID         int64  `json:"id"`
	ProviderID int64  `json:"provider_id"`
	Weekday    int    `json:"weekday"`    // 1=Mon ... 7=Sun (ISO-8601)
	StartHHMM  string `json:"start_hhmm"` // e.g., "09:00"
	EndHHMM    string `json:"end_hhmm"`   // e.g., "17:00"
}

// Appointment is a confirmed booking between a Patient and Provider.
type Appointment struct {
	ID         int64     `json:"id"`
	ClinicID   int64     `json:"clinic_id"`
	ProviderID int64     `json:"provider_id"`
	PatientID  int64     `json:"patient_id"`
	ServiceID  int64     `json:"service_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"` // scheduled|completed|cancelled
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Holiday or blackout dates for a clinic or provider (e.g., public holidays).
type Blackout struct {
	ID         int64     `json:"id"`
	ClinicID   *int64    `json:"clinic_id,omitempty"`
	ProviderID *int64    `json:"provider_id,omitempty"`
	Date       time.Time `json:"date"`
	Reason     string    `json:"reason"`
}

// --------------------------- main() ---------------------------
// main loads configuration, builds the server, and starts it. Each
// function is intentionally small & well-documented to aid learning.
func main() {
	cfg := loadConfig()
	srv := newServer(cfg)
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		srv.log.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}
