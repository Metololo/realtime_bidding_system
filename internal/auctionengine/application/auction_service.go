package application

import (
	"errors"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

type AuctionService struct {
	auctionRepository AuctionRepository
}

type CreateAuctionCommand struct {
	ItemID       uuid.UUID
	ReservePrice int64
}

type AuctionResult struct {
	ID           uuid.UUID
	ItemID       uuid.UUID
	ReservePrice int64
}

func NewAuctionService(auctionRepository AuctionRepository) *AuctionService {
	return &AuctionService{
		auctionRepository: auctionRepository,
	}
}

func (a *AuctionService) CreateAuction(auctionCommand CreateAuctionCommand) (*AuctionResult, error) {
	auction, err := domain.NewAuction(auctionCommand.ItemID, auctionCommand.ReservePrice)
	if err != nil {
		return nil, err
	}

	err = a.auctionRepository.Save(auction)
	if err != nil {
		return nil, err
	}

	return &AuctionResult{
		ID:           auction.ID(),
		ItemID:       auction.ItemID(),
		ReservePrice: auction.ReservePrice(),
	}, nil
}

func (a *AuctionService) closeAuction(id uuid.UUID) error {
	auction, err := a.auctionRepository.FindByID(id)
	if err != nil {
		return err
	}

	err = auction.Close()
	if err != nil {
		return err
	}

	winner, err := auction.Winner()

	if err != nil && !errors.Is(err, domain.ErrNoBidsPlaced) {
		return err
	}

	// TODO: publish winner to bidders
	_ = winner

	return a.auctionRepository.DeleteByID(id)
}
