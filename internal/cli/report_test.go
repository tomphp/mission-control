package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tomoram/mission-control/internal/transport"
)

// withStdin replaces os.Stdin with a pipe pre-loaded with content, and
// restores the original when the test ends.
func withStdin(t *testing.T, content string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatal(err)
	}
	w.Close()

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = orig })
}

func TestRunReportPostsEnrichedEvent(t *testing.T) {
	gotRequest := make(chan transport.Event, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ingest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var evt transport.Event
		if err := json.Unmarshal(body, &evt); err != nil {
			t.Errorf("decode body: %v", err)
		}
		gotRequest <- evt
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("MISSION_CONTROL_PORT", strings.TrimPrefix(srv.URL, "http://127.0.0.1:"))

	cwd := t.TempDir()
	payload := `{"session_id":"abc123","hook_event_name":"UserPromptSubmit","cwd":"` + cwd + `"}`
	withStdin(t, payload)

	if code := runReport(nil); code != 0 {
		t.Fatalf("runReport() = %d, want 0", code)
	}

	evt := <-gotRequest
	if evt.SessionID != "abc123" {
		t.Errorf("SessionID = %q, want abc123", evt.SessionID)
	}
	if evt.HookEventName != "UserPromptSubmit" {
		t.Errorf("HookEventName = %q, want UserPromptSubmit", evt.HookEventName)
	}
	if evt.Cwd != cwd {
		t.Errorf("Cwd = %q, want %q", evt.Cwd, cwd)
	}
}

func TestRunReportMalformedStdinReturnsZero(t *testing.T) {
	withStdin(t, "not json")
	if code := runReport(nil); code != 0 {
		t.Fatalf("runReport() = %d, want 0 (a hook must never fail the session)", code)
	}
}

func TestRunReportMissingRequiredFieldReturnsZero(t *testing.T) {
	withStdin(t, `{"hook_event_name":"UserPromptSubmit","cwd":"/tmp"}`)
	if code := runReport(nil); code != 0 {
		t.Fatalf("runReport() = %d, want 0 (a hook must never fail the session)", code)
	}
}

func TestRunReportServerUnavailableReturnsZero(t *testing.T) {
	t.Setenv("MISSION_CONTROL_PORT", "1") // nothing listening on this port
	cwd := t.TempDir()
	payload := `{"session_id":"abc123","hook_event_name":"UserPromptSubmit","cwd":"` + cwd + `"}`
	withStdin(t, payload)

	if code := runReport(nil); code != 0 {
		t.Fatalf("runReport() = %d, want 0 (a hook must never fail the session)", code)
	}
}
