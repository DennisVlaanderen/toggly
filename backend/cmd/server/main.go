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
		bindAddr = ":9100"
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
