package application

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

type AuctionService struct {
	activeAuctionManager ActiveAuctionManager
	scheduler            Scheduler
	clock                domain.Clock
	eventPublisher       EventPublisher
}

type CreateAuctionCommand struct {
	ItemID       uuid.UUID
	ReservePrice int64
}

type AuctionResult struct {
	ID           uuid.UUID
	ItemID       uuid.UUID
	ReservePrice int64
	StartTime    time.Time
	EndTime      time.Time
}

type CloseAuctionResult struct {
	AuctionID  uuid.UUID
	ItemID     uuid.UUID
	ClosedAt   time.Time
	WinnerInfo *domain.WinnerInfo
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

func NewAuctionService(
	activeAuctionManager ActiveAuctionManager,
	scheduler Scheduler,
	clock domain.Clock,
	eventPublisher EventPublisher) *AuctionService {
	return &AuctionService{
		activeAuctionManager: activeAuctionManager,
		scheduler:            scheduler,
		clock:                clock,
		eventPublisher:       eventPublisher,
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

	err = a.scheduleCloseAuction(auction.ID(), auction.EndTime())
	if err != nil {
		return nil, fmt.Errorf("failed to schedule auction %v: %w", auction.ID(), err)
	}
	slog.Info("auction closure scheduled",
		"auctionId", auction.ID(),
		"endTime", auction.EndTime().Format("2006-01-02 15:04:05"),
	)

	err = a.publishAuctionCreatedEvent(auction)
	if err != nil {
		return nil, fmt.Errorf("failed to publish auction %v: %w", auction.ID(), err)
	}

	return &AuctionResult{
		ID:           auction.ID(),
		ItemID:       auction.ItemID(),
		ReservePrice: auction.ReservePrice(),
		StartTime:    auction.StartTime(),
		EndTime:      auction.EndTime(),
	}, nil
}

func (a *AuctionService) scheduleCloseAuction(auctionId uuid.UUID, endTime time.Time) error {
	err := a.scheduler.Schedule(endTime, func() {
		err := a.closeAuction(auctionId)
		if err != nil {
			drift := time.Since(endTime)
			slog.Error("failed to close auction",
				slog.String("auctionID", auctionId.String()),
				slog.String("endTime", drift.String()),
			)
		}
	})
	return err
}

func (a *AuctionService) closeAuction(id uuid.UUID) error {

	closeAuctionResult, err := a.activeAuctionManager.CloseAuction(id)
	if err != nil {
		return err
	}

	err = a.publishAuctionClosedEvent(closeAuctionResult)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuctionService) publishAuctionCreatedEvent(auction *domain.Auction) error {
	return a.eventPublisher.PublishAuctionCreatedEvent(domain.AuctionCreatedEvent{
		BaseEvent:    domain.BaseEvent{ID: uuid.New(), At: a.clock.Now()},
		AuctionID:    auction.ID(),
		ItemID:       auction.ItemID(),
		ReservePrice: auction.ReservePrice(),
		StartedAt:    auction.StartTime(),
		EndAt:        auction.EndTime(),
	})
}

func (a *AuctionService) publishAuctionClosedEvent(closeAuctionResult CloseAuctionResult) error {

	outcome := deriveOutcome(closeAuctionResult.WinnerInfo)

	return a.eventPublisher.PublishAuctionClosedEvent(domain.AuctionClosedEvent{
		BaseEvent: domain.BaseEvent{ID: uuid.New(), At: a.clock.Now()},
		AuctionID: closeAuctionResult.AuctionID,
		ItemID:    closeAuctionResult.ItemID,
		Outcome:   outcome,
		ClosedAt:  closeAuctionResult.ClosedAt,
		Winner:    closeAuctionResult.WinnerInfo,
	})
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

func deriveOutcome(winner *domain.WinnerInfo) domain.AuctionOutcome {
	if winner == nil {
		return domain.AuctionOutcomeNoBids
	}
	return domain.AuctionOutcomeSold
}
