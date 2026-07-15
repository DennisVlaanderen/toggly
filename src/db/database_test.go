package db

import (
	"database/sql"
	"testing"
)

func TestRegistryRejectsUnknownEngine(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.Open(Config{Engine: EngineKind("sqlite"), DSN: "ignored"})
	if err == nil {
		t.Fatal("expected unsupported database engine error")
	}
}

func TestRegistrySupportsCustomFactories(t *testing.T) {
	registry := NewRegistry()
	factory := func(cfg Config) (*sql.DB, error) {
		return sql.OpenDB(nil), nil
	}
	registry.Register(EngineMySQL, factory)
	registry.Register(EnginePostgreSQL, factory)

	for _, tc := range []struct {
		name   string
		engine EngineKind
	}{
		{name: "mysql", engine: EngineMySQL},
		{name: "postgresql", engine: EnginePostgreSQL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database, err := registry.Open(Config{Engine: tc.engine, DSN: "ignored"})
			if err != nil {
				t.Fatalf("expected %s engine to be supported: %v", tc.engine, err)
			}
			defer database.Close()
		})
	}
}

func TestOpenSQLDriverRequiresInputs(t *testing.T) {
	if _, err := openSQLDriver("", "dsn"); err == nil {
		t.Fatal("expected driver name validation")
	}
	if _, err := openSQLDriver("mysql", ""); err == nil {
		t.Fatal("expected DSN validation")
	}
}
