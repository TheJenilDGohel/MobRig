package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/database"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// Server is the HTTP REST API server for MobRig.
type Server struct {
	config    *config.Config
	registry  *platform.AdapterRegistry
	repos     *database.Repositories
	mux       *http.ServeMux
	srv       *http.Server
	listener  net.Listener
	startTime time.Time
}

// New creates a new HTTP server wired to the given config, adapter registry, and database repositories.
func New(cfg *config.Config, registry *platform.AdapterRegistry, repos *database.Repositories) *Server {
	s := &Server{
		config:    cfg,
		registry:  registry,
		repos:     repos,
		mux:       http.NewServeMux(),
		startTime: time.Now(),
	}

	s.registerRoutes()

	handler := corsMiddleware(recoverMiddleware(s.mux))

	s.srv = &http.Server{
		Addr:              net.JoinHostPort(cfg.HTTPHost, fmt.Sprintf("%d", cfg.HTTPPort)),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s
}

// Listen binds the network listener and updates the server address with the actual bound address.
func (s *Server) Listen() (net.Listener, error) {
	if s.listener != nil {
		return s.listener, nil
	}
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", s.srv.Addr, err)
	}
	s.listener = ln
	s.srv.Addr = ln.Addr().String()
	return ln, nil
}

// Start begins serving HTTP requests. This blocks until the server is shut down.
func (s *Server) Start() error {
	ln, err := s.Listen()
	if err != nil {
		return err
	}
	log.Printf("[http] listening on %s", s.srv.Addr)
	if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server with the given context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("[http] shutting down...")
	return s.srv.Shutdown(ctx)
}

// Addr returns the configured or actual listening address.
func (s *Server) Addr() string {
	return s.srv.Addr
}

// registerRoutes sets up all HTTP routes.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	// Device discovery & inspection
	s.mux.HandleFunc("GET /api/v1/devices", s.handleListDevices)
	s.mux.HandleFunc("GET /api/v1/devices/{id}", s.handleGetDevice)
	s.mux.HandleFunc("GET /api/v1/devices/{id}/screenshot", s.handleGetScreenshot)
	s.mux.HandleFunc("GET /api/v1/devices/{id}/ui-tree", s.handleGetUITree)
	s.mux.HandleFunc("GET /api/v1/devices/{id}/elements", s.handleFindElements)

	// Touch & gesture inputs
	s.mux.HandleFunc("POST /api/v1/devices/{id}/tap", s.handleTap)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/double-tap", s.handleDoubleTap)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/long-press", s.handleLongPress)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/swipe", s.handleSwipe)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/type", s.handleTypeText)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/key", s.handlePressKey)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/home", s.handleGoHome)

	// Application lifecycle
	s.mux.HandleFunc("POST /api/v1/devices/{id}/apps/launch", s.handleLaunchApp)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/apps/terminate", s.handleTerminateApp)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/apps/install", s.handleInstallApp)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/apps/uninstall", s.handleUninstallApp)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/apps/clear-data", s.handleClearAppData)
	s.mux.HandleFunc("POST /api/v1/devices/{id}/open-url", s.handleOpenURL)
}

// HealthResponse is the JSON payload for GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	resp := HealthResponse{
		Status:  "ok",
		Version: s.config.Version,
		Uptime:  time.Since(s.startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[http] panic recovered: %v", rec)
				writeProblem(w, http.StatusInternalServerError, "Internal Server Error", fmt.Sprintf("Unexpected panic: %v", rec), "https://mobrig.dev/errors/panic")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
