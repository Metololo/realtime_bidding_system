package application

import (
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

type ActiveAuctionManager interface {
	Save(auction *domain.Auction) error
	PlaceBid(id uuid.UUID, bidderID uuid.UUID, amount int64) (*domain.Bid, error)
	CloseAuction(id uuid.UUID) (*domain.Bid, error)
}
