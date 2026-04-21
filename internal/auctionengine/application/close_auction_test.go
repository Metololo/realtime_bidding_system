package application

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/google/uuid"
)

func TestCloseAuctionWithNoBidsSuccessfullyClosesAuction(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auctionService.closeAuction(auctionResult.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = activeAuctionManager.PlaceBid(auctionResult.ID, uuid.New(), 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotActive, err)
	}

}

func TestCloseNonExistentAuctionReturnsError(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	nonExistentAuctionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	err := auctionService.closeAuction(nonExistentAuctionID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotActive, err)
	}
}
