package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tomoram/mission-control/internal/hook"
	"github.com/tomoram/mission-control/internal/transport"
)

func TestHTTPReporter_Report_Success(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody transport.Event

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("server: decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	reporter := transport.NewHTTPReporter(srv.URL, time.Second)
	event := transport.Event{
		Payload: hook.Payload{
			SessionID:     "sess-1",
			Cwd:           "/tmp/proj",
			HookEventName: hook.PreToolUse,
		},
		GitBranch: "main",
	}

	if err := reporter.Report(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/ingest" {
		t.Errorf("path = %q, want /ingest", gotPath)
	}
	if gotBody.SessionID != "sess-1" || gotBody.GitBranch != "main" {
		t.Errorf("body = %+v, want session_id=sess-1 git_branch=main", gotBody)
	}
}

func TestHTTPReporter_Report_ServerUnreachable_FailsFast(t *testing.T) {
	// Nothing listens on this port.
	reporter := transport.NewHTTPReporter("http://127.0.0.1:1", 300*time.Millisecond)
	event := transport.Event{Payload: hook.Payload{SessionID: "sess-1", Cwd: "/tmp", HookEventName: hook.Stop}}

	start := time.Now()
	err := reporter.Report(context.Background(), event)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error when server is unreachable, got nil")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Report took %v, want it to fail fast (well under the configured timeout budget)", elapsed)
	}
}

func TestHTTPReporter_Report_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	reporter := transport.NewHTTPReporter(srv.URL, time.Second)
	event := transport.Event{Payload: hook.Payload{SessionID: "sess-1", Cwd: "/tmp", HookEventName: hook.Stop}}

	if err := reporter.Report(context.Background(), event); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}
