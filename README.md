# Realtime Bidding System

This project is a simulation of a real-time auction platform built in Go.

It simulates short-lived auctions where bidders compete under strict latency constraints. The goal of the project is to explore concurrency, low-latency service design, messaging, RPC, backpressure, and clean boundaries.

The theme is inspired by One Piece: Devil Fruits are auctioned, and bidders such as Pirate Crews, Marines, and Revolutionaries compete using berries.

## Goals

The system is designed around the following target constraints:

- auctions are time-bounded: <= 100 ms
- bids arriving too late must be rejected
- if the system is overloaded, bids must be rejected fast
- target throughput: 1000 active auctions per second
- bids should arrive within 20 ms of auction start

This is a simulation project designed for experimentation, benchmarking, and architecture learning.

## High-Level Architecture

The project is split into several components:

### 1. Auction Generator
Creates new auctions and sends them to the auction engine.

Responsibilities:
- generate Devil Fruit auctions
- control auction creation rate
- call the auction engine synchronously to create auctions

Transport:
- gRPC from generator to auction-engine

Why gRPC here:
- auction creation is a command, not an event
- the generator benefits from an immediate response: accepted, rejected, invalid, overloaded
- it is a good place to learn protobuf + gRPC without forcing everything into RPC

### 2. Auction Engine
The core service of the project and the source of truth for auction state.

Responsibilities:
- receive and start auctions
- keep active auction state in memory
- publish newly opened auctions to NATS
- receive bid submissions over gRPC
- enforce timing and overload rules
- determine winners
- publish authoritative auction-domain events
- persist outcomes and optional detailed traces

Important ownership rule:
- the engine owns the truth
- generator and bidders send commands
- the engine publishes facts/events

### 3. Bidder Service
A bidder worker service deployed with multiple replicas.

Each replica represents one simulated bidder profile. The service binary stays the same, but each replica can load a different YAML configuration describing its strategy and identity.

Responsibilities:
- subscribe to auction.started events from NATS
- decide whether and how to bid according to its strategy
- call SubmitBid on the auction-engine over gRPC
- optionally consume result events if future strategies need feedback

Why this model is useful:
- easy horizontal scaling
- realistic distributed behavior
- easy to compare bidder strategies
- clean separation between event consumption and bid submission

Configuration:
- YAML is preferred for bidder strategy/profile config
- env vars can still be used for wiring, instance identity, or config path

### 4. Metrics / Analytics
The system should support both operational metrics and large-volume event analysis.

The project does not need a custom metrics service just because it generates many events.

Recommended approach:
- Prometheus + Grafana for live operational dashboards and standard metrics
- ClickHouse for raw event storage and ad hoc analysis
- optional ingestion tool such as Vector or Benthos to ship events into ClickHouse

This is a better fit than trying to replay every raw event inside the main app.

## Transport Strategy: NATS + gRPC

This project uses a hybrid transport model on purpose.

Use gRPC for commands:
- auction-generator -> auction-engine: CreateAuction
- bidder -> auction-engine: SubmitBid

Use NATS for fan-out events:
- auction.started
- auction.closed
- optional bid.accepted
- optional bid.rejected
- optional bidder.outbid / leader.changed

Why this split works well:
- gRPC is a good fit for direct command/response interactions
- NATS is a good fit for fan-out and event-driven reactions
- it lets the project teach both RPC and messaging with clear responsibility boundaries

## Why bids can still be accepted and later lose

A submitted bid being accepted does not mean the bidder is guaranteed to win.

It only means the bid was admitted into the auction process:
- the auction existed
- it was still open
- the bid met validation rules
- the system was not overloaded

A later higher bid may still replace it before the auction closes.

So the semantics are:
- gRPC SubmitBid = command admission result
- NATS auction/result events = evolving state and final outcome

## Event Ownership

Auction-domain events should be published by the auction-engine, not by generators or bidders.

Generator and bidder responsibilities:
- send commands
- expose their own local logs / technical telemetry if needed
- do not publish authoritative domain facts

Engine responsibilities:
- publish auction.started
- publish auction.closed
- publish rejection or outbid events if needed
- publish the final winner

This keeps one clear source of truth for analytics and debugging.

## Metrics Strategy

For the expected load, event volume is manageable.

Example:
- 1 minute run
- 1000 auctions created per second
- 5 bidders
- simple model: 1 auction.started + 5 bid submissions + 1 auction.closed per auction
- total: 420000 events

That volume is fine on a normal development machine and very manageable for ClickHouse.

### Metrics that should be easy to answer

From engine-owned events and metrics, the system should support:
- bidder that won the most auctions
- rejection counts by bidder and by reason
- total auctions created / processed
- average bid processing latency
- average per-bidder request-to-processing latency
- current active auctions
- maximum concurrent auctions

### Prometheus / Grafana

Prometheus is a good fit for pre-aggregated operational metrics such as:
- auctions_created_total
- auctions_closed_total
- auctions_won_total{bidder_id}
- bids_rejected_total{bidder_id, reason}
- bid_processing_seconds bucket histograms
- active_auctions
- active_auctions_max

Grafana can visualize these directly.

Prometheus is not the right place to store and query all raw auction events forever.

### ClickHouse

ClickHouse is the recommended store for raw event analysis because it is well suited for:
- querying hundreds of thousands or millions of events
- filtering by bidder, auction, strategy, rejection reason, or run
- drilling into sampled auctions for debugging
- building deeper analytics without writing a lot of custom aggregation code

A good practical setup is:
- engine publishes authoritative events
- Prometheus scrapes live metrics from services
- raw events are shipped to ClickHouse
- Grafana queries both Prometheus and ClickHouse

## Replay / Inspection App

A replay UI can still make sense, but it should not try to fully replay every event from every run by default..

A better direction is:
- show aggregate metrics and benchmark summaries
- allow drill-down into selected auctions
- store detailed traces for sampled, failed, or debug runs
- compare bidder strategies across runs

So the frontend is better framed as an analytics / inspection app with optional sampled replay.

## Architecture Style

Hexagonal architecture still makes sense here, but mainly because of boundaries, not because the domain is extremely complex.

A more accurate framing is:
- there are important auction rules, but they are relatively compact
- much of the complexity comes from timing, messaging, transport, in-memory coordination, and persistence boundaries
- hexagonal architecture helps keep those concerns separated from the core auction rules

It helps separate:
- domain logic: auction rules, deadlines, winner selection
- application logic: create auction, submit bid, close auction, publish events
- adapters: gRPC server, NATS publisher/subscriber, PostgreSQL, ClickHouse pipeline, in-memory store, metrics

## Suggested Flow

1. auction-generator -> gRPC -> auction-engine: CreateAuction
2. auction-engine stores active auction in memory
3. auction-engine publishes auction.started on NATS
4. bidder-service replicas receive the event and evaluate their YAML-defined strategy
5. bidder-service -> gRPC -> auction-engine: SubmitBid
6. auction-engine validates the bid and updates auction state
7. auction-engine publishes authoritative events such as auction.closed and optional rejection/outbid events
8. Prometheus scrapes live metrics, and raw events can be shipped to ClickHouse for analysis

## Practical Direction

If the goal is to build a strong low-latency distributed-systems demo, the most coherent version of the project is probably:

- gRPC for command-style interactions
- NATS for auction event fan-out
- one bidder service with many replicas and YAML strategy configs
- auction-engine as the sole publisher of domain events
- Prometheus + Grafana for live metrics
- ClickHouse for raw event analytics
- analytics-focused UI instead of replaying every raw event

That keeps the project realistic, scalable, educational, and easy to explain in an interview.
