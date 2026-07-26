package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/tomoram/mission-control/internal/hook"
	"github.com/tomoram/mission-control/internal/session"
	"github.com/tomoram/mission-control/internal/transport"
)

// IngestHandler receives enriched hook events from the CLI's report
// subcommand, updates the session store via the state machine, and
// publishes the result.
type IngestHandler struct {
	Store     session.Store
	Publisher Publisher
	Now       func() time.Time
}

func NewIngestHandler(store session.Store, publisher Publisher) *IngestHandler {
	return &IngestHandler{Store: store, Publisher: publisher, Now: time.Now}
}

func (h *IngestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var evt transport.Event
	if err := json.NewDecoder(r.Body).Decode(&evt); err != nil {
		http.Error(w, "malformed json", http.StatusBadRequest)
		return
	}
	if evt.SessionID == "" || evt.HookEventName == "" || evt.Cwd == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	id := session.ID(evt.SessionID)

	if evt.HookEventName == hook.SessionEnd {
		h.Store.Remove(id)
		h.Publisher.Remove(id)
		w.WriteHeader(http.StatusOK)
		return
	}

	now := h.Now()
	updated := h.Store.Upsert(id, func(sess *session.Session) {
		sess.Cwd = evt.Cwd
		sess.Folder = filepath.Base(evt.Cwd)
		sess.GitBranch = evt.GitBranch
		sess.LastEvent = string(evt.HookEventName)
		sess.LastEventAt = now
		sess.State, _ = session.NextState(sess.State, evt.HookEventName)
	})
	h.Publisher.Publish(updated)

	w.WriteHeader(http.StatusOK)
}
