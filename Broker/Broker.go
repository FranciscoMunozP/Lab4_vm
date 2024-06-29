package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"

	"google.golang.org/grpc"

	pb "github.com/FranciscoMunozP/Lab5_proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

type broker struct {
	pb.UnimplementedBrokerServer
	servers []string
	mu      sync.Mutex
}

func (b *broker) getRandomServer() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.servers[rand.Intn(len(b.servers))]
}

func (b *broker) GetServer(ctx context.Context, in *emptypb.Empty) (*pb.ServerResponse, error) {
	server := b.getRandomServer()
	fmt.Printf("Solicitado server %s\n", server)
	return &pb.ServerResponse{ServerAddress: server}, nil
}

func main() {
	b := &broker{
		servers: []string{"localhost:50051", "localhost:50052", "localhost:50053"},
	}

	lis, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBrokerServer(s, b)
	log.Printf("Broker listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
