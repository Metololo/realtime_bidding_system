package domain

import "github.com/google/uuid"

type Bid struct {
	auctionID uuid.UUID
	bidderID  uuid.UUID
	amount    int64
}

func NewBid(auctionID uuid.UUID, bidderID uuid.UUID, amount int64) (*Bid, error) {
	return &Bid{
		auctionID: auctionID,
		bidderID:  bidderID,
		amount:    amount,
	}, nil
}
