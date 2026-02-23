package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/joho/godotenv"
)

func run(ctx context.Context, stdout, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	if err := godotenv.Load("server/.env"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load server/.env file: %w", err)
	}
	if err := godotenv.Load(".env"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	relayPortMin, _ := strconv.ParseUint(os.Getenv("RELAYPORTMIN"), 10, 16)
	relayPortMax, _ := strconv.ParseUint(os.Getenv("RELAYPORTMAX"), 10, 16)

	cfg := Config{
		Host:         os.Getenv("HOST"),
		Port:         os.Getenv("PORT"),
		TURNPort:     os.Getenv("TURNPORT"),
		TURNRealm:    os.Getenv("TURNREALM"),
		TURNSecret:   os.Getenv("TURNSECRET"),
		PublicHost:   os.Getenv("PUBLICHOST"),
		PublicIp:     os.Getenv("PUBLICIP"),
		RelayPortMin: uint16(relayPortMin),
		RelayPortMax: uint16(relayPortMax),
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
