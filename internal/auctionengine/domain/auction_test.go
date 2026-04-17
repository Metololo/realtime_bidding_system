package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewAuctionGivenARequest(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got : %v", err)
	}

	if auction == nil {
		t.Fatal("Auction is nil")
	}

	if auction.itemID != itemID {
		t.Fatal("expected auction itemID to be set")
	}

	if auction.reservePrice != reservePrice {
		t.Fatal("expected auction reservePrice to match request")
	}

	if auction.status == "" {
		t.Fatal("expected auction status to be set")
	}

	if auction.id == uuid.Nil {
		t.Fatal("expected auction ID to be set")
	}

	if auction.startAt.IsZero() {
		t.Fatal("expected startAt to be set")
	}

	if auction.endAt.IsZero() {
		t.Fatal("expected endAt to be set")
	}
}

func TestNewAuctionSetsStatusOpen(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got : %v", err)
	}

	if auction.status != StatusOpen {
		t.Fatal("expected auction status to be OPEN")
	}
}

func TestNewAuctionSetsEndAtAfterConfiguredDuration(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got : %v", err)
	}

	auctionDuration := auction.endAt.Sub(auction.startAt)

	if auctionDuration != AuctionDuration {
		t.Fatalf("expected auction duration to be %s, got %s", AuctionDuration, auctionDuration)
	}
}

func TestNewAuctionReturnsErrorForNegativeReservePrice(t *testing.T) {
	itemID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reservePrice := int64(-1)

	auction, err := NewAuction(itemID, reservePrice)

	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, ErrNegativeReservePrice) {
		t.Fatalf("expected ErrNegativeReservePrice, got %v", err)
	}
}

func newTestAuctionRequest() (uuid.UUID, int64) {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), 150
}
