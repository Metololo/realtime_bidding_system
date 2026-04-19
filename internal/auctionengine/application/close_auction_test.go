package application

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/repository/inmemory"
	"github.com/google/uuid"
)

func TestCloseAuctionWithNoBidsSuccessfullyClosesAuction(t *testing.T) {
	auctionRepository := inmemory.NewAuctionRepository()
	auctionService := NewAuctionService(auctionRepository)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auctionService.closeAuction(auctionResult.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrievedAuction, err := auctionRepository.FindByID(auctionResult.ID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if retrievedAuction != nil {
		t.Fatal("expected retrieved auction to be nil")
	}
	if !errors.Is(err, inmemory.ErrAuctionNotFound) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotFound, err)
	}

}

func TestCloseNonExistentAuctionReturnsError(t *testing.T) {
	auctionRepository := inmemory.NewAuctionRepository()
	auctionService := NewAuctionService(auctionRepository)

	nonExistentAuctionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	err := auctionService.closeAuction(nonExistentAuctionID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, inmemory.ErrAuctionNotFound) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotFound, err)
	}
}
