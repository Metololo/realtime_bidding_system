package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrNilAuctionID = errors.New("auctionID is nil")
var ErrNilBidderID = errors.New("bidderID is nil")
var ErrInvalidBidAmount = errors.New("amount should be > 0")

type Bid struct {
	id        uuid.UUID
	auctionID uuid.UUID
	bidderID  uuid.UUID
	amount    int64
}

func NewBid(auctionID uuid.UUID, bidderID uuid.UUID, amount int64) (*Bid, error) {

	if auctionID == uuid.Nil {
		return nil, ErrNilAuctionID
	}
	if bidderID == uuid.Nil {
		return nil, ErrNilBidderID
	}
	if amount <= 0 {
		return nil, ErrInvalidBidAmount
	}

	return &Bid{
		id:        uuid.New(),
		auctionID: auctionID,
		bidderID:  bidderID,
		amount:    amount,
	}, nil
}
