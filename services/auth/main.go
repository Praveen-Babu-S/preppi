package main

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"preppi.com/pkg/auth"
	"preppi.com/pkg/config"
	"preppi.com/pkg/database"
	"preppi.com/pkg/logger"
	"preppi.com/pkg/middleware"
	pb "preppi.com/proto/auth/v1"
	"preppi.com/services/auth/handler"
	"preppi.com/services/auth/repository"
	"preppi.com/services/auth/service"
)

func main() {
	cfg := config.Load("auth")
	log := logger.New(cfg.LogLevel)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	accessTTL := time.Duration(cfg.JWT.AccessTTL) * time.Minute
	refreshTTL := time.Duration(cfg.JWT.RefreshTTL) * time.Hour
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 24 * time.Hour
	}

	tokenManager, err := auth.NewManager(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, accessTTL, refreshTTL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init jwt manager")
	}

	repo := repository.New(db)
	svc := service.New(repo, tokenManager)
	h := handler.New(svc)

	port := cfg.GRPCPort
	if port == 0 {
		port = 50051
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.RecoveryInterceptor(),
			middleware.LoggingInterceptor(log),
		),
	)
	pb.RegisterAuthServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	log.Info().Int("port", port).Msg("auth service starting")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
