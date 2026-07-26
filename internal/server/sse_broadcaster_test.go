package server_test

import (
	"testing"
	"time"

	"github.com/tomoram/mission-control/internal/server"
	"github.com/tomoram/mission-control/internal/session"
)

func TestSSEBroadcaster_PublishDeliversToSubscribers(t *testing.T) {
	b := server.NewSSEBroadcaster()

	ch1, unsub1 := b.Subscribe()
	defer unsub1()
	ch2, unsub2 := b.Subscribe()
	defer unsub2()

	sess := session.Session{ID: "sess-1", State: session.StateWorking}
	b.Publish(sess)

	for i, ch := range []<-chan server.Message{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg.Event != "update" {
				t.Errorf("subscriber %d: Event = %q, want %q", i, msg.Event, "update")
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out waiting for publish", i)
		}
	}
}

func TestSSEBroadcaster_RemoveDeliversRemoveEvent(t *testing.T) {
	b := server.NewSSEBroadcaster()
	ch, unsub := b.Subscribe()
	defer unsub()

	b.Remove("sess-1")

	select {
	case msg := <-ch:
		if msg.Event != "remove" {
			t.Errorf("Event = %q, want %q", msg.Event, "remove")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remove event")
	}
}

func TestSSEBroadcaster_UnsubscribeStopsDelivery(t *testing.T) {
	b := server.NewSSEBroadcaster()
	ch, unsub := b.Subscribe()
	unsub()

	b.Publish(session.Session{ID: "sess-1"})

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("received message on unsubscribed channel")
		}
		// channel closed, as expected.
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}

func TestSSEBroadcaster_SlowSubscriberDoesNotBlockOthers(t *testing.T) {
	b := server.NewSSEBroadcaster()
	slow, unsubSlow := b.Subscribe()
	defer unsubSlow()
	fast, unsubFast := b.Subscribe()
	defer unsubFast()

	// Fill the slow subscriber's buffer without ever draining it, then
	// publish more — this must not block the broadcaster or the fast
	// subscriber.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			b.Publish(session.Session{ID: "sess-1"})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a slow subscriber")
	}

	select {
	case <-fast:
	case <-time.After(time.Second):
		t.Fatal("fast subscriber never received a message")
	}
	_ = slow
}
