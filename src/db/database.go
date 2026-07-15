package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EngineKind identifies the supported relational database engine.
type EngineKind string

const (
	EngineMySQL      EngineKind = "mysql"
	EnginePostgreSQL EngineKind = "postgresql"
)

// Config describes the shared settings required to open a database connection.
type Config struct {
	Engine EngineKind
	DSN    string
}

// DriverFactory constructs a database/sql driver for a given engine.
type DriverFactory func(cfg Config) (*sql.DB, error)

// Registry stores factories for all supported relational engines.
type Registry struct {
	factories map[EngineKind]DriverFactory
}

// NewRegistry creates a registry with the built-in engine factories.
func NewRegistry() *Registry {
	return &Registry{factories: map[EngineKind]DriverFactory{}}
}

// Register adds or replaces a factory for the provided engine.
func (r *Registry) Register(engine EngineKind, factory DriverFactory) {
	if r.factories == nil {
		r.factories = map[EngineKind]DriverFactory{}
	}
	r.factories[engine] = factory
}

// Open opens a connection using the registered factory for the configured engine.
func (r *Registry) Open(cfg Config) (*sql.DB, error) {
	factory, ok := r.factories[cfg.Engine]
	if !ok {
		return nil, fmt.Errorf("unsupported database engine: %s", cfg.Engine)
	}
	return factory(cfg)
}

// DefaultRegistry returns a registry preconfigured with common relational engines.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(EngineMySQL, func(cfg Config) (*sql.DB, error) {
		return openSQLDriver("mysql", cfg.DSN)
	})
	r.Register(EnginePostgreSQL, func(cfg Config) (*sql.DB, error) {
		return openSQLDriver("postgres", cfg.DSN)
	})
	return r
}

func openSQLDriver(driverName, dsn string) (*sql.DB, error) {
	if strings.TrimSpace(driverName) == "" {
		return nil, fmt.Errorf("database driver name is required")
	}
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("database DSN is required")
	}
	return sql.Open(driverName, dsn)
}
