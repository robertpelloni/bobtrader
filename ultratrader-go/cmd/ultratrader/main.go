package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/app"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func main() {
	configPath := flag.String("config", "", "optional path to ultratrader config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := application.Start(ctx); err != nil {
		log.Fatalf("start app: %v", err)
	}

	addr := application.Address()
	fmt.Printf("UltraTrader Go running. event_log=%s accounts=%d address=%s\n", cfg.EventLog.Path, len(cfg.Accounts), addr)
	if addr != "" {
		fmt.Printf("Dashboard: http://%s/\n", addr)
	}

	// Block until signal received
	<-ctx.Done()
	fmt.Println("\nShutting down...")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := application.Shutdown(shutCtx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
	}
}
