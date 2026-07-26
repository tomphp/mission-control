package session_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tomoram/mission-control/internal/session"
)

func TestStore_UpsertCreatesNewSession(t *testing.T) {
	s := session.NewStore()

	got := s.Upsert("sess-1", func(sess *session.Session) {
		sess.Cwd = "/tmp/proj"
		sess.State = session.StateIdle
	})

	if got.ID != "sess-1" {
		t.Errorf("ID = %q, want %q", got.ID, "sess-1")
	}
	if got.Cwd != "/tmp/proj" {
		t.Errorf("Cwd = %q, want %q", got.Cwd, "/tmp/proj")
	}

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("List() len = %d, want 1", len(list))
	}
}

func TestStore_UpsertMutatesExisting(t *testing.T) {
	s := session.NewStore()
	s.Upsert("sess-1", func(sess *session.Session) { sess.State = session.StateIdle })
	got := s.Upsert("sess-1", func(sess *session.Session) { sess.State = session.StateWorking })

	if got.State != session.StateWorking {
		t.Errorf("State = %q, want %q", got.State, session.StateWorking)
	}
	if len(s.List()) != 1 {
		t.Fatalf("List() len = %d, want 1 (upsert on existing id must not create a duplicate)", len(s.List()))
	}
}

func TestStore_Remove(t *testing.T) {
	s := session.NewStore()
	s.Upsert("sess-1", func(sess *session.Session) {})
	s.Remove("sess-1")

	if len(s.List()) != 0 {
		t.Fatalf("List() len = %d, want 0 after Remove", len(s.List()))
	}
}

func TestStore_Prune_RemovesStaleSessions(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s := session.NewStoreWithClock(func() time.Time { return now })

	s.Upsert("stale", func(sess *session.Session) { sess.LastEventAt = now.Add(-25 * time.Hour) })
	s.Upsert("fresh", func(sess *session.Session) { sess.LastEventAt = now.Add(-1 * time.Minute) })

	removed := s.Prune(24 * time.Hour)

	if len(removed) != 1 || removed[0] != "stale" {
		t.Errorf("Prune removed = %v, want [stale]", removed)
	}
	list := s.List()
	if len(list) != 1 || list[0].ID != "fresh" {
		t.Errorf("List() after prune = %v, want only 'fresh'", list)
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := session.NewStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			id := session.ID(string(rune('a' + i%26)))
			s.Upsert(id, func(sess *session.Session) { sess.State = session.StateWorking })
		}(i)
		go func() {
			defer wg.Done()
			s.List()
		}()
	}
	wg.Wait()
}
