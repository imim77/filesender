package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

func run(ctx context.Context, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	externalIceServers, err := parseExternalIceServers(os.Getenv("EXTERNAL_ICE_SERVERS_JSON"))
	if err != nil {
		return fmt.Errorf("failed to parse ICE servers")
	}
	cfg := Config{
		Host:               os.Getenv("HOST"),
		Port:               os.Getenv("PORT"),
		ExternalIceServers: externalIceServers,
	}
	core := NewCore()
	go core.run()
	wsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(core, w, r)
	})

	srv := NewServer(cfg, wsHandler)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		Handler: srv,
	}

	fmt.Fprintf(stdout, "Signaling server starting on PORT :%s\n", cfg.Port)
	go func() {
		log.Printf("Listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // waitgroup.go??? mozda im
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()

	return nil

}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
