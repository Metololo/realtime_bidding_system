package strategies

import (
	"testing"

	"github.com/google/uuid"
)

func TestAggressiveBidStrategyDecideBid(t *testing.T) {
	auction := newTestAuction(t, 1000)
	bidderID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	strategy, err := NewAggressiveBidStrategy(bidderID)
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
			aggressiveMinMarkupPercent,
			aggressiveMaxMarkupPercent,
		)
	}
}
