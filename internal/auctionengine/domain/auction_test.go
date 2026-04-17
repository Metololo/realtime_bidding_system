package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewAuctionGivenARequest(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.status != StatusOpen {
		t.Fatal("expected auction status to be OPEN")
	}
}

func TestNewAuctionSetsEndAtAfterConfiguredDuration(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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

	if !errors.Is(err, ErrInvalidReservePrice) {
		t.Fatalf("expected ErrNonPositiveReservePrice, got %v", err)
	}
}

func TestNewAuctionReturnsErrorForZeroReservePrice(t *testing.T) {
	itemID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reservePrice := int64(0)

	auction, err := NewAuction(itemID, reservePrice)

	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, ErrInvalidReservePrice) {
		t.Fatalf("expected ErrNonPositiveReservePrice, got %v", err)
	}
}

func TestNewAuctionReturnsErrorForNilItemID(t *testing.T) {
	itemID := uuid.Nil
	reservePrice := int64(10)

	auction, err := NewAuction(itemID, reservePrice)

	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, ErrInvalidItemID) {
		t.Fatalf("expected ErrInvalidItemID, got %v", err)
	}
}

func TestCloseAnExistingAuction(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auction.Close()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.status != StatusClosed {
		t.Fatal("expected auction status to be CLOSED")
	}
}

func TestCannotCloseAnAlreadyClosedAuction(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auction.Close()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auction.Close()

	if err == nil {
		t.Fatal("error is nil")
	}

	if !errors.Is(err, ErrAuctionIsClosed) {
		t.Fatalf("expected ErrAuctionIsClosed, got %v", err)
	}
}

func TestNewAuctionHasNoLeadingBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.leadingBid != nil {
		t.Fatal("expected bid to be nil")
	}
}

func TestAuctionPlaceBidAcceptsFirstBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

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
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(reservePrice - 10)

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
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(reservePrice + 10)

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
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = auction.Close()

	if err != nil {
		t.Fatal("failed to close auction")
	}

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(reservePrice + 10)

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
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := NewAuction(itemID, reservePrice)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	leadingBidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	leadingAmount := int64(reservePrice + 50)

	bidderID := uuid.MustParse("444e4444-e29b-41d4-a716-446655440000")
	amount := int64(reservePrice + 10)

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

func newTestAuctionRequest() (uuid.UUID, int64) {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), int64(150)
}
