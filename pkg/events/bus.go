package events

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// Bus is the port every service uses to publish and subscribe.
type Bus interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(ctx context.Context, subject string, handler func(subject string, data []byte)) (Unsubscribe, error)
	Close() error
}

// Unsubscribe cancels a subscription.
type Unsubscribe func() error

// NATSBus is the production implementation backed by JetStream for durable
// subjects and core NATS for ephemeral subscriptions.
type NATSBus struct {
	conn *nats.Conn
}

// DialNATS opens a NATS connection.
func DialNATS(url string) (*NATSBus, error) {
	c, err := nats.Connect(url, nats.Timeout(2*time.Second), nats.MaxReconnects(-1))
	if err != nil {
		return nil, err
	}
	return &NATSBus{conn: c}, nil
}

// Publish encodes the payload as JSON and publishes on subject.
func (b *NATSBus) Publish(_ context.Context, subject string, payload any) error {
	if b == nil || b.conn == nil {
		return errors.New("nats: not connected")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return b.conn.Publish(subject, body)
}

// Subscribe creates an ephemeral subscription invoking handler per message.
func (b *NATSBus) Subscribe(_ context.Context, subject string, handler func(subject string, data []byte)) (Unsubscribe, error) {
	sub, err := b.conn.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Subject, m.Data)
	})
	if err != nil {
		return nil, err
	}
	return func() error { return sub.Unsubscribe() }, nil
}

// Close drains and closes the connection.
func (b *NATSBus) Close() error {
	if b == nil || b.conn == nil {
		return nil
	}
	return b.conn.Drain()
}

// InMemoryBus is a process-local implementation for tests and the stub mode.
type InMemoryBus struct {
	mu   sync.RWMutex
	subs map[string][]func(string, []byte)
}

// NewInMemoryBus returns a fresh in-memory bus.
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{subs: map[string][]func(string, []byte){}}
}

// Publish fan-outs synchronously.
func (b *InMemoryBus) Publish(_ context.Context, subject string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	b.mu.RLock()
	handlers := append([]func(string, []byte){}, b.subs[subject]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		h(subject, body)
	}
	return nil
}

// Subscribe registers a handler.
func (b *InMemoryBus) Subscribe(_ context.Context, subject string, handler func(subject string, data []byte)) (Unsubscribe, error) {
	b.mu.Lock()
	b.subs[subject] = append(b.subs[subject], handler)
	idx := len(b.subs[subject]) - 1
	b.mu.Unlock()
	return func() error {
		b.mu.Lock()
		defer b.mu.Unlock()
		if list, ok := b.subs[subject]; ok && idx < len(list) {
			list[idx] = nil
		}
		return nil
	}, nil
}

// Close is a no-op for the in-memory bus.
func (b *InMemoryBus) Close() error { return nil }
