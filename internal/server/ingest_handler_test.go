package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tomoram/mission-control/internal/server"
	"github.com/tomoram/mission-control/internal/session"
)

type fakePublisher struct {
	published []session.Session
	removed   []session.ID
}

func (f *fakePublisher) Publish(s session.Session) { f.published = append(f.published, s) }
func (f *fakePublisher) Remove(id session.ID)      { f.removed = append(f.removed, id) }

func TestIngestHandler_ValidEvent_UpsertsAndPublishes(t *testing.T) {
	store := session.NewStore()
	pub := &fakePublisher{}
	h := server.NewIngestHandler(store, pub)

	body := `{"session_id":"sess-1","cwd":"/Users/tom/project","hook_event_name":"UserPromptSubmit","git_branch":"main"}`
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("store has %d sessions, want 1", len(list))
	}
	got := list[0]
	if got.ID != "sess-1" || got.Cwd != "/Users/tom/project" || got.Folder != "project" || got.GitBranch != "main" {
		t.Errorf("unexpected session: %+v", got)
	}
	if got.State != session.StateWorking {
		t.Errorf("State = %q, want %q", got.State, session.StateWorking)
	}

	if len(pub.published) != 1 {
		t.Fatalf("Publish called %d times, want 1", len(pub.published))
	}
}

func TestIngestHandler_SessionEnd_RemovesFromStore(t *testing.T) {
	store := session.NewStore()
	store.Upsert("sess-1", func(s *session.Session) { s.State = session.StateIdle })
	pub := &fakePublisher{}
	h := server.NewIngestHandler(store, pub)

	body := `{"session_id":"sess-1","cwd":"/tmp","hook_event_name":"SessionEnd"}`
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if len(store.List()) != 0 {
		t.Errorf("store has %d sessions after SessionEnd, want 0", len(store.List()))
	}
	if len(pub.removed) != 1 || pub.removed[0] != "sess-1" {
		t.Errorf("removed = %v, want [sess-1]", pub.removed)
	}
}

func TestIngestHandler_MalformedJSON_Returns400(t *testing.T) {
	h := server.NewIngestHandler(session.NewStore(), &fakePublisher{})

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString("{not json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestIngestHandler_MissingRequiredFields_Returns400(t *testing.T) {
	h := server.NewIngestHandler(session.NewStore(), &fakePublisher{})

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(`{"cwd":"/tmp"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestIngestHandler_WrongMethod_Returns405(t *testing.T) {
	h := server.NewIngestHandler(session.NewStore(), &fakePublisher{})

	req := httptest.NewRequest(http.MethodGet, "/ingest", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestIngestHandler_RepeatedEvents_UpdateSameSession(t *testing.T) {
	store := session.NewStore()
	pub := &fakePublisher{}
	h := server.NewIngestHandler(store, pub)

	post := func(eventName string) {
		body := `{"session_id":"sess-1","cwd":"/tmp/proj","hook_event_name":"` + eventName + `"}`
		req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("event %s: status = %d", eventName, w.Code)
		}
	}

	post("SessionStart")
	post("UserPromptSubmit")
	post("PreToolUse")
	post("Notification")
	post("Stop")

	list := store.List()
	if len(list) != 1 {
		t.Fatalf("store has %d sessions, want 1", len(list))
	}
	if list[0].State != session.StateIdle {
		t.Errorf("final State = %q, want %q", list[0].State, session.StateIdle)
	}
	if list[0].LastEvent != "Stop" {
		t.Errorf("LastEvent = %q, want %q", list[0].LastEvent, "Stop")
	}
	if time.Since(list[0].LastEventAt) > 5*time.Second {
		t.Errorf("LastEventAt not updated to recent time: %v", list[0].LastEventAt)
	}
}
