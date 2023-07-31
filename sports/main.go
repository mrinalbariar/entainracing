package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	grpcSportsEndpoint = flag.String("grpc-sports-endpoint", "localhost:7000", "gRPC sports server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc sports server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":7000")
	if err != nil {
		return err
	}

	sportsDB, err := sql.Open("sqlite3", "./db/sports.db")
	if err != nil {
		return err
	}

	sportsRepo := db.NewSportsRepo(sportsDB)
	if err := sportsRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	sports.RegisterSportsServer(
		grpcServer,
		service.NewSportsService(
			sportsRepo,
		),
	)

	log.Printf("gRPC server listening on: %s\n", *grpcSportsEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}
