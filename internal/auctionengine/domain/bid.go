package domain

import (
	"errors"

	"github.com/google/uuid"
)

var ErrNilAuctionID = errors.New("auctionID is nil")
var ErrNilBidderID = errors.New("bidderID is nil")

type Bid struct {
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

	return &Bid{
		auctionID: auctionID,
		bidderID:  bidderID,
		amount:    amount,
	}, nil
}
