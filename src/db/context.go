package db

import "database/sql"

// Context provides a lightweight entry point for opening database connections
// through a configurable registry of relational engine factories.
type Context struct {
	registry *Registry
}

// NewContext creates a database context using the provided registry.
// If registry is nil, the default registry is used.
func NewContext(registry *Registry) *Context {
	if registry == nil {
		registry = DefaultRegistry()
	}
	return &Context{registry: registry}
}

// Open opens a connection using the configured engine and DSN.
func (c *Context) Open(cfg Config) (*sql.DB, error) {
	if c == nil || c.registry == nil {
		return DefaultRegistry().Open(cfg)
	}
	return c.registry.Open(cfg)
}

// Register adds or replaces a factory for the provided engine.
func (c *Context) Register(engine EngineKind, factory DriverFactory) {
	if c == nil {
		return
	}
	if c.registry == nil {
		c.registry = NewRegistry()
	}
	c.registry.Register(engine, factory)
}

// Open is a convenience helper that uses the default registry.
func Open(cfg Config) (*sql.DB, error) {
	return DefaultRegistry().Open(cfg)
}
