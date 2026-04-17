package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewAuctionGivenARequest(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction := NewAuction(itemID, reservePrice)

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
	auction := NewAuction(itemID, reservePrice)

	if auction.status != StatusOpen {
		t.Fatal("expected auction status to be OPEN")
	}
}

func TestNewAuctionSetsEndAtAfterConfiguredDuration(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction := NewAuction(itemID, reservePrice)

	auctionDuration := auction.endAt.Sub(auction.startAt)

	if auctionDuration != AuctionDuration {
		t.Fatalf("expected auction duration to be %s, got %s", AuctionDuration, auctionDuration)
	}
}

func newTestAuctionRequest() (uuid.UUID, int) {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), 150
}
