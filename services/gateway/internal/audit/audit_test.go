package audit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAddAndList(t *testing.T) {
	m := New()
	m.Add(Event{RunID: "r1", Kind: "started", At: time.Now()})
	m.Add(Event{RunID: "r1", Kind: "finished", At: time.Now().Add(time.Second)})
	m.Add(Event{RunID: "r2", Kind: "started", At: time.Now()})
	got, _ := m.List(context.Background(), "r1", 10)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

func TestSubscribe(t *testing.T) {
	m := New()
	var wg sync.WaitGroup
	wg.Add(1)
	_, err := m.Subscribe(context.Background(), "r1", func(e Event) { wg.Done() })
	if err != nil {
		t.Fatal(err)
	}
	m.Add(Event{RunID: "r1", Kind: "x"})
	wg.Wait()
}
