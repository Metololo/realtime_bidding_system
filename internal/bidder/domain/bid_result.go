package domain

import "errors"

var ErrAuctionUnavailable = errors.New("auction is unavailable")
var ErrAuctionEnded = errors.New("auction has ended")
var ErrBidTooLow = errors.New("bid is too low")
var ErrBidderAlreadyPlacedBid = errors.New("bidder has already placed a bid")
