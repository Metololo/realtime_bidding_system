package inmemory

import (
	"errors"
	"testing"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

func TestAuctionRepositorySaveAuctionAndFindByID(t *testing.T) {
	auctionRepository := NewAuctionRepository()
	auction := newTestAuction(t)

	err := auctionRepository.Save(auction)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrievedAuction, err := auctionRepository.FindByID(auction.ID())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retrievedAuction == nil {
		t.Fatal("retrieved auction is nil")
	}

	if retrievedAuction.ID() != auction.ID() {
		t.Fatalf("expected retrieved auction ID to be %v, got %v", auction.ID(), retrievedAuction.ID())
	}
}

func TestAuctionRepositoryFindByIDReturnsErrorIfAuctionNotFound(t *testing.T) {
	auctionRepository := NewAuctionRepository()
	nonExistentID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	_, err := auctionRepository.FindByID(nonExistentID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionNotFound) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionNotFound, err)
	}
}

func TestAuctionRepositorySaveReturnsErrorIfAuctionIsNil(t *testing.T) {
	auctionRepository := NewAuctionRepository()

	err := auctionRepository.Save(nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNilAuction) {
		t.Fatalf("expected error to be %v, got %v", ErrNilAuction, err)
	}
}

func TestAuctionRepositorySaveReturnsErrorIfAuctionAlreadyExists(t *testing.T) {
	auctionRepository := NewAuctionRepository()
	auction := newTestAuction(t)

	err := auctionRepository.Save(auction)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = auctionRepository.Save(auction)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionAlreadyExists) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionAlreadyExists, err)
	}
}

func TestLockAuctionReturnsErrorIfAuctionNotFound(t *testing.T) {
	auctionRepository := NewAuctionRepository()
	nonExistentID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	_, err := auctionRepository.LockAuction(nonExistentID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrAuctionNotFound) {
		t.Fatalf("expected error to be %v, got %v", ErrAuctionNotFound, err)
	}
}

func newTestAuction(t *testing.T) *domain.Auction {
	t.Helper()

	itemID, reservePrice := newTestAuctionRequest()
	auction, err := domain.NewAuction(itemID, reservePrice)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	return auction
}

func newTestAuctionRequest() (uuid.UUID, int64) {
	return uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), 100
}
