package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventAuctionCreated EventType = "auction.created"
	EventAuctionClosed  EventType = "auction.closed"
)

type Event interface {
	EventID() uuid.UUID
	EventType() EventType
	OccurredAt() time.Time
}

type BaseEvent struct {
	ID uuid.UUID
	At time.Time
}

func (e BaseEvent) EventID() uuid.UUID    { return e.ID }
func (e BaseEvent) OccurredAt() time.Time { return e.At }

type AuctionCreatedEvent struct {
	BaseEvent
	AuctionID    uuid.UUID
	ItemID       uuid.UUID
	ReservePrice int64
	StartedAt    time.Time
	EndAt        time.Time
}

func (AuctionCreatedEvent) EventType() EventType {
	return EventAuctionCreated
}

type AuctionOutcome string

const (
	AuctionOutcomeSold   AuctionOutcome = "sold"
	AuctionOutcomeNoBids AuctionOutcome = "no_bids"
)

type AuctionClosedEvent struct {
	BaseEvent
	AuctionID uuid.UUID
	ItemID    uuid.UUID
	Outcome   AuctionOutcome
	ClosedAt  time.Time
	Winner    *WinnerInfo
}

func (AuctionClosedEvent) EventType() EventType {
	return EventAuctionClosed

}

type WinnerInfo struct {
	BidderID uuid.UUID
	Amount   int64
}
