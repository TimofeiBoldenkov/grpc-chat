package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	pb "github.com/TimofeiBoldenkov/grpc-chat/grpcchat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:50052", "The server socket")
	action     = flag.String("action", "get", "The performed action (get|send)")
)

func sendMessages(client pb.GrpcChatClient) error {
	log.Println("sending messages...")
	stream, err := client.SendMessages(context.Background())
	if err != nil {
		return fmt.Errorf("failed to call client.SendMessages: %v", err)
	}

	for {
		var text string
		reader := bufio.NewReader(os.Stdin)
		text, _ = reader.ReadString('\n')
		text = text[:len(text)-1]
		if text == `\q` {
			break
		}

		if err = stream.Send(&pb.Text{Text: text}); err != nil {
			return err
		}
	}

	return nil
}

func getNMessages(client pb.GrpcChatClient, n *pb.Amount) error {
	if n.GetAmount() != 0 {
		log.Printf("getting %v messages...", n)
	} else {
		log.Println("getting all messages...")
	}
	stream, err := client.GetMessages(context.Background(), n)
	if err != nil {
		return fmt.Errorf("failed to call client.GetMessages: %v", err)
	}

	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Printf("failed to get a message: %v", err)
		} else {
			fmt.Printf("%v %v\n", message.Username, message.Time)
			fmt.Printf("%v\n", message.Text)
			fmt.Printf("--------------------------------------------------------------\n")
		}
	}

	return nil
}

func getAllMessages(client pb.GrpcChatClient) error {
	return getNMessages(client, &pb.Amount{Amount: 0})
}

func main() {
	flag.Parse()
	conn, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to %v: %v", *serverAddr, err)
	}
	defer conn.Close()
	client := pb.NewGrpcChatClient(conn)

	if *action == "get" {
		err = getAllMessages(client)
		if err != nil {
			log.Fatalf("failed to get messages: %v", err)
		}
	} else if *action == "send" {
		err = sendMessages(client)
		if err != nil {
			log.Fatalf("failed to send messages: %v", err)
		}
	} else {
		log.Fatalf("invalid action: %v", *action)
	}
}
