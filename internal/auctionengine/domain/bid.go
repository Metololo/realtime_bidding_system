package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrNilBidderID = errors.New("bidderID is nil")
var ErrInvalidBidAmount = errors.New("amount should be > 0")

type Bid struct {
	bidderID uuid.UUID
	amount   int64
}

func (b *Bid) BidderID() uuid.UUID {
	return b.bidderID
}
func (b *Bid) Amount() int64 {
	return b.amount
}

func NewBid(bidderID uuid.UUID, amount int64) (*Bid, error) {

	if bidderID == uuid.Nil {
		return nil, ErrNilBidderID
	}
	if amount <= 0 {
		return nil, ErrInvalidBidAmount
	}

	return &Bid{
		bidderID: bidderID,
		amount:   amount,
	}, nil
}
