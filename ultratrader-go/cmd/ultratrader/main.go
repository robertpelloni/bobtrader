package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	fmt.Printf("UltraTrader Go scaffold initialized. event_log=%s accounts=%d\n", cfg.EventLog.Path, len(cfg.Accounts))
}
