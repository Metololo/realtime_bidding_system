package application

import "github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"

type EventPublisher interface {
	Publish(event domain.Event) error
}
