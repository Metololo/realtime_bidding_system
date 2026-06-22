package strategies

import (
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

const (
	balancedMinMarkupPercent int64 = 10
	balancedMaxMarkupPercent int64 = 65
)

type BalancedBidStrategy struct {
	randomMarkupStrategy
}

func NewBalancedBidStrategy(bidderID uuid.UUID) (*BalancedBidStrategy, error) {
	strategy, err := newRandomMarkupStrategy(
		bidderID,
		balancedMinMarkupPercent,
		balancedMaxMarkupPercent,
	)
	if err != nil {
		return nil, err
	}

	return &BalancedBidStrategy{randomMarkupStrategy: strategy}, nil
}

func (s *BalancedBidStrategy) DecideBid(auction domain.Auction) (domain.BidProposal, error) {
	return s.decideBid(auction)
}
