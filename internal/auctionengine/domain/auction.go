package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuctionStatus string

const (
	StatusOpen      AuctionStatus = "OPEN"
	StatusClosed    AuctionStatus = "CLOSED"
	AuctionDuration               = 100 * time.Millisecond
)

var ErrInvalidReservePrice = errors.New("reserve price should be > 0")
var ErrInvalidItemID = errors.New("itemID is nil")
var ErrAuctionAlreadyClosed = errors.New("auction is already closed")

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int64
	startAt      time.Time
	endAt        time.Time
	status       AuctionStatus
}

func NewAuction(itemID uuid.UUID, reservePrice int64) (*Auction, error) {
	if reservePrice <= 0 {
		return nil, ErrInvalidReservePrice
	}

	if itemID == uuid.Nil {
		return nil, ErrInvalidItemID
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

func (a *Auction) Close() error {
	if a.status == StatusClosed {
		return ErrAuctionAlreadyClosed
	}

	a.status = StatusClosed
	return nil
}
