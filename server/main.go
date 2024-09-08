package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	helloworldpb "server/gen"
)

var (
	errUnauthenticated = status.Errorf(codes.Unauthenticated, "unauthenticated")
	secret             = []byte("secret")
)

func timeoutInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	return handler(ctx, req)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	token, err := CreateToken("email|02c013b0493d4a971d6959fe")
	if err != nil {
		os.Exit(1)
	}
	fmt.Println(token)

	dsn := "postgres:postgres@docker.for.mac.localhost:5432/postgres"
	urlExample := fmt.Sprintf("postgres://%s", dsn)
	dbpool, err := pgxpool.New(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	m, err := migrate.New(
		"file://./migrations",
		fmt.Sprintf("pgx5://%s", dsn))
	if err != nil {
		logger.Error("Failed to listen", slog.Any("error", err))
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			logger.Error("Migration:", slog.Any("error", err))
		} else {
			logger.Info("Migration: no change")
		}
	}

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer(
		grpc.WriteBufferSize(1024*1024),  // increase from 32kB -> 1MB
		grpc.ReadBufferSize(1024*1024),   // increase from 32kB -> 1MB
		grpc.MaxConcurrentStreams(10000), // default is 100.  10k is likely too high
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:              5 * time.Minute,  // GRPC Ping the client if it is idle for 60 seconds to ensure the connection is still active (default is 2 * time.Hour)
				Timeout:           10 * time.Second, // Wait 10 seconds for the ping ack before assuming the connection is dead.  (10 seconds may be too long??!!)
				MaxConnectionIdle: 20 * time.Minute, // If a client is idle for 20 minutes, send a GOAWAY (default infinity)
			},
		),
		grpc.ChainUnaryInterceptor(
			timeoutInterceptor,
			AuthorizationInterceptor(dbpool),
		),
	)
	// Attach the Greeter service to the server
	helloworldpb.RegisterGreeterServer(s, NewServer(dbpool))
	// Serve gRPC server
	logger.Info("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(s.Serve(lis))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	client_conn, err := grpc.NewClient(
		"0.0.0.0:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = helloworldpb.RegisterGreeterHandler(context.Background(), gwmux, client_conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
