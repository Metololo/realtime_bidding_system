package strategies

import (
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/bidder/domain"
	"github.com/google/uuid"
)

func TestDecideBid(t *testing.T) {

	auction, err := domain.NewAuction(
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		uuid.MustParse("1e234536-e29b-41d4-a716-446655440000"),
		int64(100),
		time.Now(),
		time.Now().Add(1*time.Hour),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	strategy := NewFakeBidStrategy(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"))

	bidProposal, err := strategy.DecideBid(auction)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bidProposal == (domain.BidProposal{}) {
		t.Fatalf("expected a bid proposal, got nil")
	}

	if bidProposal.AuctionID != auction.ID() {
		t.Fatalf("expected auction ID %s, got %s", auction.ID(), bidProposal.AuctionID)
	}

	if bidProposal.BidderID == uuid.Nil {
		t.Fatalf("expected a valid bidder ID, got nil")
	}

	if bidProposal.Amount <= 0 {
		t.Fatalf("expected a positive bid amount, got %d", bidProposal.Amount)
	}

	if bidProposal.Amount != auction.ReservePrice() {
		t.Fatalf("expected bid amount to be equal to reserve price %d, got %d", auction.ReservePrice(), bidProposal.Amount)
	}

}
