package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAuctionPlaceBidAcceptsFirstBid(t *testing.T) {
	auction := newTestAuction(t)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(150)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bid == nil {
		t.Fatal("placed bid is nil")
	}

	if auction.leadingBid == nil {
		t.Fatal("auction leading bid is nil")
	}

	if auction.leadingBid != bid {
		t.Fatalf("expected returned bid to be auction leading bid")
	}

	if auction.leadingBid.amount != amount {
		t.Fatalf("expected leading bid amount to be %d, got %d", amount, auction.leadingBid.amount)
	}
}
func TestAuctionPlaceBidReturnsErrorWhenAmountIsLowerThanReservePrice(t *testing.T) {
	auction := newTestAuction(t)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.reservePrice - 10)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, ErrAmountLowerThanReservePrice) {
		t.Fatalf("expected error ErrAmountLowerThanReservePrice, got %v", err)
	}
}

func TestAuctionPlaceBidReturnsErrorWhenAuctionIsExpired(t *testing.T) {
	auction := newTestAuction(t)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.reservePrice + 10)

	auction.endAt = time.Now().Add(-100 * time.Millisecond)
	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, ErrAuctionIsExpired) {
		t.Fatalf("expected error ErrAuctionIsExpired, got %v", err)
	}
}

func TestAuctionPlaceBidReturnsErrorWhenAuctionIsClosed(t *testing.T) {
	auction := newTestAuction(t)
	err := auction.Close()

	if err != nil {
		t.Fatal("failed to close auction")
	}

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.reservePrice + 10)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, ErrAuctionIsClosed) {
		t.Fatalf("expected error ErrAuctionIsClosed, got %v", err)
	}
}

func TestAuctionPlaceBidReturnsErrorWhenLowerThanLeadingBidAmount(t *testing.T) {
	auction := newTestAuction(t)

	leadingBidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	leadingAmount := int64(auction.reservePrice + 50)

	bidderID := uuid.MustParse("444e4444-e29b-41d4-a716-446655440000")
	amount := int64(auction.reservePrice + 10)

	leadingBid, err := auction.PlaceBid(leadingBidderID, leadingAmount)
	if err != nil {
		t.Fatalf("failed to place bid for leading bid, got %v", err)
	}

	bid, err := auction.PlaceBid(bidderID, amount)

	if err == nil {
		t.Fatal("error is nil", err)
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if auction.leadingBid != leadingBid {
		t.Fatal("expected leading bid to not change")
	}

	if !errors.Is(err, ErrAmountLowerThanHighestBid) {
		t.Fatalf("expected error to be ErrAmountLowerThanHighestBid, got %v", err)
	}

}

func TestAuctionPlaceBidAcceptsHigherAmountThanLeadingBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("error should be nil, got %v", err)
	}

	leadingBidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	leadingAmount := int64(reservePrice + 10)

	bidderID := uuid.MustParse("444e4444-e29b-41d4-a716-446655440000")
	amount := int64(reservePrice + 50)

	_, err = auction.PlaceBid(leadingBidderID, leadingAmount)
	if err != nil {
		t.Fatalf("failed to place bid for leading bid, got %v", err)
	}

	bid, err := auction.PlaceBid(bidderID, amount)

	if err != nil {
		t.Fatal("error should be nil")
	}

	if bid == nil {
		t.Fatal("bid is nil")
	}

	if auction.leadingBid != bid {
		t.Fatal("expected leading bid to be set to highest bid")
	}
	if auction.leadingBid.amount != bid.amount {
		t.Fatalf("expected leading bid amount to be %v, got %v", bid.amount, auction.leadingBid.amount)
	}
	if auction.leadingBid.bidderID != bid.bidderID {
		t.Fatalf("expected leadin bid id to be %v, got %v", bid.bidderID, auction.leadingBid.bidderID)
	}
}

func newTestAuction(t *testing.T) *Auction {
	t.Helper()

	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	return auction
}
