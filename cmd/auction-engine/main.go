package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/application"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/active_auction_manager/inmemory"
	"github.com/Metololo/realtime_bidding_system/internal/auctionengine/infrastructure/scheduler"
	auctionpb "github.com/Metololo/realtime_bidding_system/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	fmt.Printf("starting auction-engine")
	grpcPort := env("AUCTION_ENGINE_GRPC_PORT", "9001")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}

	eventPublisher := infrastructure.NewNatsEventPublisher(env("NATS_SERVER_URL", "nats://localhost:4222"))
	if err := eventPublisher.StartClient(); err != nil {
		log.Fatalf("failed to connect to nats: %v", err)
	}
	defer eventPublisher.Close()

	activeAuctionManager := inmemory.NewActiveAuctionManager()
	auctionScheduler := scheduler.NewTimerScheduler()
	auctionService := application.NewAuctionService(
		activeAuctionManager,
		auctionScheduler,
		infrastructure.NewSystemClock(),
		eventPublisher)

	grpcHandler := infrastructure.NewBidPlacerGRCP(auctionService)

	grpcServer := grpc.NewServer()
	auctionpb.RegisterAuctionEngineServer(grpcServer, grpcHandler)

	reflection.Register(grpcServer)
	go func() {
		log.Println("gRPC server running on :" + grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	httpHandler := infrastructure.NewAuctionCreatorHTTP(auctionService).Handler()

	port := env("AUCTION_ENGINE_SELL_PORT", "8080")
	addr := ":" + port

	log.Printf("http handler running on port %s\n", port)
	if err = http.ListenAndServe(addr, httpHandler); err != nil {
		panic("failed to start http server: " + err.Error())
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
