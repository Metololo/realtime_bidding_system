package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuctionStatus string

const (
	StatusOpen      AuctionStatus = "OPEN"
	AuctionDuration               = 100 * time.Millisecond
)

var ErrNegativeReservePrice = errors.New("negative reserve price")

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int64
	startAt      time.Time
	endAt        time.Time
	status       AuctionStatus
}

func NewAuction(itemID uuid.UUID, reservePrice int64) (*Auction, error) {
	if reservePrice < 0 {
		return nil, ErrNegativeReservePrice
	}

	startAt := time.Now()
	endAt := startAt.Add(AuctionDuration)

	return &Auction{
		id:           uuid.New(),
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
		status:       StatusOpen,
	}, nil
}
