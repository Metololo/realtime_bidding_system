package main

import (
	"log"
	"net"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/scheduler"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}

	activeAuctionManager := inmemory.NewActiveAuctionManager()
	scheduler := scheduler.NewTimerScheduler()
	auctionService := application.NewAuctionService(
		activeAuctionManager,
		scheduler,
		infrastructure.NewSystemClock(),
		&testutils.FakeEventPublisher{})

	handler := infrastructure.NewBidPlacerGRCP(auctionService)

	grpcServer := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(grpcServer, handler)

	reflection.Register(grpcServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start grpc server")
	}

	// fmt.Printf("starting auction-engine")

	// httpHandler := infrastructure.NewAuctionCreatorHTTP(auctionService).Handler()

	// err := http.ListenAndServe(":8080", httpHandler)
	// if err != nil {
	// 	panic("failed to start http server: " + err.Error())
	// }
}
