package main

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLoadConfigFromYAML(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	path := writeTempConfig(t, `
bidder:
  id: "33333333-3333-3333-3333-333333333333"
  strategy: aggressive
nats:
  url: "nats://nats-server:4222"
  queue_group: ""
auction_engine:
  grpc_addr: "auction-engine:9001"
  timeout: 5s
`)
	t.Setenv("CONFIG_PATH", path)
	// These env vars must not override CONFIG_PATH-driven YAML config.
	t.Setenv("BIDDER_ID", "22222222-2222-2222-2222-222222222222")
	t.Setenv("BIDDER_STRATEGY", "balanced")

	cfg, err := loadConfig()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.BidderID != uuid.MustParse("33333333-3333-3333-3333-333333333333") {
		t.Fatalf("unexpected bidder ID %s", cfg.BidderID)
	}
	if cfg.Strategy != "aggressive" {
		t.Fatalf("expected aggressive strategy, got %s", cfg.Strategy)
	}
	if cfg.NATSURL != "nats://nats-server:4222" {
		t.Fatalf("unexpected NATS URL %s", cfg.NATSURL)
	}
	if cfg.AuctionGRPCAddr != "auction-engine:9001" {
		t.Fatalf("unexpected gRPC addr %s", cfg.AuctionGRPCAddr)
	}
	if cfg.AuctionGRPCTimeout != 5*time.Second {
		t.Fatalf("unexpected timeout %s", cfg.AuctionGRPCTimeout)
	}
}

func TestLoadConfigFromEnvWhenNoConfigPath(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	t.Setenv("BIDDER_ID", "11111111-1111-1111-1111-111111111111")
	t.Setenv("BIDDER_STRATEGY", "conservative")
	t.Setenv("NATS_SERVER_URL", "nats://example:4222")
	t.Setenv("BIDDER_NATS_QUEUE_GROUP", "")
	t.Setenv("AUCTION_ENGINE_GRPC_ADDR", "auction-engine:9001")
	t.Setenv("AUCTION_ENGINE_GRPC_TIMEOUT", "2s")

	cfg, err := loadConfig()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.BidderID != uuid.MustParse("11111111-1111-1111-1111-111111111111") {
		t.Fatalf("unexpected bidder ID %s", cfg.BidderID)
	}
	if cfg.Strategy != "conservative" {
		t.Fatalf("expected conservative strategy, got %s", cfg.Strategy)
	}
	if cfg.NATSURL != "nats://example:4222" {
		t.Fatalf("unexpected NATS URL %s", cfg.NATSURL)
	}
	if cfg.AuctionGRPCTimeout != 2*time.Second {
		t.Fatalf("unexpected timeout %s", cfg.AuctionGRPCTimeout)
	}
}

func TestLoadConfigRejectsInvalidStrategyAtFactory(t *testing.T) {
	_, err := strategyFromName("invalid", uuid.New())
	if err == nil {
		t.Fatal("expected invalid strategy error")
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "bidder-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close temp config: %v", err)
	}
	return file.Name()
}
