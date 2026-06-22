package application

import "github.com/Metololo/realtime_bidding_system/internal/bidder/domain"

type BidPlacer struct {
	strategy      BidStrategy
	auctionEngine AuctionEngine
}

func NewBidPlacer(strategy BidStrategy, auctionEngine AuctionEngine) *BidPlacer {
	return &BidPlacer{
		strategy:      strategy,
		auctionEngine: auctionEngine,
	}
}

func (bp *BidPlacer) PlaceBid(auction domain.Auction) (domain.BidProposal, error) {
	bidProposal, err := bp.strategy.DecideBid(auction)
	if err != nil {
		return domain.BidProposal{}, err
	}

	return bp.auctionEngine.SubmitBid(bidProposal)
}
