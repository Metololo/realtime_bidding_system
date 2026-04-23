package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestNewAuctionGivenARequest(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	clock := testutils.NewFakeClock(time.Now())
	auction, err := domain.NewAuction(itemID, reservePrice, clock)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction == nil {
		t.Fatal("Auction is nil")
	}

	if auction.ItemID() != itemID {
		t.Fatal("expected auction itemID to be set")
	}

	if auction.ReservePrice() != reservePrice {
		t.Fatal("expected auction reservePrice to match request")
	}

	if auction.Status() == "" {
		t.Fatal("expected auction status to be set")
	}

	if auction.ID() == uuid.Nil {
		t.Fatal("expected auction ID to be set")
	}

	if auction.StartTime().IsZero() {
		t.Fatal("expected startAt to be set")
	}

	if auction.EndTime().IsZero() {
		t.Fatal("expected endAt to be set")
	}

	if !auction.ClosedAt().IsZero() {
		t.Fatal("expected closeAt to have zero value")
	}
}

func TestNewAuctionSetsStatusOpen(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.Status() != domain.StatusOpen {
		t.Fatal("expected auction status to be OPEN")
	}
}

func TestNewAuctionSetsEndAtAfterConfiguredDuration(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	auctionDuration := auction.EndTime().Sub(auction.StartTime())

	if auctionDuration != domain.AuctionDuration {
		t.Fatalf("expected auction duration to be %s, got %s", domain.AuctionDuration, auctionDuration)
	}
}

func TestNewAuctionReturnsErrorForNegativeReservePrice(t *testing.T) {
	itemID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reservePrice := int64(-1)

	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, domain.ErrInvalidReservePrice) {
		t.Fatalf("expected ErrNonPositiveReservePrice, got %v", err)
	}
}

func TestNewAuctionReturnsErrorForZeroReservePrice(t *testing.T) {
	itemID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	reservePrice := int64(0)

	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, domain.ErrInvalidReservePrice) {
		t.Fatalf("expected ErrNonPositiveReservePrice, got %v", err)
	}
}

func TestNewAuctionReturnsErrorForNilItemID(t *testing.T) {
	itemID := uuid.Nil
	reservePrice := int64(10)

	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))
	if err == nil {
		t.Fatalf("error is nil")
	}

	if auction != nil {
		t.Fatal("expected no auction to be created")
	}

	if !errors.Is(err, domain.ErrNilItemID) {
		t.Fatalf("expected ErrNilItemId, got %v", err)
	}
}

func TestCloseAnExistingAuction(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	fakeClock := testutils.NewFakeClock(time.Now())
	auction, err := domain.NewAuction(itemID, reservePrice, fakeClock)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !auction.ClosedAt().IsZero() {
		t.Fatal("expected auction closeAt to have zero value")
	}

	err = auction.Close()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.Status() != domain.StatusClosed {
		t.Fatal("expected auction status to be CLOSED")
	}

	if auction.ClosedAt().IsZero() {
		t.Fatal("expected auction closeAt to be set")
	}

	if auction.ClosedAt() != fakeClock.Now() {
		t.Fatalf("expected auction closeAt to be %v, got %v", fakeClock.Now(), auction.ClosedAt())
	}
}

func TestCannotCloseAnAlreadyClosedAuction(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

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

	if !errors.Is(err, domain.ErrAuctionIsClosed) {
		t.Fatalf("expected ErrAuctionIsClosed, got %v", err)
	}
}

func TestNewAuctionHasNoLeadingBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auction.LeadingBid() != nil {
		t.Fatal("expected bid to be nil")
	}
}

func TestAuctionWinnerReturnsLeadingBid(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	amount := int64(200)

	_, err = auction.PlaceBid(bidderID, amount)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auction.Close()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	winner, err := auction.Winner()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if winner == nil {
		t.Fatal("expected winner to be not nil")
	}

	if winner.BidderID() != bidderID {
		t.Fatalf("expected winner bidderID to be %v, got %v", bidderID, winner.BidderID())
	}
}

func TestAuctionWinnerReturnsErrorIfAuctionIsNotClosed(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = auction.Winner()

	if err == nil {
		t.Fatal("error is nil")
	}
	if !errors.Is(err, domain.ErrAuctionIsOpen) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrAuctionIsOpen, err)
	}
}

func TestAuctionWinnerReturnsNilIfNoBidsPlaced(t *testing.T) {
	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice, testutils.NewFakeClock(time.Now()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auction.Close()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	winner, err := auction.Winner()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if winner != nil {
		t.Fatal("expected winner to be nil")
	}
}

func newTestAuctionRequest() (uuid.UUID, int64) {
	return uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), 100
}
