package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewBid(t *testing.T) {

	auctionID := uuid.MustParse("123e4566-e29b-41d4-a716-446655440000")
	bidderID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	amount := int64(100)

	bid, err := NewBid(auctionID, bidderID, amount)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bid == nil {
		t.Fatal("bid id nil")
	}

	if bid.auctionID != auctionID {
		t.Fatalf("expected auctionId to be %v, got %v", auctionID, bid.auctionID)
	}

	if bid.bidderID != bidderID {
		t.Fatalf("expected bidderID to be %v, got %v", bidderID, bid.bidderID)
	}

	if bid.amount != amount {
		t.Fatalf("expected amount to be %v, got %v", amount, bid.amount)
	}
}
