package strategies

import (
	"testing"

	"github.com/google/uuid"
)

func TestConservativeBidStrategyDecideBid(t *testing.T) {
	auction := newTestAuction(t, 1000)
	bidderID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	strategy, err := NewConservativeBidStrategy(bidderID)
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
			conservativeMinMarkupPercent,
			conservativeMaxMarkupPercent,
		)
	}
}
