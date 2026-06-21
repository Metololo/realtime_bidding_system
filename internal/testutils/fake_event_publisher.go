package testutils

import "github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"

type FakeEventPublisher struct {
	EventsPublished []domain.Event
}

func (f *FakeEventPublisher) PublishAuctionClosedEvent(event domain.AuctionClosedEvent) error {
	f.EventsPublished = append(f.EventsPublished, event)
	return nil
}

func (f *FakeEventPublisher) PublishAuctionCreatedEvent(event domain.AuctionCreatedEvent) error {
	f.EventsPublished = append(f.EventsPublished, event)
	return nil
}

func (f *FakeEventPublisher) Reset() {
	f.EventsPublished = []domain.Event{}
}
