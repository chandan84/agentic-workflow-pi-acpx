package events

import (
	"context"
	"sync"
	"testing"
)

func TestInMemoryBusRoundtrip(t *testing.T) {
	b := NewInMemoryBus()
	var got []byte
	var mu sync.Mutex
	done := make(chan struct{})
	_, err := b.Subscribe(context.Background(), "foo", func(_ string, data []byte) {
		mu.Lock()
		got = data
		mu.Unlock()
		close(done)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Publish(context.Background(), "foo", map[string]string{"hello": "world"}); err != nil {
		t.Fatal(err)
	}
	<-done
	mu.Lock()
	defer mu.Unlock()
	if string(got) == "" {
		t.Fatal("no payload received")
	}
}

func TestAllSubjects(t *testing.T) {
	if len(AllSubjects()) < 10 {
		t.Fatal("expected catalog")
	}
}
