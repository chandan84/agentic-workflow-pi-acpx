package authn

import (
	"context"
	"testing"
)

func TestNone(t *testing.T) {
	p, err := None{}.Authenticate(context.Background(), "")
	if err != nil || p.Subject != "anonymous" {
		t.Fatalf("none: %v %+v", err, p)
	}
}

func TestStaticToken(t *testing.T) {
	a := StaticToken{Token: "s3cret"}
	if _, err := a.Authenticate(context.Background(), "Bearer s3cret"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Authenticate(context.Background(), "Bearer wrong"); err == nil {
		t.Fatal("expected unauthorized")
	}
}

func TestFromMode(t *testing.T) {
	if _, ok := From("none", "").(None); !ok {
		t.Fatal("expected None")
	}
	if _, ok := From("static-token", "x").(StaticToken); !ok {
		t.Fatal("expected StaticToken")
	}
}
