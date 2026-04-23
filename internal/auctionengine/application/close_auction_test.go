package application_test

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestCloseAuctionWithNoBidsSuccessfullyClosesAuction(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	fakeScheduler := &testutils.FakeManualScheduler{}
	auctionService := application.NewAuctionService(
		activeAuctionManager,
		fakeScheduler,
		infrastructure.NewSystemClock(),
		&testutils.FakeEventPublisher{})

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auctionService.CloseAuction(auctionResult.ID)

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
	auctionService := newTestAuctionService()

	nonExistentAuctionID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	err := auctionService.CloseAuction(nonExistentAuctionID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotActive, err)
	}
}
