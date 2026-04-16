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

func NewAuction(itemID uuid.UUID, reservePrice int) *Auction {
	const auctionDuration = 100 * time.Millisecond
	startAt := time.Now()
	endAt := startAt.Add(auctionDuration)

	return &Auction{
		id:           uuid.New(),
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
		status:       StatusOpen,
	}
}
