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

type BidCommand struct {
	AuctionID uuid.UUID
	BidderID  uuid.UUID
	Amount    int64
}

type BidResult struct {
	AuctionID uuid.UUID
	BidderID  uuid.UUID
	Amount    int64
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

	err := a.auctionRepository.SetAuctionClosing(id)
	if err != nil {
		return err
	}

	unlock, err := a.auctionRepository.LockAuction(id)
	if err != nil {
		return err
	}
	defer unlock()

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

	// made the choice to delete the auction after closing it, but we could also keep it for historical purposes
	return a.auctionRepository.DeleteByID(id)
}

func (a *AuctionService) PlaceBid(bidCommand BidCommand) (*BidResult, error) {

	isClosing, err := a.auctionRepository.IsAuctionClosing(bidCommand.AuctionID)
	if err != nil {
		return nil, err
	}
	if isClosing {
		return nil, err
	}

	unlock, err := a.auctionRepository.LockAuction(bidCommand.AuctionID)
	if err != nil {
		return nil, err
	}
	defer unlock()

	auction, err := a.auctionRepository.FindByID(bidCommand.AuctionID)
	if err != nil {
		return nil, err
	}

	bid, err := auction.PlaceBid(bidCommand.BidderID, bidCommand.Amount)
	if err != nil {
		return nil, err
	}

	return &BidResult{
		AuctionID: auction.ID(),
		BidderID:  bid.BidderID(),
		Amount:    bid.Amount(),
	}, nil
}
