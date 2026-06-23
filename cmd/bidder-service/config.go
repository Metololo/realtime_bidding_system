package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Bidder        BidderConfig        `yaml:"bidder"`
	NATS          NATSConfig          `yaml:"nats"`
	AuctionEngine AuctionEngineConfig `yaml:"auction_engine"`
}

type BidderConfig struct {
	ID       string `yaml:"id"`
	Strategy string `yaml:"strategy"`
}

type NATSConfig struct {
	URL        string `yaml:"url"`
	QueueGroup string `yaml:"queue_group"`
}

type AuctionEngineConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
	Timeout  string `yaml:"timeout"`
}

type ResolvedConfig struct {
	BidderID           uuid.UUID
	Strategy           string
	NATSURL            string
	NATSQueueGroup     string
	AuctionGRPCAddr    string
	AuctionGRPCTimeout time.Duration
}

func loadConfig() (ResolvedConfig, error) {
	cfg := defaultConfig()

	if path := os.Getenv("CONFIG_PATH"); path != "" {
		loaded, err := loadYAMLConfig(path)
		if err != nil {
			return ResolvedConfig{}, err
		}
		cfg = mergeConfig(cfg, loaded)
	} else {
		cfg = mergeConfig(cfg, configFromEnv())
	}

	return resolveConfig(cfg)
}

func defaultConfig() Config {
	return Config{
		Bidder: BidderConfig{
			ID:       "22222222-2222-2222-2222-222222222222",
			Strategy: "balanced",
		},
		NATS: NATSConfig{
			URL: "nats://localhost:4222",
		},
		AuctionEngine: AuctionEngineConfig{
			GRPCAddr: "localhost:9001",
			Timeout:  "3s",
		},
	}
}

func loadYAMLConfig(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func configFromEnv() Config {
	return Config{
		Bidder: BidderConfig{
			ID:       os.Getenv("BIDDER_ID"),
			Strategy: os.Getenv("BIDDER_STRATEGY"),
		},
		NATS: NATSConfig{
			URL:        os.Getenv("NATS_SERVER_URL"),
			QueueGroup: os.Getenv("BIDDER_NATS_QUEUE_GROUP"),
		},
		AuctionEngine: AuctionEngineConfig{
			GRPCAddr: os.Getenv("AUCTION_ENGINE_GRPC_ADDR"),
			Timeout:  os.Getenv("AUCTION_ENGINE_GRPC_TIMEOUT"),
		},
	}
}

func mergeConfig(base, override Config) Config {
	if override.Bidder.ID != "" {
		base.Bidder.ID = override.Bidder.ID
	}
	if override.Bidder.Strategy != "" {
		base.Bidder.Strategy = override.Bidder.Strategy
	}
	if override.NATS.URL != "" {
		base.NATS.URL = override.NATS.URL
	}
	if override.NATS.QueueGroup != "" {
		base.NATS.QueueGroup = override.NATS.QueueGroup
	}
	if override.AuctionEngine.GRPCAddr != "" {
		base.AuctionEngine.GRPCAddr = override.AuctionEngine.GRPCAddr
	}
	if override.AuctionEngine.Timeout != "" {
		base.AuctionEngine.Timeout = override.AuctionEngine.Timeout
	}
	return base
}

func resolveConfig(cfg Config) (ResolvedConfig, error) {
	bidderID, err := uuid.Parse(cfg.Bidder.ID)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("bidder.id must be a valid UUID: %w", err)
	}
	if bidderID == uuid.Nil {
		return ResolvedConfig{}, fmt.Errorf("bidder.id must not be nil")
	}

	timeout, err := time.ParseDuration(cfg.AuctionEngine.Timeout)
	if err != nil {
		return ResolvedConfig{}, fmt.Errorf("auction_engine.timeout must be a valid duration: %w", err)
	}
	if timeout <= 0 {
		return ResolvedConfig{}, fmt.Errorf("auction_engine.timeout must be greater than zero")
	}

	return ResolvedConfig{
		BidderID:           bidderID,
		Strategy:           cfg.Bidder.Strategy,
		NATSURL:            cfg.NATS.URL,
		NATSQueueGroup:     cfg.NATS.QueueGroup,
		AuctionGRPCAddr:    cfg.AuctionEngine.GRPCAddr,
		AuctionGRPCTimeout: timeout,
	}, nil
}
