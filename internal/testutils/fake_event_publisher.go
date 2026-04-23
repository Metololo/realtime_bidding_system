package testutils

import "github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"

type FakeEventPublisher struct {
	EventsPublished []domain.Event
}

func (f *FakeEventPublisher) Publish(event domain.Event) error {
	f.EventsPublished = append(f.EventsPublished, event)
	return nil
}

func (f *FakeEventPublisher) Reset() {
	f.EventsPublished = []domain.Event{}
}
