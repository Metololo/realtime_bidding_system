package inmemory

import (
	"errors"
	"sync"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

var ErrNilAuction = errors.New("auction is nil")
var ErrAuctionNotFound = errors.New("auction not found")
var ErrAuctionAlreadyExists = errors.New("auction with the same ID already exists")

type AuctionRepository struct {
	mu       sync.RWMutex
	auctions map[uuid.UUID]*domain.Auction
}

func NewAuctionRepository() *AuctionRepository {
	return &AuctionRepository{
		auctions: make(map[uuid.UUID]*domain.Auction),
	}
}

func (r *AuctionRepository) Save(auction *domain.Auction) error {

	if auction == nil {
		return ErrNilAuction
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := auction.ID()

	if _, exists := r.auctions[id]; exists {
		return ErrAuctionAlreadyExists
	}

	r.auctions[auction.ID()] = auction
	return nil
}
func (r *AuctionRepository) FindByID(id uuid.UUID) (*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	auction, exists := r.auctions[id]
	if !exists {
		return nil, ErrAuctionNotFound
	}
	return auction, nil
}

func (r *AuctionRepository) DeleteByID(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.auctions[id]; !exists {
		return ErrAuctionNotFound
	}
	delete(r.auctions, id)
	return nil
}
