package inmemory

import (
	"errors"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestActiveAuctionManagerSaveAuction(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)

	err := auctionStore.Save(auction)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestActiveAuctionManagerSaveNilAuctionReturnsError(t *testing.T) {
	auctionStore := NewActiveAuctionManager()

	err := auctionStore.Save(nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNilAuction) {
		t.Fatalf("expected error to be %v, got %v", ErrNilAuction, err)
	}
}

func TestActiveAuctionManagerSaveDuplicateAuctionReturnsError(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)

	err := auctionStore.Save(auction)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auctionStore.Save(auction)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionAlreadyExists) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionAlreadyExists, err)
	}
}

func TestActiveAuctionManagerPlaceBidOnActiveAuction(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)
	reservePrice := auction.ReservePrice()

	err := auctionStore.Save(auction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	bidAmount := int64(reservePrice)

	bid, err := auctionStore.PlaceBid(auction.ID(), bidderID, bidAmount)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bid == nil {
		t.Fatal("bid is nil")
	}

	if bid.BidderID() != bidderID {
		t.Fatalf("expected bidder ID to be %v, got %v", bidderID, bid.BidderID())
	}

	if bid.Amount() != bidAmount {
		t.Fatalf("expected bid amount to be %v, got %v", bidAmount, bid.Amount())
	}
}

func TestActiveAuctionManagerPlaceBidOnNonExistentAuctionReturnsError(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	bidderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	auctionID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	_, err := auctionStore.PlaceBid(auctionID, bidderID, 100)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionNotActive, err)
	}
}

func TestActiveAuctionManagerPlaceBidOnClosedAuctionReturnsError(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)

	err := auctionStore.Save(auction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = auctionStore.CloseAuction(auction.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	_, err = auctionStore.PlaceBid(auction.ID(), bidderID, 100)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionNotActive, err)
	}
}

func TestActiveAuctionManagerCloseAuction(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)

	err := auctionStore.Save(auction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	auctionResult, err := auctionStore.CloseAuction(auction.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auctionResult.WinnerInfo != nil {
		t.Fatalf("expected winner bid to be nil, got %v", auctionResult)
	}
}

func TestActiveAuctionManagerCloseNonExistentAuctionReturnsError(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auctionID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")

	_, err := auctionStore.CloseAuction(auctionID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionNotActive, err)
	}
}

func TestActiveAuctionReturnWinningBidOnClosedAuction(t *testing.T) {
	auctionStore := NewActiveAuctionManager()
	auction := newTestAuction(t)
	reservePrice := auction.ReservePrice()

	err := auctionStore.Save(auction)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	bidAmount := int64(reservePrice + 50)

	_, err = auctionStore.PlaceBid(auction.ID(), bidderID, bidAmount)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidder2ID := uuid.MustParse("123e6666-e89b-12d3-a456-426614174000")

	_, err = auctionStore.PlaceBid(auction.ID(), bidder2ID, bidAmount+70)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	closeAuctionResult, err := auctionStore.CloseAuction(auction.ID())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if closeAuctionResult.WinnerInfo == nil {
		t.Fatal("expected winning bid to be not nil")
	}

	if closeAuctionResult.WinnerInfo.BidderID != bidder2ID {
		t.Fatalf("expected winning bidder ID to be %v, got %v", bidder2ID, closeAuctionResult.WinnerInfo.BidderID)
	}

	if closeAuctionResult.WinnerInfo.Amount != bidAmount+70 {
		t.Fatalf("expected winning bid amount to be %v, got %v", bidAmount+70, closeAuctionResult.WinnerInfo.Amount)
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

func newTestAuctionRequest() (uuid.UUID, int64) {
	return uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), 100
}
