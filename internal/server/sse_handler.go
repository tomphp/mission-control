package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/tomoram/mission-control/internal/session"
)

// heartbeatInterval keeps long-lived SSE connections (and any proxies in
// between) alive with a periodic comment ping.
const heartbeatInterval = 15 * time.Second

// SSEHandler streams session state to a connected UI client: an initial
// full snapshot, then live "update"/"remove" events from the broadcaster.
type SSEHandler struct {
	Store       session.Store
	Broadcaster *SSEBroadcaster
}

func NewSSEHandler(store session.Store, broadcaster *SSEBroadcaster) *SSEHandler {
	return &SSEHandler{Store: store, Broadcaster: broadcaster}
}

func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	snapshot, err := json.Marshal(h.Store.List())
	if err != nil {
		http.Error(w, "failed to encode snapshot", http.StatusInternalServerError)
		return
	}
	writeSSE(w, "snapshot", snapshot)
	flusher.Flush()

	msgs, unsubscribe := h.Broadcaster.Subscribe()
	defer unsubscribe()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			writeSSE(w, msg.Event, msg.Data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, data []byte) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}
