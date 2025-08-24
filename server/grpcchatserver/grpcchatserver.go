package grpcchatserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	pb "github.com/TimofeiBoldenkov/grpc-chat/grpcchat"
	"github.com/TimofeiBoldenkov/grpc-chat/server/db"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type GrpcChatServer struct {
	pb.UnimplementedGrpcChatServer
}

func (s *GrpcChatServer) GetMessages(amount *pb.Amount, stream pb.GrpcChat_GetMessagesServer) error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("unable to open .env file: %v", err)
	}
	var (
		grpcTableName = os.Getenv("GRPC_TABLE_NAME")
	)

	conn, err := db.ConnectOrCreateDb()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var query string
	if amount.GetAmount() != 0 {
		query = fmt.Sprintf(`
			SELECT username, time, message 
			FROM %v ORDER BY time DESC LIMIT %v
			`, grpcTableName, amount.GetAmount())
	} else {
		query = fmt.Sprintf(`
			SELECT username, time, message
			FROM %v ORDER BY time
			`, grpcTableName)
	}
	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("unable to get messages from db: %v", err)
	}

	var username, message string
	var messageTime time.Time
	_, err = pgx.ForEachRow(rows, []any{&username, &messageTime, &message}, func() error {
		return stream.Send(&pb.Message{Text: message, Username: username, Time: messageTime.Format(time.RFC3339)})
	})
	if err != nil {
		return fmt.Errorf("unable to parse messages from db: %v", err)
	}

	return nil
}

func (s *GrpcChatServer) SendMessages(stream pb.GrpcChat_SendMessagesServer) error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("unable to open .env file: %v", err)
	}
	var (
		grpcTableName = os.Getenv("GRPC_TABLE_NAME")
	)

	conn, err := db.ConnectOrCreateDb()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var amount uint64

	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&pb.Amount{Amount: amount})
		}
		if err != nil {
			return err
		}

		query := fmt.Sprintf(`
			INSERT INTO %v (username, time, message) VALUES ($1, $2, $3)
			`, grpcTableName)
		_, err = conn.Exec(context.Background(),
			query, "TEST", time.Now().Format(time.RFC3339), message.GetText())
		if err != nil {
			return fmt.Errorf("unabled to insert message into db: %v", err)
		}

		amount++
	}
}
