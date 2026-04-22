package application_test

import (
	"testing"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	"github.com/google/uuid"
)

func TestCreateAuctionPublishEvent(t *testing.T) {
	activeAuctionManager := inmemory.NewActiveAuctionManager()
	fakeScheduler := &testutils.FakeManualScheduler{}
	fakeEventPublisher := &testutils.FakeEventPublisher{}
	fakeClock := testutils.NewFakeClock(time.Now())
	auctionService := application.NewAuctionService(activeAuctionManager, fakeScheduler, fakeClock, fakeEventPublisher)

	itemId := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	reservePrice := int64(100)

	_, err := auctionService.CreateAuction(application.CreateAuctionCommand{
		ItemID:       itemId,
		ReservePrice: reservePrice,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fakeEventPublisher.EventsPublished) != 1 {
		t.Fatalf("expected 1 event to be published, got %d", len(fakeEventPublisher.EventsPublished))
	}

	eventReceived := fakeEventPublisher.EventsPublished[0]

	if eventReceived == nil {
		t.Fatal("expected event to be published, got nil")
	}

	if eventReceived.EventID() == uuid.Nil {
		t.Fatal("expected event to have a non-nil ID")
	}
	if eventReceived.OccurredAt() != fakeClock.Now() {
		t.Fatalf("expected event to have occurred at %v, got %v", fakeClock.Now(), eventReceived.OccurredAt())
	}
	if eventReceived.EventType() != domain.EventAuctionCreated {
		t.Fatalf("expected event to have type %v, got %v", domain.EventAuctionCreated, eventReceived.EventType())
	}

	createEvent, ok := eventReceived.(domain.AuctionCreatedEvent)
	if !ok {
		t.Fatalf("expected event to be of type %T, got %T", &domain.AuctionCreatedEvent{}, eventReceived)
	}

	if createEvent.ItemID != itemId {
		t.Fatalf("expected event to have item ID %v, got %v", itemId, createEvent.ItemID)
	}
	if createEvent.ReservePrice != reservePrice {
		t.Fatalf("expected event to have reserve price %v, got %v", reservePrice, createEvent.ReservePrice)
	}
	if createEvent.StartedAt != fakeClock.Now() {
		t.Fatalf("expected event to have started at %v, got %v", fakeClock.Now(), createEvent.StartedAt)
	}
	if createEvent.EndedAt != fakeClock.Now().Add(100*time.Millisecond) {
		t.Fatalf("expected event to have ended at %v, got %v", fakeClock.Now().Add(100*time.Millisecond), createEvent.EndedAt)
	}
}
