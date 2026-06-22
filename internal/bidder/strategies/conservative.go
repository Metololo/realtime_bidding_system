package strategies

import (
	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

const (
	conservativeMinMarkupPercent int64 = 1
	conservativeMaxMarkupPercent int64 = 35
)

type ConservativeBidStrategy struct {
	randomMarkupStrategy
}

func NewConservativeBidStrategy(bidderID uuid.UUID) (*ConservativeBidStrategy, error) {
	strategy, err := newRandomMarkupStrategy(
		bidderID,
		conservativeMinMarkupPercent,
		conservativeMaxMarkupPercent,
	)
	if err != nil {
		return nil, err
	}

	return &ConservativeBidStrategy{randomMarkupStrategy: strategy}, nil
}

func (s *ConservativeBidStrategy) DecideBid(auction domain.Auction) (domain.BidProposal, error) {
	return s.decideBid(auction)
}
