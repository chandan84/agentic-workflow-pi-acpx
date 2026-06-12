// Package config is the layered configuration loader shared by every service.
//
// Layering order (later wins): built-in defaults → repo config.yaml → user file
// (path from <PREFIX>_CONFIG env) → <PREFIX>_… env vars → flags supplied via Apply.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Base is embedded by every service's typed Config struct. It carries cross-
// cutting concerns: service identity, logging, telemetry, auth, dependency
// endpoints, feature flags.
type Base struct {
	Service   string            `yaml:"service"`
	Env       string            `yaml:"env"`
	GRPCAddr  string            `yaml:"grpc_addr"`
	HTTPAddr  string            `yaml:"http_addr"`
	LogLevel  string            `yaml:"log_level"`
	LogFormat string            `yaml:"log_format"`
	Postgres  Postgres          `yaml:"postgres"`
	NATS      NATS              `yaml:"nats"`
	Temporal  Temporal          `yaml:"temporal"`
	Auth      Auth              `yaml:"auth"`
	OTel      OTel              `yaml:"otel"`
	Features  map[string]bool   `yaml:"features"`
	Stubs     map[string]string `yaml:"stubs"` // port name → "real"|"stub"
}

type Postgres struct {
	DSN    string `yaml:"dsn"`
	Schema string `yaml:"schema"`
}

type NATS struct {
	URL    string `yaml:"url"`
	Stream string `yaml:"stream"`
}

type Temporal struct {
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace"`
	TaskQueue string `yaml:"task_queue"`
}

type Auth struct {
	Mode  string `yaml:"mode"`  // none|static-token
	Token string `yaml:"token"` // when mode=static-token
}

type OTel struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// Defaults returns sensible local-dev defaults for the Base.
func Defaults(service string) Base {
	return Base{
		Service:   service,
		Env:       "dev",
		GRPCAddr:  ":0",
		HTTPAddr:  ":0",
		LogLevel:  "info",
		LogFormat: "json",
		Postgres: Postgres{
			DSN:    "postgres://awpa:awpa@localhost:5432/awpa?sslmode=disable",
			Schema: service,
		},
		NATS:     NATS{URL: "nats://localhost:4222", Stream: "AWPA"},
		Temporal: Temporal{HostPort: "localhost:7233", Namespace: "default", TaskQueue: "awpa-" + service},
		Auth:     Auth{Mode: "none"},
		OTel:     OTel{Enabled: false},
		Features: map[string]bool{},
		Stubs:    map[string]string{},
	}
}

// Load reads, in order:
//  1. defaults (caller's zero-value or pre-populated struct)
//  2. ./config.yaml (repo defaults) if present
//  3. file at $<PREFIX>_CONFIG if set
//  4. env vars (<PREFIX>_<KEY>) — flat YAML-style keys, underscores split
//
// PREFIX is uppercase-snake of the service name (e.g. "AGENTS", "FLOW").
// Errors carry the precise field path that failed.
func Load[T any](service string, dst *T) error {
	prefix := strings.ToUpper(strings.ReplaceAll(service, "-", "_"))

	if _, err := os.Stat("config.yaml"); err == nil {
		if err := readYAML("config.yaml", dst); err != nil {
			return fmt.Errorf("config.yaml: %w", err)
		}
	}
	if p := os.Getenv(prefix + "_CONFIG"); p != "" {
		ap, _ := filepath.Abs(p)
		if err := readYAML(ap, dst); err != nil {
			return fmt.Errorf("%s: %w", ap, err)
		}
	}
	if err := overlayEnv(prefix, dst); err != nil {
		return fmt.Errorf("env overlay: %w", err)
	}
	return nil
}

func readYAML(path string, dst any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, dst)
}

// overlayEnv flattens the struct's yaml keys to <PREFIX>_FOO_BAR style and
// pulls values from the environment when set. Only string, int, bool, and
// string maps are walked.
func overlayEnv(prefix string, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("overlayEnv: pointer-to-struct required")
	}
	return walkEnv(prefix, v.Elem())
}

func walkEnv(prefix string, v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		tag := ft.Tag.Get("yaml")
		name, opts := parseYAMLTag(tag, ft.Name)
		if name == "-" {
			continue
		}
		// ",inline" or anonymous embedded struct: keep parent prefix.
		var key string
		if opts.inline || (ft.Anonymous && fv.Kind() == reflect.Struct && name == strings.ToLower(ft.Name)) {
			key = prefix
		} else {
			key = strings.ToUpper(prefix + "_" + strings.ReplaceAll(name, "-", "_"))
		}
		switch fv.Kind() {
		case reflect.Struct:
			if err := walkEnv(key, fv); err != nil {
				return err
			}
		case reflect.String:
			if e := os.Getenv(key); e != "" {
				fv.SetString(e)
			}
		case reflect.Bool:
			if e := os.Getenv(key); e != "" {
				b, err := strconv.ParseBool(e)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				fv.SetBool(b)
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if e := os.Getenv(key); e != "" {
				n, err := strconv.ParseInt(e, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: %w", key, err)
				}
				fv.SetInt(n)
			}
		}
	}
	return nil
}

type yamlOpts struct{ inline bool }

func parseYAMLTag(tag, fallback string) (string, yamlOpts) {
	if tag == "" {
		return strings.ToLower(fallback), yamlOpts{}
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		name = strings.ToLower(fallback)
	}
	var opts yamlOpts
	for _, p := range parts[1:] {
		if p == "inline" {
			opts.inline = true
		}
	}
	return name, opts
}

// Validate runs basic required-field checks on Base.
func (b Base) Validate() error {
	if b.Service == "" {
		return fmt.Errorf("service: required")
	}
	if b.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn: required")
	}
	return nil
}
