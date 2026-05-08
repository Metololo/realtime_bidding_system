package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/seller-service/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/seller-service/repo"
	"github.com/Metololo/realtime_bidding_system/internal/seller-service/service"
)

func main() {
	slog.SetDefault(infrastructure.NewSellerServiceLogger(slog.LevelInfo))

	auctionEngineURL := env("AUCTION_ENGINE_URL", "http://localhost:8080")
	interval, err := intervalFromEnv()
	if err != nil {
		slog.Error("invalid SELLER_INTERVAL", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	devilFruitRepository := repo.NewInMemoryDevilFruitRepository()
	sellService := service.NewSellService(devilFruitRepository)
	auctionEngineClient := infrastructure.NewAuctionEngineClient(auctionEngineURL)

	slog.Info("seller service started",
		slog.String("auction_engine_url", auctionEngineURL),
		slog.Duration("interval", interval),
	)

	sellService.Run(ctx, auctionEngineClient, interval)
	slog.Info("seller service stopped")
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func intervalFromEnv() (time.Duration, error) {
	interval, err := time.ParseDuration(env("SELLER_INTERVAL", "10s"))
	if err != nil {
		return 0, err
	}
	if interval <= 0 {
		return 0, fmt.Errorf("must be greater than zero")
	}
	return interval, nil
}
