package application

type AuctionCreator interface {
	CreateAuction(command CreateAuctionCommand) (*AuctionResult, error)
}
