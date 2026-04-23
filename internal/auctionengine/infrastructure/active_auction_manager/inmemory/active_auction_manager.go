package inmemory

import (
	"errors"
	"sync"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

var ErrNilAuction = errors.New("auction is nil")
var ErrAuctionNotActive = errors.New("auction is not active")
var ErrAuctionAlreadyExists = errors.New("auction with the same ID already exists")

type ActiveAuctionManager struct {
	mu       sync.RWMutex
	auctions map[uuid.UUID]*auctionEntry
}

func NewActiveAuctionManager() *ActiveAuctionManager {
	return &ActiveAuctionManager{
		auctions: make(map[uuid.UUID]*auctionEntry),
	}
}

func (r *ActiveAuctionManager) Save(auction *domain.Auction) error {

	if auction == nil {
		return ErrNilAuction
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	id := auction.ID()

	if _, exists := r.auctions[id]; exists {
		return ErrAuctionAlreadyExists
	}

	r.auctions[auction.ID()] = &auctionEntry{auction: auction}
	return nil
}

func (r *ActiveAuctionManager) PlaceBid(id uuid.UUID, bidderID uuid.UUID, amount int64) (*domain.Bid, error) {

	auctionEntry, err := r.findEntryById(id)
	if err != nil {
		return nil, err
	}

	bid, err := auctionEntry.PlaceBid(bidderID, amount)
	if err != nil {
		return nil, err
	}

	return bid, nil
}

func (r *ActiveAuctionManager) CloseAuction(id uuid.UUID) (application.CloseAuctionResult, error) {

	auctionEntry, err := r.findEntryById(id)
	if err != nil {
		return application.CloseAuctionResult{}, err
	}

	closeAuctionResult, err := auctionEntry.CloseAuction()
	if err != nil {
		return application.CloseAuctionResult{}, err
	}

	err = r.deleteByID(id)
	if err != nil {
		return application.CloseAuctionResult{}, err
	}

	return closeAuctionResult, nil
}

func (r *ActiveAuctionManager) deleteByID(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.auctions[id]; !exists {
		return ErrAuctionNotActive
	}
	delete(r.auctions, id)
	return nil
}

func (r *ActiveAuctionManager) findEntryById(id uuid.UUID) (*auctionEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auctionEntry, exists := r.auctions[id]
	if !exists {
		return nil, ErrAuctionNotActive
	}
	return auctionEntry, nil
}
