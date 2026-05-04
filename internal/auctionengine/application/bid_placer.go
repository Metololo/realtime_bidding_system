package application

type BidPlacer interface {
	PlaceBid(command BidCommand) (*BidResult, error)
}
