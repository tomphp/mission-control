// Package server wires the session store and state machine to HTTP:
// ingesting hook events and streaming session state to UI clients over SSE.
package server

import (
	"encoding/json"
	"sync"

	"github.com/tomoram/mission-control/internal/session"
)

// Message is one SSE event: Event names the event type ("update" or
// "remove"), Data is its already-JSON-encoded payload.
type Message struct {
	Event string
	Data  []byte
}

// Publisher is the fan-out boundary IngestHandler depends on, so it has no
// knowledge of the concrete SSE mechanism.
type Publisher interface {
	Publish(session.Session)
	Remove(session.ID)
}

// SSEBroadcaster fans out session changes to any number of subscribers
// (one per connected SSE client).
type SSEBroadcaster struct {
	mu   sync.Mutex
	subs map[chan Message]struct{}
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{subs: make(map[chan Message]struct{})}
}

// Subscribe registers a new subscriber and returns its channel along with
// an unsubscribe function that must be called exactly once to release it.
func (b *SSEBroadcaster) Subscribe() (<-chan Message, func()) {
	ch := make(chan Message, 16)

	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
	return ch, unsubscribe
}

func (b *SSEBroadcaster) Publish(s session.Session) {
	data, err := json.Marshal(s)
	if err != nil {
		return
	}
	b.broadcast(Message{Event: "update", Data: data})
}

func (b *SSEBroadcaster) Remove(id session.ID) {
	data, err := json.Marshal(struct {
		ID session.ID `json:"id"`
	}{ID: id})
	if err != nil {
		return
	}
	b.broadcast(Message{Event: "remove", Data: data})
}

// broadcast never blocks on a slow subscriber: a full buffer means the
// message is dropped for that subscriber rather than stalling everyone else.
func (b *SSEBroadcaster) broadcast(msg Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default:
		}
	}
}
