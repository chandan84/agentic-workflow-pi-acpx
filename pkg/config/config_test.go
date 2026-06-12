package config

import (
	"os"
	"testing"
)

type svcCfg struct {
	Base `yaml:",inline"`
}

func TestDefaults(t *testing.T) {
	b := Defaults("agents")
	if b.Service != "agents" {
		t.Fatalf("service: %q", b.Service)
	}
	if b.Postgres.Schema != "agents" {
		t.Fatalf("schema: %q", b.Postgres.Schema)
	}
}

func TestEnvOverlay(t *testing.T) {
	cfg := svcCfg{Base: Defaults("agents")}
	t.Setenv("AGENTS_LOG_LEVEL", "debug")
	t.Setenv("AGENTS_POSTGRES_DSN", "postgres://x/y")
	if err := Load("agents", &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("log level: %q", cfg.LogLevel)
	}
	if cfg.Postgres.DSN != "postgres://x/y" {
		t.Fatalf("dsn: %q", cfg.Postgres.DSN)
	}
}

func TestValidate(t *testing.T) {
	b := Defaults("agents")
	if err := b.Validate(); err != nil {
		t.Fatal(err)
	}
	b.Postgres.DSN = ""
	if err := b.Validate(); err == nil {
		t.Fatal("expected dsn error")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
