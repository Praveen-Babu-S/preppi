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
	pb "preppi.com/proto/chat/v1"
	"preppi.com/services/chat/handler"
	"preppi.com/services/chat/repository"
	"preppi.com/services/chat/service"
)

func main() {
	cfg := config.Load("chat")
	log := logger.New(cfg.LogLevel)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	tokenManager, err := auth.NewManager(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, 15*time.Minute, 24*time.Hour)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init jwt manager")
	}

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	port := cfg.GRPCPort
	if port == 0 {
		port = 50056
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.RecoveryInterceptor(),
			middleware.LoggingInterceptor(log),
			middleware.AuthInterceptor(tokenManager),
		),
		grpc.MaxConcurrentStreams(1000),
	)
	pb.RegisterChatServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	log.Info().Int("port", port).Msg("chat service starting")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
