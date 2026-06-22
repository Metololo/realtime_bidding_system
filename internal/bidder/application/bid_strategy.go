package application

import "github.com/Metololo/realtime_bidding_system/internal/bidder/domain"

type BidStrategy interface {
	DecideBid(auction domain.Auction) (domain.BidProposal, error)
}
