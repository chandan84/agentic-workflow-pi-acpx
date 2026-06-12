package obs

import (
	"context"
	"testing"
)

func TestLoggerAndCtx(t *testing.T) {
	l := Logger("svc", "debug", "json")
	if l == nil {
		t.Fatal("nil logger")
	}
	ctx := WithRunID(context.Background(), "abc")
	l2 := FromContext(ctx, l)
	if l2 == nil {
		t.Fatal("nil logger after ctx")
	}
}

func TestInitTracingStub(t *testing.T) {
	shutdown, err := InitTracing(context.Background(), "svc", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
