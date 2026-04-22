package application

import (
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

type AuctionService struct {
	activeAuctionManager ActiveAuctionManager
	scheduler            Scheduler
	clock                domain.Clock
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

func NewAuctionService(activeAuctionManager ActiveAuctionManager, scheduler Scheduler, clock domain.Clock) *AuctionService {
	return &AuctionService{
		activeAuctionManager: activeAuctionManager,
		scheduler:            scheduler,
		clock:                clock,
	}
}

func (a *AuctionService) CreateAuction(auctionCommand CreateAuctionCommand) (*AuctionResult, error) {
	auction, err := domain.NewAuction(auctionCommand.ItemID, auctionCommand.ReservePrice, a.clock)
	if err != nil {
		return nil, err
	}

	err = a.activeAuctionManager.Save(auction)
	if err != nil {
		return nil, err
	}

	err = a.scheduler.Schedule(auction.EndTime(), func() {
		err = a.closeAuction(auction.ID())
		if err != nil {
			_ = err // TODO: idk what to do with it yet, retry ?
		}
	})
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

	winner, err := a.activeAuctionManager.CloseAuction(id)
	if err != nil {
		return err
	}

	// TODO: publish winner to bidders
	_ = winner

	return nil

}

func (a *AuctionService) PlaceBid(bidCommand BidCommand) (*BidResult, error) {

	auctionID, bidderID, amount := bidCommand.AuctionID, bidCommand.BidderID, bidCommand.Amount
	bid, err := a.activeAuctionManager.PlaceBid(auctionID, bidderID, amount)
	if err != nil {
		return nil, err
	}

	return &BidResult{
		AuctionID: auctionID,
		BidderID:  bid.BidderID(),
		Amount:    bid.Amount(),
	}, nil
}
