package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"

	"github.com/peacock0803sz/vibra/back/gen/vibra/agent/v1/agentv1connect"
	"github.com/peacock0803sz/vibra/back/internal/auth"
	"github.com/peacock0803sz/vibra/back/internal/container"
	"github.com/peacock0803sz/vibra/back/internal/sandbox"
	"github.com/peacock0803sz/vibra/back/internal/server"
)

// ビルド時に -ldflags で注入されるバージョン情報
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("vibra %s (commit %s, built %s)\n", version, commit, date)
		os.Exit(0)
	}

	listenAddr := envOr("VIBRA_LISTEN_ADDR", "127.0.0.1:3001")

	// Initialize container runtime.
	dockerRT, err := container.NewDockerRuntime()
	if err != nil {
		log.Fatalf("container runtime: %v", err)
	}
	runner := container.NewRunner(dockerRT)

	// Load sandbox configuration from environment.
	sandboxCfg := sandbox.NewConfigFromEnv()
	if len(sandboxCfg.AllowedDirs) == 0 {
		log.Fatal("VIBRA_ALLOWED_DIRS is required (comma-separated list of allowed working directories)")
	}

	// Auth interceptor (Tailscale headers + Bearer token fallback).
	authInterceptor := auth.NewInterceptor(&auth.InterceptorConfig{
		DevUser: os.Getenv("VIBRA_DEV_USER"),
	})

	// Default working directory for requests that omit it (useful for local dev).
	defaultWorkDir := os.Getenv("VIBRA_DEFAULT_WORKDIR")

	// Register ConnectRPC handlers.
	mux := http.NewServeMux()
	path, handler := agentv1connect.NewAgentServiceHandler(
		server.NewAgentServer(runner, sandboxCfg, defaultWorkDir),
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(path, handler)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{envOr("VIBRA_CORS_ORIGIN", "http://127.0.0.1:5173")},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           corsHandler.Handler(mux),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		// WriteTimeout is intentionally omitted for streaming RPCs.
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("vibra server listening on %s", listenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
