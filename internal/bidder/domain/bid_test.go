package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateBidProposal(t *testing.T) {

	auctionId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	bidderID := uuid.MustParse("1e234536-e29b-41d4-a716-446655440000")
	amount := int64(100)

	bidProposal, err := NewBidProposal(auctionId, bidderID, amount)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bidProposal == nil {
		t.Fatal("bidProposal is nil")
	}

	if bidProposal.AuctionID != auctionId {
		t.Fatalf("expected auctionID to be %v, got %v", auctionId, bidProposal.AuctionID)
	}

	if bidProposal.BidderID != bidderID {
		t.Fatalf("expected bidderID to be %v, got %v", bidderID, bidProposal.BidderID)
	}

	if bidProposal.Amount != amount {
		t.Fatalf("expected amount to be %v, got %v", amount, bidProposal.Amount)
	}
}
