package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(Options{Logger: logger, Build: BuildInfo{Version: "test", Commit: "abc"}})
}

func do(t *testing.T, s *Server, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	s.srv.Handler.ServeHTTP(rec, req)
	return rec
}

func TestHealthz(t *testing.T) {
	rec := do(t, newTestServer(t), http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Errorf("body = %q", got)
	}
}

func TestVersionReportsBuildInfo(t *testing.T) {
	rec := do(t, newTestServer(t), http.MethodGet, "/version", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got BuildInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Version != "test" || got.Commit != "abc" {
		t.Errorf("build info = %+v", got)
	}
}

func TestReadyzReflectsDrainState(t *testing.T) {
	s := newTestServer(t)
	if rec := do(t, s, http.MethodGet, "/readyz", nil); rec.Code != http.StatusOK {
		t.Fatalf("ready status = %d, want 200", rec.Code)
	}

	s.ready.Store(false)
	rec := do(t, s, http.MethodGet, "/readyz", nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("draining status = %d, want 503", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("content-type = %q, want application/problem+json", ct)
	}
}

func TestUnknownPathReturnsProblem(t *testing.T) {
	rec := do(t, newTestServer(t), http.MethodGet, "/nope", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	var p problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.Type != "urn:brynn:problem:not-found" || p.Status != http.StatusNotFound {
		t.Errorf("problem = %+v", p)
	}
	if p.RequestID == "" {
		t.Error("problem response must carry the request id")
	}
}

func TestWrongMethodReturns405(t *testing.T) {
	rec := do(t, newTestServer(t), http.MethodPost, "/healthz", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want \"GET, HEAD\"", got)
	}
}

func TestRequestIDHandling(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodGet, "/healthz", map[string]string{"X-Request-ID": "client-supplied-123"})
	if got := rec.Header().Get("X-Request-ID"); got != "client-supplied-123" {
		t.Errorf("a well-formed inbound ID should be kept, got %q", got)
	}

	for _, bad := range []string{"short", "has space in it", "line\nbreak-injection"} {
		rec = do(t, s, http.MethodGet, "/healthz", map[string]string{"X-Request-ID": bad})
		got := rec.Header().Get("X-Request-ID")
		if got == bad || len(got) != 32 {
			t.Errorf("unsafe inbound ID %q should be replaced by a generated one, got %q", bad, got)
		}
	}
}

func TestRecovererTurnsPanicIntoProblem(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := requestID(recoverer(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "boom") {
		t.Error("the panic value must not leak into the response")
	}
}

func TestRunDrainsAndStops(t *testing.T) {
	s := newTestServer(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, ln, 0, 2*time.Second) }()

	url := "http://" + ln.Addr().String() + "/readyz"
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(3 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready in time (last error: %v)", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() = %v, want nil after a clean shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after the context was cancelled")
	}
	if s.ready.Load() {
		t.Error("server should report not-ready once shutdown has begun")
	}
}
