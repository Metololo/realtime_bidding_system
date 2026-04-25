package application_test

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

func TestCreateAuctionReturnsAuctionResult(t *testing.T) {
	auctionService := newTestAuctionService()

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if auctionResult == nil {
		t.Fatal("auction is nil")

	}

	if auctionResult.ItemID != auctionCommand.ItemID {
		t.Fatalf("expected itemID to be %v, got %v", auctionCommand.ItemID, auctionResult.ItemID)
	}

	if auctionResult.ReservePrice != auctionCommand.ReservePrice {
		t.Fatalf("expected reserve price to be %v, got %v", auctionCommand.ReservePrice, auctionResult.ReservePrice)
	}

}

func TestCreateAuctionSaveInAuctionManager(t *testing.T) {
	auctionService := newTestAuctionService()

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidCommand := newTestPlaceBidCommand(auctionResult.ID)
	_, err = auctionService.PlaceBid(bidCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

}

func TestCreateAuctionReturnsErrorIfItemIDIsNil(t *testing.T) {
	auctionService := newTestAuctionService()

	auctionCommand := newTestCreateAuctionCommand()
	auctionCommand.ItemID = uuid.Nil

	auction, err := auctionService.CreateAuction(auctionCommand)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if auction != nil {
		t.Fatal("expected auction to be nil")
	}

	if !errors.Is(err, domain.ErrNilItemID) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrNilItemID, err)
	}

}

func TestCreateAuctionReturnsErrorIfReservePriceIsInvalid(t *testing.T) {
	auctionService := newTestAuctionService()

	auctionCommand := newTestCreateAuctionCommand()
	auctionCommand.ReservePrice = -1

	auction, err := auctionService.CreateAuction(auctionCommand)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if auction != nil {
		t.Fatal("expected auction to be nil")
	}

	if !errors.Is(err, domain.ErrInvalidReservePrice) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrInvalidReservePrice, err)
	}

}
