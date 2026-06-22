package strategies

import (
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

type FakeBidStrategy struct {
	bidderID uuid.UUID
}

func (s *FakeBidStrategy) DecideBid(auction domain.Auction) (domain.BidProposal, error) {
	return domain.BidProposal{
		AuctionID: auction.ID(),
		BidderID:  s.bidderID,
		Amount:    auction.ReservePrice(),
	}, nil
}

func NewFakeBidStrategy(bidderID uuid.UUID) *FakeBidStrategy {
	return &FakeBidStrategy{
		bidderID: bidderID,
	}
}
