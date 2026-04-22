package application

import (
	"errors"
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestCreateAuctionSchedulesClosesAuction(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	fakeScheduler := &testutils.FakeManualScheduler{}
	auctionService := NewAuctionService(activeAuctionManager, fakeScheduler, testutils.NewFakeClock(time.Now()))

	auctionCommand := newTestCreateAuctionCommand()
	auctionResult, err := auctionService.CreateAuction(auctionCommand)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeScheduler.ScheduledJobs) != 1 {
		t.Fatalf("scheduled calls = %d, want 1", len(fakeScheduler.ScheduledJobs))
	}

	fakeScheduler.ExecuteLastScheduledTask()

	_, err = activeAuctionManager.PlaceBid(auctionResult.ID, uuid.New(), 100)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, inmemory.ErrAuctionNotActive) {
		t.Fatalf("expected error to be %v, got %v", inmemory.ErrAuctionNotActive, err)
	}
}

func newTestCreateAuctionCommand() CreateAuctionCommand {
	return CreateAuctionCommand{
		ItemID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		ReservePrice: 100,
	}
}
