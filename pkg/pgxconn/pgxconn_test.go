package pgxconn

import (
	"context"
	"testing"
)

func TestOpenEmptyDSN(t *testing.T) {
	if _, err := Open(context.Background(), Options{}); err == nil {
		t.Fatal("expected error on empty DSN")
	}
}
