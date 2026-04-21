package application

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/google/uuid"
)

func TestPlaceBidAcceptsValidBid(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidCommand := newTestPlaceBidCommand(auctionResult.ID)
	bidResult, err := auctionService.PlaceBid(bidCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if bidResult == nil {
		t.Fatal("bid result is nil")
	}

	if bidResult.AuctionID != bidCommand.AuctionID {
		t.Fatalf("expected auction ID to be %v, got %v", bidCommand.AuctionID, bidResult.AuctionID)
	}

	if bidResult.BidderID != bidCommand.BidderID {
		t.Fatalf("expected bidder ID to be %v, got %v", bidCommand.BidderID, bidResult.BidderID)
	}

	if bidResult.Amount != bidCommand.Amount {
		t.Fatalf("expected amount to be %v, got %v", bidCommand.Amount, bidResult.Amount)
	}
}

func TestPlaceBidWithInvalidAuctionIDReturnsError(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	bidCommand := newTestPlaceBidCommand(uuid.MustParse("00000000-0000-0000-0000-000000000000"))
	_, err := auctionService.PlaceBid(bidCommand)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected auction not found error, got %v", err)
	}
}

func TestPlaceBidWithAmountLessThanReservePriceReturnsError(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bidCommand := newTestPlaceBidCommand(auctionResult.ID)
	bidCommand.Amount = 50
	_, err = auctionService.PlaceBid(bidCommand)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrAmountLowerThanReservePrice) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrAmountLowerThanReservePrice, err)
	}
}

func TestPlaceBidWithAmountLessThanCurrentLeadingBidReturnsError(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	firstBidCommand := newTestPlaceBidCommand(auctionResult.ID)
	firstBidCommand.Amount = 150
	_, err = auctionService.PlaceBid(firstBidCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	secondBidCommand := newTestPlaceBidCommand(auctionResult.ID)
	secondBidCommand.BidderID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	secondBidCommand.Amount = 120
	_, err = auctionService.PlaceBid(secondBidCommand)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrAmountNotHigherThanHighestBid) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrAmountNotHigherThanHighestBid, err)
	}
}

func TestPlaceBidOnRemovedAuctionReturnsNotFound(t *testing.T) {
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

	bidCommand := newTestPlaceBidCommand(auctionResult.ID)
	_, err = auctionService.PlaceBid(bidCommand)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotActive, err)
	}
}

func TestPlaceBidWithEqualAmountToCurrentLeadingBidReturnsError(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionService := NewAuctionService(activeAuctionManager)

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	firstBidCommand := newTestPlaceBidCommand(auctionResult.ID)
	firstBidCommand.Amount = 150
	_, err = auctionService.PlaceBid(firstBidCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	secondBidCommand := newTestPlaceBidCommand(auctionResult.ID)
	secondBidCommand.BidderID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
	secondBidCommand.Amount = 150
	_, err = auctionService.PlaceBid(secondBidCommand)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrAmountNotHigherThanHighestBid) {
		t.Fatalf("expected error to be %v, got %v", domain.ErrAmountNotHigherThanHighestBid, err)
	}
}

func newTestPlaceBidCommand(auctionID uuid.UUID) BidCommand {
	return BidCommand{
		AuctionID: auctionID,
		BidderID:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Amount:    150,
	}
}
