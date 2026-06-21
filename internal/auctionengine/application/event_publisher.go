package application

import "github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"

type EventPublisher interface {
	PublishAuctionClosedEvent(event domain.AuctionClosedEvent) error
	PublishAuctionCreatedEvent(event domain.AuctionCreatedEvent) error
}
