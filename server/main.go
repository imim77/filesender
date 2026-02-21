package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
)

func run(ctx context.Context, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	cfg := Config{
		Host:         "",
		Port:         "9000",
		TURNPort:     "3478",
		TURNRealm:    "justdrop",
		TURNSecret:   "change-me-in-production",
		PublicHost:   "",
		RelayPortMin: 49152,
		RelayPortMax: 65535,
	}

	srv, err := NewServer(&cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	fmt.Fprintf(stdout, "Signaling server starting on PORT :%s\n", cfg.Port)
	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	<-ctx.Done()
	return nil

}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
