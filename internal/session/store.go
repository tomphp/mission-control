package session

import (
	"sort"
	"sync"
	"time"
)

// Store is the persistence boundary for tracked sessions. It has no
// knowledge of HTTP, SSE, or JSON — just session bookkeeping.
type Store interface {
	// Upsert creates a session for id if absent, then applies mutate to it
	// (new or existing) and returns the resulting value. mutate is
	// responsible for setting any fields it cares about, including
	// LastEventAt — the store does not implicitly touch it.
	Upsert(id ID, mutate func(*Session)) Session
	Remove(id ID)
	List() []Session
	// Prune removes sessions whose LastEventAt is older than ttl and
	// returns the IDs removed.
	Prune(ttl time.Duration) []ID
}

// MemoryStore is an in-memory, concurrency-safe Store implementation.
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[ID]*Session
	now      func() time.Time
}

// NewStore returns a MemoryStore using the real wall clock.
func NewStore() *MemoryStore {
	return NewStoreWithClock(time.Now)
}

// NewStoreWithClock returns a MemoryStore using the given clock, so
// TTL-based pruning can be tested deterministically.
func NewStoreWithClock(now func() time.Time) *MemoryStore {
	return &MemoryStore{
		sessions: make(map[ID]*Session),
		now:      now,
	}
}

func (s *MemoryStore) Upsert(id ID, mutate func(*Session)) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		now := s.now()
		sess = &Session{ID: id, StartedAt: now, LastEventAt: now, State: StateIdle}
		s.sessions[id] = sess
	}
	mutate(sess)
	return *sess
}

func (s *MemoryStore) Remove(id ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *MemoryStore) List() []Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		out = append(out, *sess)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *MemoryStore) Prune(ttl time.Duration) []ID {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	var removed []ID
	for id, sess := range s.sessions {
		if now.Sub(sess.LastEventAt) > ttl {
			delete(s.sessions, id)
			removed = append(removed, id)
		}
	}
	sort.Slice(removed, func(i, j int) bool { return removed[i] < removed[j] })
	return removed
}
