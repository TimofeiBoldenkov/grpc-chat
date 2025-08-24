package db

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func ConnectOrCreateDb() (*pgx.Conn, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("unable to load .env file: %v", err)
	}
	var (
		grpcDatabaseUrl    = os.Getenv("GRPC_DATABASE_URL")
		grpcDatabaseName   = os.Getenv("GRPC_DATABASE_NAME")
		messagesTableName      = os.Getenv("GRPC_TABLE_NAME")
	)
	var defaultDatabaseUrl = os.Getenv("DEFAULT_DATABASE_URL")
	if defaultDatabaseUrl == "" {
		defaultDatabaseUrl = "postgres://postgres@localhost:5432/postgres"
	}

	defaultDBConn, err := pgx.Connect(context.Background(), defaultDatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %v: %v", defaultDatabaseUrl, err)
	}
	defer defaultDBConn.Close(context.Background())

	var exists string
	err = defaultDBConn.QueryRow(context.Background(),
		"SELECT 'true' FROM pg_database WHERE datname = $1",
		grpcDatabaseName).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("unable to check whether %v exists: %v", grpcDatabaseUrl, err)
	}
	if exists != "true" {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %v", grpcDatabaseName)
		_, err = defaultDBConn.Exec(context.Background(), createDBQuery)
		if err != nil {
			return nil, fmt.Errorf("unable to create database %v: %v", grpcDatabaseName, err)
		}
	}

	conn, err := pgx.Connect(context.Background(), grpcDatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to %v: %v", grpcDatabaseUrl, err)
	}

	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %v (
			id			SERIAL 		PRIMARY KEY,
			username	TEXT 		NOT NULL,
			time		TIMESTAMP	NOT NULL,
			message		TEXT		NOT NULL
		)
		`, messagesTableName)
	_, err = conn.Exec(context.Background(), createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to create %v table: %v", messagesTableName, err)
	}

	createIndexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %v_time_index ON %v(time)
		`, messagesTableName, messagesTableName)
	_, err = conn.Exec(context.Background(), createIndexQuery)
	if err != nil {
		return nil, fmt.Errorf("unable to create index for time on %v table: %v",
			messagesTableName, err)
	}

	return conn, nil
}
