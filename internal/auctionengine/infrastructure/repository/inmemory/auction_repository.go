package inmemory

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

var ErrNilAuction = errors.New("auction is nil")
var ErrAuctionNotFound = errors.New("auction not found")
var ErrAuctionAlreadyExists = errors.New("auction with the same ID already exists")
var ErrAuctionClosing = errors.New("auction is closing")

type AuctionRepository struct {
	mu       sync.RWMutex
	auctions map[uuid.UUID]*auctionEntry
}
type auctionEntry struct {
	auction *domain.Auction
	closing atomic.Bool
	mu      sync.Mutex
}

func NewAuctionRepository() *AuctionRepository {
	return &AuctionRepository{
		auctions: make(map[uuid.UUID]*auctionEntry),
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

	r.auctions[auction.ID()] = &auctionEntry{auction: auction}
	return nil
}
func (r *AuctionRepository) FindByID(id uuid.UUID) (*domain.Auction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auctionEntry, exists := r.auctions[id]
	if !exists {
		return nil, ErrAuctionNotFound
	}
	return auctionEntry.auction, nil
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

func (r *AuctionRepository) findEntryById(id uuid.UUID) (*auctionEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	auctionEntry, exists := r.auctions[id]
	if !exists {
		return nil, ErrAuctionNotFound
	}
	return auctionEntry, nil
}

func (r *AuctionRepository) LockAuction(id uuid.UUID) (func(), error) {
	auctionEntry, err := r.findEntryById(id)
	if err != nil {
		return nil, err
	}

	auctionEntry.mu.Lock()

	return func() {
		auctionEntry.mu.Unlock()
	}, nil
}

func (r *AuctionRepository) SetAuctionClosing(id uuid.UUID) error {
	auctionEntry, err := r.findEntryById(id)
	if err != nil {
		return err
	}

	auctionEntry.closing.Store(true)
	return nil
}

func (r *AuctionRepository) IsAuctionClosing(id uuid.UUID) (bool, error) {
	auctionEntry, err := r.findEntryById(id)
	if err != nil {
		return false, err
	}
	return auctionEntry.closing.Load(), nil
}
