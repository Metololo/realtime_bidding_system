package main

import (
	"fmt"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/scheduler"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
)

func main() {
	fmt.Printf("starting auction-engine")

	activeAuctionManager := inmemory.NewActiveAuctionManager()
	scheduler := scheduler.NewTimerScheduler()
	auctionService := application.NewAuctionService(
		activeAuctionManager,
		scheduler,
		infrastructure.NewSystemClock(),
		&testutils.FakeEventPublisher{})

	httpHandler := infrastructure.NewAuctionCreatorHTTP(auctionService)
	err := httpHandler.StartHTTPServer()
	if err != nil {
		panic("failed to start http server")
	}
}
