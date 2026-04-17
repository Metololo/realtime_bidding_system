package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuctionStatus string

const (
	StatusOpen AuctionStatus = "OPEN"
)

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int
	startAt      time.Time
	endAt        time.Time
	status       AuctionStatus
}

const AuctionDuration = 100 * time.Millisecond

func NewAuction(itemID uuid.UUID, reservePrice int) *Auction {
	startAt := time.Now()
	endAt := startAt.Add(AuctionDuration)

	return &Auction{
		id:           uuid.New(),
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
		status:       StatusOpen,
	}
}
