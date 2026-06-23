package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/application"
	"github.com/Metololo/realtime_bidding_system/internal/bidder/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/bidder/strategies"
	"github.com/google/uuid"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid bidder config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	strategy, err := strategyFromName(cfg.Strategy, cfg.BidderID)
	if err != nil {
		slog.Error("invalid bidder strategy", slog.String("err", err.Error()))
		os.Exit(1)
	}

	auctionEngineClient, err := infrastructure.NewAuctionEngineGRPCClient(cfg.AuctionGRPCAddr)
	if err != nil {
		slog.Error("failed to create auction engine gRPC client", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := auctionEngineClient.Close(); err != nil {
			slog.Error("failed to close auction engine gRPC client", slog.String("err", err.Error()))
		}
	}()
	auctionEngineClient.SetTimeout(cfg.AuctionGRPCTimeout)

	bidPlacer := application.NewBidPlacer(strategy, auctionEngineClient)
	subscriber := infrastructure.NewNATSAuctionEventSubscriber(
		cfg.NATSURL,
		cfg.NATSQueueGroup,
		cfg.BidderID,
		bidPlacer,
		os.Stdout,
	)

	slog.Info("bidder service starting",
		slog.String("config_path", os.Getenv("CONFIG_PATH")),
		slog.String("bidder_id", cfg.BidderID.String()),
		slog.String("strategy", cfg.Strategy),
		slog.String("auction_engine_grpc_addr", cfg.AuctionGRPCAddr),
		slog.Duration("auction_engine_grpc_timeout", cfg.AuctionGRPCTimeout),
		slog.String("nats_url", cfg.NATSURL),
		slog.String("nats_queue_group", cfg.NATSQueueGroup),
	)

	if err := subscriber.Start(ctx); err != nil {
		slog.Error("bidder service stopped with error", slog.String("err", err.Error()))
		os.Exit(1)
	}

	slog.Info("bidder service stopped")
}

func strategyFromName(name string, bidderID uuid.UUID) (application.BidStrategy, error) {
	switch strings.ToLower(name) {
	case "aggressive":
		return strategies.NewAggressiveBidStrategy(bidderID)
	case "balanced":
		return strategies.NewBalancedBidStrategy(bidderID)
	case "conservative":
		return strategies.NewConservativeBidStrategy(bidderID)
	default:
		return nil, fmt.Errorf("must be one of aggressive, balanced, conservative")
	}
}
