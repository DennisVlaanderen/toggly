package main

import (
	"testing"

	"toggly/db"
)

func TestResolveDatabaseConfigFromEnvironmentUsesExplicitDSN(t *testing.T) {
	t.Setenv("TOGGLY_DB_ENGINE", "postgresql")
	t.Setenv("TOGGLY_DB_DSN", "postgres://user:pass@db.example.com:5432/app")
	t.Setenv("TOGGLY_DB_USER", "")
	t.Setenv("TOGGLY_DB_PASSWORD", "")
	t.Setenv("TOGGLY_DB_HOST", "")
	t.Setenv("TOGGLY_DB_PORT", "")
	t.Setenv("TOGGLY_DB_NAME", "")

	cfg, ok, err := resolveDatabaseConfigFromEnvironment()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatal("expected database config to be resolved")
	}
	if cfg.Engine != db.EnginePostgreSQL {
		t.Fatalf("expected engine %q, got %q", db.EnginePostgreSQL, cfg.Engine)
	}
	if cfg.DSN != "postgres://user:pass@db.example.com:5432/app" {
		t.Fatalf("expected explicit DSN to be preserved, got %q", cfg.DSN)
	}
}

func TestResolveDatabaseConfigFromEnvironmentBuildsDSNFromCredentials(t *testing.T) {
	t.Setenv("TOGGLY_DB_ENGINE", "mysql")
	t.Setenv("TOGGLY_DB_DSN", "")
	t.Setenv("TOGGLY_DB_USER", "app-user")
	t.Setenv("TOGGLY_DB_PASSWORD", "super-secret")
	t.Setenv("TOGGLY_DB_HOST", "localhost")
	t.Setenv("TOGGLY_DB_PORT", "3306")
	t.Setenv("TOGGLY_DB_NAME", "toggly")

	cfg, ok, err := resolveDatabaseConfigFromEnvironment()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatal("expected database config to be resolved")
	}
	if cfg.Engine != db.EngineMySQL {
		t.Fatalf("expected engine %q, got %q", db.EngineMySQL, cfg.Engine)
	}
	if cfg.DSN != "app-user:super-secret@tcp(localhost:3306)/toggly" {
		t.Fatalf("expected constructed DSN, got %q", cfg.DSN)
	}
}
