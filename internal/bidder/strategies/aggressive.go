package strategies

import (
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

const (
	aggressiveMinMarkupPercent int64 = 25
	aggressiveMaxMarkupPercent int64 = 90
)

type AggressiveBidStrategy struct {
	randomMarkupStrategy
}

func NewAggressiveBidStrategy(bidderID uuid.UUID) (*AggressiveBidStrategy, error) {
	strategy, err := newRandomMarkupStrategy(
		bidderID,
		aggressiveMinMarkupPercent,
		aggressiveMaxMarkupPercent,
	)
	if err != nil {
		return nil, err
	}

	return &AggressiveBidStrategy{randomMarkupStrategy: strategy}, nil
}

func (s *AggressiveBidStrategy) DecideBid(auction domain.Auction) (domain.BidProposal, error) {
	return s.decideBid(auction)
}
