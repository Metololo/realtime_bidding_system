package strategies

import (
	"testing"

	"github.com/google/uuid"
)

func TestBalancedBidStrategyDecideBid(t *testing.T) {
	auction := newTestAuction(t, 1000)
	bidderID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	strategy, err := NewBalancedBidStrategy(bidderID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for range 100 {
		proposal, err := strategy.DecideBid(auction)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertBidProposalInMarkupRange(
			t,
			proposal,
			auction,
			bidderID,
			balancedMinMarkupPercent,
			balancedMaxMarkupPercent,
		)
	}
}
