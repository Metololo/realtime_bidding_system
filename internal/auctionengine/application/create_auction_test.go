package application

import (
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/repository/inmemory"
	"github.com/google/uuid"
)

func TestCreateAuctionReturnsAuctionResult(t *testing.T) {
	auctionRepository := inmemory.NewAuctionRepository()
	auctionService := NewAuctionService(auctionRepository)

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

func newTestCreateAuctionCommand() CreateAuctionCommand {
	return CreateAuctionCommand{
		ItemID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		ReservePrice: 100,
	}
}
