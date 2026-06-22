package application

import "github.com/Metololo/realtime_bidding_system/internal/bidder/domain"

type AuctionEngine interface {
	SubmitBid(bidProposal domain.BidProposal) (domain.BidProposal, error)
}
