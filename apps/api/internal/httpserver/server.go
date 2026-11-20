// Package httpserver wires the HTTP surface of the API process: routing,
// middleware, health and readiness probes, and the graceful shutdown sequence
// that lets a rolling deploy drain traffic before the process exits.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// BuildInfo identifies the running build. It is injected at link time.
type BuildInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// Options configures New.
type Options struct {
	// Logger receives structured logs. Defaults to slog.Default().
	Logger *slog.Logger
	// Build is reported by GET /version.
	Build BuildInfo
	// MaxBodyBytes caps request bodies. Defaults to 1 MiB.
	MaxBodyBytes int64
}

// Server is the API's HTTP server. It tracks readiness separately from
// liveness: a draining server is alive but no longer ready, which tells load
// balancers to stop sending it new requests.
type Server struct {
	srv   *http.Server
	log   *slog.Logger
	ready atomic.Bool
}

// New builds the server and its handler chain. It does not start listening.
//
// The HTTP timeouts below suit short request/response APIs. The future
// realtime gateway (WebSockets) is a separate process role and will use its
// own limits.
func New(opts Options) *Server {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}
	maxBody := opts.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = 1 << 20
	}

	s := &Server{log: logger}
	s.ready.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", allow(http.MethodGet, s.handleHealthz))
	mux.HandleFunc("/readyz", allow(http.MethodGet, s.handleReadyz))
	mux.HandleFunc("/version", allow(http.MethodGet, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, opts.Build)
	}))
	mux.HandleFunc("/", handleNotFound)

	// Innermost first: the last wrapper added is the first to see a request.
	var h http.Handler = mux
	h = limitBody(maxBody, h)
	h = securityHeaders(h)
	h = recoverer(logger, h)
	h = accessLog(logger, h)
	h = requestID(h)

	s.srv = &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
	return s
}

// Run serves on ln until ctx is cancelled, then shuts down gracefully:
//  1. report not-ready so load balancers stop routing here,
//  2. wait drainDelay,
//  3. stop accepting connections and let in-flight requests finish, for at
//     most shutdownTimeout.
//
// It returns nil after a clean shutdown.
func (s *Server) Run(ctx context.Context, ln net.Listener, drainDelay, shutdownTimeout time.Duration) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.srv.Serve(ln) }()

	select {
	case err := <-serveErr:
		// Serve only returns before shutdown if the listener failed.
		return fmt.Errorf("http serve: %w", err)
	case <-ctx.Done():
	}

	s.log.Info("shutdown_started", slog.Duration("drain_delay", drainDelay))
	s.ready.Store(false)
	if drainDelay > 0 {
		time.Sleep(drainDelay)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		_ = s.srv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http serve: %w", err)
	}
	s.log.Info("shutdown_complete")
	return nil
}

// allow restricts a handler to one method (plus HEAD for GET) and answers
// everything else with a 405 problem response.
func allow(method string, h http.HandlerFunc) http.HandlerFunc {
	allowed := method
	if method == http.MethodGet {
		allowed = "GET, HEAD"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method || (method == http.MethodGet && r.Method == http.MethodHead) {
			h(w, r)
			return
		}
		w.Header().Set("Allow", allowed)
		writeProblem(w, r, http.StatusMethodNotAllowed, "method-not-allowed", "Method not allowed", "")
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, r, http.StatusNotFound, "not-found", "Resource not found", "")
}

// handleHealthz is the liveness probe: the process is up and serving.
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleReadyz is the readiness probe: it fails while the server drains.
// Future dependency checks (database, queue) belong here.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !s.ready.Load() {
		writeProblem(w, r, http.StatusServiceUnavailable, "not-ready", "Service is not accepting new work", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
