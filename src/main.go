package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"toggly/api"
	"toggly/db"
	"toggly/fqdp"
	"toggly/ws"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if _, err := initializeDatabaseFromEnvironment(); err != nil {
		log.Printf("database initialization warning: %v", err)
	}

	go func() {
		if err := fqdp.StartTCPServer(ctx, ":9000"); err != nil {
			log.Printf("fqdp server stopped: %v", err)
		}
	}()

	go func() {
		if err := ws.StartTCPServer(ctx, ":9001"); err != nil {
			log.Printf("raw tcp server stopped: %v", err)
		}
	}()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("http api listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
}

func initializeDatabaseFromEnvironment() (*db.Context, error) {
	cfg, ok, err := resolveDatabaseConfigFromEnvironment()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	ctx := db.NewContext(nil)
	conn, err := ctx.Open(cfg)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := conn.Close(); err != nil {
		return nil, err
	}

	log.Printf("database initialization configured for engine %s", cfg.Engine)
	return ctx, nil
}

func resolveDatabaseConfigFromEnvironment() (db.Config, bool, error) {
	engineName := strings.TrimSpace(os.Getenv("TOGGLY_DB_ENGINE"))
	if engineName == "" {
		log.Println("TOGGLY_DB_ENGINE not set; skipping database initialization")
		return db.Config{}, false, nil
	}

	dsn, dsnAvailable := os.LookupEnv("TOGGLY_DB_DSN")
	if dsnAvailable && strings.TrimSpace(dsn) != "" {
		return db.Config{Engine: db.EngineKind(engineName), DSN: dsn}, true, nil
	}

	log.Println("TOGGLY_DB_DSN not set; checking explicit configuration for database connection")

	credentialMap := map[string]string{
		"TOGGLY_DB_USER":     os.Getenv("TOGGLY_DB_USER"),
		"TOGGLY_DB_PASSWORD": os.Getenv("TOGGLY_DB_PASSWORD"),
		"TOGGLY_DB_HOST":     os.Getenv("TOGGLY_DB_HOST"),
		"TOGGLY_DB_PORT":     os.Getenv("TOGGLY_DB_PORT"),
		"TOGGLY_DB_NAME":     os.Getenv("TOGGLY_DB_NAME"),
	}

	missingVars := []string{}
	for envVar, value := range credentialMap {
		if strings.TrimSpace(value) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		log.Printf("missing required database configuration: %s", strings.Join(missingVars, ", "))
		return db.Config{}, false, nil
	}

	return db.Config{Engine: db.EngineKind(engineName), DSN: constructDSN(engineName, credentialMap)}, true, nil
}

func constructDSN(engineName string, creds map[string]string) string {
	switch strings.ToLower(engineName) {
	case "mysql":
		return creds["TOGGLY_DB_USER"] + ":" + creds["TOGGLY_DB_PASSWORD"] + "@tcp(" + creds["TOGGLY_DB_HOST"] + ":" + creds["TOGGLY_DB_PORT"] + ")/" + creds["TOGGLY_DB_NAME"]
	case "postgresql", "postgres":
		return "postgres://" + creds["TOGGLY_DB_USER"] + ":" + creds["TOGGLY_DB_PASSWORD"] + "@" + creds["TOGGLY_DB_HOST"] + ":" + creds["TOGGLY_DB_PORT"] + "/" + creds["TOGGLY_DB_NAME"] + "?sslmode=disable"
	default:
		log.Printf("unsupported database engine: %s", engineName)
		return ""
	}
}
