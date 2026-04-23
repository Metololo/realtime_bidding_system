package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
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

	if auction.LeadingBid() == nil {
		t.Fatal("auction leading bid is nil")
	}

	if auction.LeadingBid() != bid {
		t.Fatalf("expected returned bid to be auction leading bid")
	}

	if auction.LeadingBid().Amount() != amount {
		t.Fatalf("expected leading bid amount to be %d, got %d", amount, auction.LeadingBid().Amount())
	}
}
func TestAuctionPlaceBidReturnsErrorWhenAmountIsLowerThanReservePrice(t *testing.T) {
	auction := newTestAuction(t)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.ReservePrice() - 10)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, domain.ErrAmountLowerThanReservePrice) {
		t.Fatalf("expected error ErrAmountLowerThanReservePrice, got %v", err)
	}
}

func TestAuctionPlaceBidReturnsErrorWhenAuctionIsExpired(t *testing.T) {
	fakeClock := testutils.NewFakeClock(time.Now())
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, fakeClock)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fakeClock.Advance(101 * time.Millisecond)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.ReservePrice() + 10)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, domain.ErrAuctionIsExpired) {
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
	amount := int64(auction.ReservePrice() + 10)

	bid, err := auction.PlaceBid(bidderID, amount)
	if err == nil {
		t.Fatal("error is nil")
	}

	if bid != nil {
		t.Fatal("bid is not nil")
	}

	if !errors.Is(err, domain.ErrAuctionIsClosed) {
		t.Fatalf("expected error ErrAuctionIsClosed, got %v", err)
	}
}

func TestAuctionPlaceBidReturnsErrorWhenLowerThanLeadingBidAmount(t *testing.T) {
	auction := newTestAuction(t)

	leadingBidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	leadingAmount := int64(auction.ReservePrice() + 50)

	bidderID := uuid.MustParse("444e4444-e29b-41d4-a716-446655440000")
	amount := int64(auction.ReservePrice() + 10)

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

	if auction.LeadingBid() != leadingBid {
		t.Fatal("expected leading bid to not change")
	}

	if !errors.Is(err, domain.ErrAmountNotHigherThanHighestBid) {
		t.Fatalf("expected error to be ErrAmountLowerThanHighestBid, got %v", err)
	}

}

func TestAuctionPlaceBidReturnsErrorWhenAmountEqualsLeadingBidAmount(t *testing.T) {
	auction := newTestAuction(t)

	leadingBidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	leadingAmount := int64(auction.ReservePrice())

	bidderID := uuid.MustParse("444e4444-e29b-41d4-a716-446655440000")
	amount := int64(auction.ReservePrice())

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

	if auction.LeadingBid() != leadingBid {
		t.Fatal("expected leading bid to not change")
	}

	if !errors.Is(err, domain.ErrAmountNotHigherThanHighestBid) {
		t.Fatalf("expected error to be ErrAmountLowerThanHighestBid, got %v", err)
	}

}

func TestAuctionPlaceBidAcceptsHigherAmountThanLeadingBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

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

	if auction.LeadingBid() != bid {
		t.Fatal("expected leading bid to be set to highest bid")
	}
	if auction.LeadingBid().Amount() != bid.Amount() {
		t.Fatalf("expected leading bid amount to be %v, got %v", bid.Amount(), auction.LeadingBid().Amount())
	}
	if auction.LeadingBid().BidderID() != bid.BidderID() {
		t.Fatalf("expected leading bid id to be %v, got %v", bid.BidderID(), auction.LeadingBid().BidderID())
	}
}

func TestAuctionPlaceBidRejectIfBidderAlreadyPlacedBid(t *testing.T) {
	auction := newTestAuction(t)

	bidderID := uuid.MustParse("123e1234-e29b-41d4-a716-446655440000")
	amount := int64(auction.ReservePrice() + 10)

	_, err := auction.PlaceBid(bidderID, amount)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = auction.PlaceBid(bidderID, amount+10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrBidderAlreadyPlacedBid) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrBidderAlreadyPlacedBid, err)
	}
}

func newTestAuction(t *testing.T) *domain.Auction {
	t.Helper()

	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	return auction
}
