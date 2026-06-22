package strategies

import (
	"errors"
	"math/rand"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

var ErrInvalidMarkupRange = errors.New("invalid markup range")

type randomMarkupStrategy struct {
	bidderID         uuid.UUID
	minMarkupPercent int64
	maxMarkupPercent int64
}

func newRandomMarkupStrategy(bidderID uuid.UUID, minMarkupPercent, maxMarkupPercent int64) (randomMarkupStrategy, error) {
	if minMarkupPercent < 0 || maxMarkupPercent < minMarkupPercent {
		return randomMarkupStrategy{}, ErrInvalidMarkupRange
	}

	return randomMarkupStrategy{
		bidderID:         bidderID,
		minMarkupPercent: minMarkupPercent,
		maxMarkupPercent: maxMarkupPercent,
	}, nil
}

func (s randomMarkupStrategy) decideBid(auction domain.Auction) (domain.BidProposal, error) {
	markupPercent := s.minMarkupPercent
	if s.maxMarkupPercent > s.minMarkupPercent {
		markupPercent += rand.Int63n(s.maxMarkupPercent - s.minMarkupPercent + 1)
	}

	amount := auction.ReservePrice() + auction.ReservePrice()*markupPercent/100

	proposal, err := domain.NewBidProposal(auction.ID(), s.bidderID, amount)
	if err != nil {
		return domain.BidProposal{}, err
	}

	return *proposal, nil
}
