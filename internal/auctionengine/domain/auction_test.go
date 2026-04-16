package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewAuctionGivenARequest(t *testing.T) {
	itemID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reservePrice := 150
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
		t.Fatal("expected auction status to bet set")
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
