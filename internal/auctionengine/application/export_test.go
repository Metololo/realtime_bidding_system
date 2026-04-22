package application

import "github.com/google/uuid"

func (a *AuctionService) CloseAuctionForTest(id uuid.UUID) error {
	return a.closeAuction(id)
}
