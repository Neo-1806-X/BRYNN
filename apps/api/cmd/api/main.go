// Command api is the BRYNN API process.
//
// Usage:
//
//	api              run the HTTP server (configured through BRYNN_* variables)
//	api healthcheck  probe a locally running server; exit 0 if it is ready
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"brynn/api/internal/config"
	"brynn/api/internal/httpserver"
)

// Injected at link time: -ldflags "-X main.version=... -X main.commit=...".
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

// run is main without the process-global state, so it can be tested.
func run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		if args[0] == "healthcheck" {
			return healthcheck(getenv, stderr)
		}
		fmt.Fprintf(stderr, "unknown command %q (supported: healthcheck)\n", args[0])
		return 2
	}

	cfg, err := config.Load(getenv)
	if err != nil {
		fmt.Fprintf(stderr, "configuration error:\n%v\n", err)
		return 2
	}

	logger := slog.New(slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: cfg.LogLevel})).With(
		slog.String("service", "brynn-api"),
		slog.String("version", version),
		slog.String("env", string(cfg.Env)),
	)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.HTTPAddr)
	if err != nil {
		logger.Error("listen_failed", slog.String("addr", cfg.HTTPAddr), slog.String("error", err.Error()))
		return 1
	}

	srv := httpserver.New(httpserver.Options{
		Logger:       logger,
		Build:        httpserver.BuildInfo{Version: version, Commit: commit},
		MaxBodyBytes: cfg.MaxBodyBytes,
	})
	logger.Info("starting", slog.String("addr", ln.Addr().String()))
	if err := srv.Run(ctx, ln, cfg.DrainDelay, cfg.ShutdownTimeout); err != nil {
		logger.Error("server_failed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

// healthcheck probes /readyz on the local server. The container image has no
// shell or curl, so the binary checks itself; Docker calls this subcommand.
func healthcheck(getenv func(string) string, stderr io.Writer) int {
	addr := getenv("BRYNN_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		fmt.Fprintf(stderr, "healthcheck: invalid BRYNN_HTTP_ADDR %q: %v\n", addr, err)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/readyz", nil)
	if err != nil {
		fmt.Fprintf(stderr, "healthcheck: %v\n", err)
		return 2
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(stderr, "healthcheck: %v\n", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(stderr, "healthcheck: unexpected status %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
