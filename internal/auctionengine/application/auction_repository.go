package application

import (
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

type AuctionRepository interface {
	Save(auction *domain.Auction) error
	FindByID(id uuid.UUID) (*domain.Auction, error)
	DeleteByID(id uuid.UUID) error
	LockAuction(id uuid.UUID) (func(), error)
}
