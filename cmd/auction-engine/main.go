package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/testutils"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	fmt.Printf("starting auction-engine")
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}

	activeAuctionManager := inmemory.NewActiveAuctionManager()
	scheduler := &testutils.FakeManualScheduler{}
	auctionService := application.NewAuctionService(
		activeAuctionManager,
		scheduler,
		testutils.NewFakeClock(time.Now()),
		&testutils.FakeEventPublisher{})

	handler := infrastructure.NewBidPlacerGRCP(auctionService)

	grpcServer := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(grpcServer, handler)

	reflection.Register(grpcServer)
	go func() {
		log.Println("gRPC server running on :9001")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	httpHandler := infrastructure.NewAuctionCreatorHTTP(auctionService).Handler()

	log.Println("http handler running on port 8080")
	if err = http.ListenAndServe(":8080", httpHandler); err != nil {
		panic("failed to start http server: " + err.Error())
	}
}
