package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AuctionStatus string

const (
	StatusOpen      AuctionStatus = "OPEN"
	StatusClosed    AuctionStatus = "CLOSED"
	AuctionDuration               = 100 * time.Millisecond
)

var ErrInvalidReservePrice = errors.New("reserve price should be > 0")
var ErrInvalidItemID = errors.New("itemID is nil")
var ErrAuctionIsClosed = errors.New("auction is already closed")
var ErrAmountLowerThanReservePrice = errors.New("bid amount is lower than reserve price")
var ErrAuctionIsExpired = errors.New("auction is expired")
var ErrAmountNotHigherThanHighestBid = errors.New("bid amount is not higher than the highest auction bid amount")
var ErrAuctionIsOpen = errors.New("auction is not closed yet")
var ErrNoBidsPlaced = errors.New("no bids have been placed on this auction")
var ErrBidderAlreadyPlacedBid = errors.New("bidder has already placed a bid on this auction")

type Auction struct {
	id           uuid.UUID
	itemID       uuid.UUID
	reservePrice int64
	startAt      time.Time
	endAt        time.Time
	status       AuctionStatus
	leadingBid   *Bid
	bidsPlaced   []Bid
}

func (a *Auction) ID() uuid.UUID {
	return a.id
}

func (a *Auction) ItemID() uuid.UUID {
	return a.itemID
}

func (a *Auction) ReservePrice() int64 {
	return a.reservePrice
}

func (a *Auction) LeadingBid() *Bid {
	return a.leadingBid
}

func NewAuction(itemID uuid.UUID, reservePrice int64) (*Auction, error) {
	if reservePrice <= 0 {
		return nil, ErrInvalidReservePrice
	}

	if itemID == uuid.Nil {
		return nil, ErrInvalidItemID
	}

	startAt := time.Now()
	endAt := startAt.Add(AuctionDuration)

	return &Auction{
		id:           uuid.New(),
		itemID:       itemID,
		reservePrice: reservePrice,
		startAt:      startAt,
		endAt:        endAt,
		status:       StatusOpen,
		leadingBid:   nil,
	}, nil
}

func (a *Auction) Close() error {
	if a.status == StatusClosed {
		return ErrAuctionIsClosed
	}

	a.status = StatusClosed
	return nil
}

func (a *Auction) Winner() (*Bid, error) {
	if a.status != StatusClosed {
		return nil, ErrAuctionIsOpen
	}
	if a.leadingBid == nil {
		return nil, ErrNoBidsPlaced
	}
	return a.leadingBid, nil
}

func (auction *Auction) PlaceBid(bidderID uuid.UUID, amount int64) (*Bid, error) {

	if auction.isClosed() {
		return nil, ErrAuctionIsClosed
	}

	if auction.isExpired() {
		return nil, ErrAuctionIsExpired
	}

	if auction.hasBidderAlreadyPlacedBid(bidderID) {
		return nil, ErrBidderAlreadyPlacedBid
	}

	bid, err := NewBid(bidderID, amount)

	if err != nil {
		return nil, err
	}

	if bid.amount < auction.reservePrice {
		return nil, ErrAmountLowerThanReservePrice
	}

	if auction.leadingBid != nil && bid.amount <= auction.leadingBid.amount {
		return nil, ErrAmountNotHigherThanHighestBid
	}

	auction.bidsPlaced = append(auction.bidsPlaced, *bid)

	auction.leadingBid = bid
	return bid, nil
}

func (a *Auction) hasBidderAlreadyPlacedBid(bidderID uuid.UUID) bool {
	for _, b := range a.bidsPlaced {
		if b.BidderID() == bidderID {
			return true
		}
	}
	return false
}

func (a *Auction) isExpired() bool {
	return time.Now().After(a.endAt)
}

func (a *Auction) isClosed() bool {
	return a.status == StatusClosed
}
