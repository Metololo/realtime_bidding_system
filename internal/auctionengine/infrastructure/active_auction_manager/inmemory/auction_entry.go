package inmemory

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/domain"
	"github.com/google/uuid"
)

var ErrAuctionClosing = errors.New("auction is closing")
var ErrAuctionAlreadyClosing = errors.New("auction already closing")

type auctionEntry struct {
	auction *domain.Auction
	closing atomic.Bool
	mu      sync.Mutex
}

func (a *auctionEntry) Lock() {
	a.mu.Lock()
}

func (a *auctionEntry) Unlock() {
	a.mu.Unlock()
}

func (a *auctionEntry) TrySetClosing() bool {
	return a.closing.CompareAndSwap(false, true)
}

func (a *auctionEntry) isClosing() bool {
	return a.closing.Load()
}

func (a *auctionEntry) PlaceBid(bidderID uuid.UUID, amount int64) (*domain.Bid, error) {

	if a.isClosing() {
		return nil, ErrAuctionClosing
	}

	a.Lock()
	defer a.Unlock()

	if a.isClosing() {
		return nil, ErrAuctionClosing
	}

	auction := a.auction

	bid, err := auction.PlaceBid(bidderID, amount)
	if err != nil {
		return nil, err
	}

	return bid, nil
}

func (a *auctionEntry) CloseAuction() (application.CloseAuctionResult, error) {
	if !a.TrySetClosing() {
		return application.CloseAuctionResult{}, ErrAuctionAlreadyClosing
	}

	a.Lock()
	defer a.Unlock()

	err := a.auction.Close()
	if err != nil {
		return application.CloseAuctionResult{}, err
	}

	winnerBid, _ := a.auction.Winner()

	var winnerInfo *domain.WinnerInfo
	if winnerBid != nil {
		winnerInfo = &domain.WinnerInfo{
			BidderID: winnerBid.BidderID(),
			Amount:   winnerBid.Amount(),
		}
	}

	return application.CloseAuctionResult{
		AuctionID:  a.auction.ID(),
		ItemID:     a.auction.ItemID(),
		ClosedAt:   a.auction.ClosedAt(),
		WinnerInfo: winnerInfo,
	}, nil
}
