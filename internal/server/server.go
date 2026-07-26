package server

import (
	"context"
	"io/fs"
	"net/http"
	"time"

	"github.com/tomoram/mission-control/internal/session"
)

// New wires the ingest, SSE, and static routes into a single handler.
func New(store session.Store, broadcaster *SSEBroadcaster, uiFS fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ingest", NewIngestHandler(store, broadcaster))
	mux.Handle("/events", NewSSEHandler(store, broadcaster))
	mux.Handle("/", NewStaticHandler(uiFS))
	return mux
}

// RunPruneLoop periodically removes sessions that haven't reported an event
// within ttl, as a safety net for sessions whose SessionEnd hook never
// fired (e.g. a killed terminal), and tells publisher about each removal so
// connected UI clients stay in sync. It blocks until ctx is done.
func RunPruneLoop(ctx context.Context, store session.Store, publisher Publisher, ttl, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, id := range store.Prune(ttl) {
				publisher.Remove(id)
			}
		}
	}
}
