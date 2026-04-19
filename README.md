# Realtime Bidding System

This project is a simulation of a real-time auction platform built in Go.

It simulates short-lived auctions where bidders compete under strict latency constraints. The goal of this project is to explores differents programming concepts like concurrency, network latency, clean architecture.

The theme is inspired by **One Piece**: Devil Fruits are auctioned, and bidders such as Pirate Crews, Marines, and Revolutionaries compete using berries.

## Goals

The system is designed around the following **target** constraints:

- auctions are time-bounded: **<= 100 ms**
- bids arriving too late must be rejected
- if the system is overloaded, bids must be rejected fast
- target throughput: **1000 active auctions per second**
- bids should arrive within **20 ms** of auction start

This is a simulation project designed for experimentation and performance measurement.

## High-Level Architecture

The project is split into several components:

### 1. Auction Generator
Creates new auctions and sends them to the auction engine over HTTP.

Responsibilities:
- generate Devil Fruit auctions
- control auction creation rate

### 2. Auction Engine
The core service of the project.

Responsibilities:
- receive and start auctions
- keep active auction state in memory
- notify bidder simulator about newly opened auctions
- receive bids over gRPC
- enforce timing and overload rules
- determine winners
- persist closed auctions and replayable events to PostgreSQL

### 3. Bidder Simulator
Simulates multiple bidders with different strategies.

Responsibilities:
- receive newly started auctions from the engine
- create bidders from different factions
- submit bids over gRPC
- react to auction outcomes and rejection reasons

### 4. Replay Web App
A Next.js frontend that replays completed auctions from stored events.

Responsibilities:
- show auction timelines
- replay bids in browser
- display scores and faction statistics
- visualize winners, berries spent, and Devil Fruits won

## Hexagonal Architecture

The auction engine has a lot of business logic and use several technical boundaries (gRCP, HTTP, Postgre, In-memory state). Also, it will be useful to test the application layer and make sure that the core logic is robust.

It helps separate:

- **domain logic**: auction rules, highest bid selection, deadline handling
- **application logic**: starting auctions, submitting bids, closing auctions
- **adapters**: HTTP, gRPC, Postgres, in-memory state, event persistence