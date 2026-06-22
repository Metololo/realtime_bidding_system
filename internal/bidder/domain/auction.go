package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidReservePrice = errors.New("reserve price should be > 0")
var ErrNilItemID = errors.New("itemID is nil")

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int64
	startAt      time.Time
	endAt        time.Time
}

func (a *Auction) ID() uuid.UUID {
	return a.id
}

func (a *Auction) ItemID() uuid.UUID {
	return a.itemID
}

func (a *Auction) ReservePrice() int64 {
	return a.reservePrice
}

func NewAuction(
	id uuid.UUID,
	itemID uuid.UUID,
	reservePrice int64,
	startAt time.Time,
	endAt time.Time,
) (Auction, error) {
	if id == uuid.Nil {
		return Auction{}, errors.New("auction id is nil")
	}

	if itemID == uuid.Nil {
		return Auction{}, ErrNilItemID
	}

	if reservePrice <= 0 {
		return Auction{}, ErrInvalidReservePrice
	}

	if !endAt.After(startAt) {
		return Auction{}, errors.New("auction endAt should be after startAt")
	}

	return Auction{
		id:           id,
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
	}, nil
}
