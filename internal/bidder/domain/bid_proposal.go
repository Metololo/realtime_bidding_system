package domain

import (
	"errors"

	"github.com/google/uuid"
)

type BidProposal struct {
	AuctionID uuid.UUID
	BidderID  uuid.UUID
	Amount    int64
}

var ErrNilAuctionID = errors.New("auctionID is nil")
var ErrNilBidderID = errors.New("bidderID is nil")
var ErrInvalidBidAmount = errors.New("amount should be > 0")

func NewBidProposal(auctionID, bidderID uuid.UUID, amount int64) (*BidProposal, error) {
	if auctionID == uuid.Nil {
		return nil, ErrNilAuctionID
	}

	if bidderID == uuid.Nil {
		return nil, ErrNilBidderID
	}

	if amount <= 0 {
		return nil, ErrInvalidBidAmount
	}

	return &BidProposal{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    amount,
	}, nil
}
