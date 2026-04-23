package application

import "github.com/google/uuid"

func (a *AuctionService) CloseAuction(id uuid.UUID) error {
	return a.closeAuction(id)
}
