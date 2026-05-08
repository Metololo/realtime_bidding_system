package main

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
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

	logger := infrastructure.NewAuctionEngineLogger(slog.LevelDebug)
	slog.SetDefault(logger)

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

	grpcHandler := infrastructure.NewBidPlacerGRCP(auctionService)

	grpcServer := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(grpcServer, grpcHandler)

	reflection.Register(grpcServer)
	go func() {
		log.Println("gRPC server running on :9001")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	httpHandler := infrastructure.NewAuctionCreatorHTTP(auctionService).Handler()

	port := os.Getenv("AUCTION_ENGINE_SELL_PORT")
	addr := ":" + port

	log.Printf("http handler running on port %s\n", port)
	if err = http.ListenAndServe(addr, httpHandler); err != nil {
		panic("failed to start http server: " + err.Error())
	}
}
