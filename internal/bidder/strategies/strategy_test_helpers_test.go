package strategies

import (
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

func newTestAuction(t *testing.T, reservePrice int64) domain.Auction {
	t.Helper()

	auction, err := domain.NewAuction(
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		uuid.MustParse("1e234536-e29b-41d4-a716-446655440000"),
		reservePrice,
		time.Now(),
		time.Now().Add(1*time.Hour),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	return auction
}

func assertBidProposalInMarkupRange(
	t *testing.T,
	proposal domain.BidProposal,
	auction domain.Auction,
	bidderID uuid.UUID,
	minMarkupPercent int64,
	maxMarkupPercent int64,
) {
	t.Helper()

	minAmount := auction.ReservePrice() + auction.ReservePrice()*minMarkupPercent/100
	maxAmount := auction.ReservePrice() + auction.ReservePrice()*maxMarkupPercent/100

	if proposal.AuctionID != auction.ID() {
		t.Fatalf("expected auction ID %s, got %s", auction.ID(), proposal.AuctionID)
	}

	if proposal.BidderID != bidderID {
		t.Fatalf("expected bidder ID %s, got %s", bidderID, proposal.BidderID)
	}

	if proposal.Amount < minAmount || proposal.Amount > maxAmount {
		t.Fatalf("expected amount between %d and %d, got %d", minAmount, maxAmount, proposal.Amount)
	}
}
