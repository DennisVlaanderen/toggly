package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"toggly/backend/internal/api"
	"toggly/backend/internal/auth"
	"toggly/backend/internal/fqdp"
	"toggly/backend/internal/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	flagStore, err := store.Open(storeConfigFromEnvironment())
	if err != nil {
		log.Fatalf("failed to open flag store: %v", err)
	}
	defer flagStore.Close()

	go func() {
		if err := fqdp.StartTCPServer(ctx, ":9000"); err != nil {
			log.Printf("fqdp server stopped: %v", err)
		}
	}()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, flagStore)

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

	if err := auth.SeedAdminGroupAndUser(flagStore, adminConfigFromEnvironment()); err != nil {
		log.Fatalf("failed to seed admin account: %v", err)
	}

	<-ctx.Done()
	log.Println("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
}

func storeConfigFromEnvironment() store.Config {
	nodeID := strings.TrimSpace(os.Getenv("TOGGLY_RAFT_NODE_ID"))
	if nodeID == "" {
		nodeID = "node1"
	}

	bindAddr := strings.TrimSpace(os.Getenv("TOGGLY_RAFT_BIND_ADDR"))
	if bindAddr == "" {
		// A wildcard address (":9100") isn't advertisable to raft peers on
		// some hosts (observed on Windows outside Docker); loopback is a
		// safe default for a single-node local run.
		bindAddr = "127.0.0.1:9100"
	}

	dataDir := strings.TrimSpace(os.Getenv("TOGGLY_RAFT_DATA_DIR"))
	if dataDir == "" {
		dataDir = "./data"
	}

	bootstrap := true
	if raw := strings.TrimSpace(os.Getenv("TOGGLY_RAFT_BOOTSTRAP")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			log.Fatalf("invalid TOGGLY_RAFT_BOOTSTRAP value %q: %v", raw, err)
		}
		bootstrap = parsed
	}

	return store.Config{
		NodeID:    nodeID,
		BindAddr:  bindAddr,
		DataDir:   dataDir,
		Bootstrap: bootstrap,
	}
}

func adminConfigFromEnvironment() auth.AdminConfig {
	defaults := auth.DefaultAdminConfig()

	username := strings.TrimSpace(os.Getenv("TOGGLY_ADMIN_USERNAME"))
	if username == "" {
		username = defaults.Username
	}

	password := os.Getenv("TOGGLY_ADMIN_PASSWORD")
	if strings.TrimSpace(password) == "" {
		if isProductionEnvironment() {
			log.Fatal("TOGGLY_ADMIN_PASSWORD must be set when TOGGLY_ENV=production")
		}
		log.Println("TOGGLY_ADMIN_PASSWORD not set; using insecure development default")
		password = defaults.Password
	}

	return auth.AdminConfig{Username: username, Password: password}
}

// isProductionEnvironment reports whether TOGGLY_ENV is set to
// "production" -- the switch that turns insecure-default fallbacks (JWT
// secret, admin password) into hard startup failures instead of warnings.
// Left unset, behavior is unchanged from before this flag existed.
func isProductionEnvironment() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("TOGGLY_ENV")), "production")
}
