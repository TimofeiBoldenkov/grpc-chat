package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/TimofeiBoldenkov/grpc-chat/grpcchat"
	"github.com/TimofeiBoldenkov/grpc-chat/server/db"
	"github.com/TimofeiBoldenkov/grpc-chat/server/grpcchatserver"
	"google.golang.org/grpc"
)

var (
	port	= flag.Int("port", 50052, "The server port")
)

func newServer() *grpcchatserver.GrpcChatServer {
	return &grpcchatserver.GrpcChatServer{}
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	_, err = db.ConnectOrCreateDb()
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGrpcChatServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
