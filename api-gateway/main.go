package main

import (
	"time"

	"github.com/gin-gonic/gin"

	"preppi.com/api-gateway/clients"
	"preppi.com/api-gateway/routes"
	"preppi.com/pkg/auth"
	"preppi.com/pkg/config"
	"preppi.com/pkg/logger"
)

func main() {
	cfg := config.Load("gateway")
	log := logger.New(cfg.LogLevel)

	// Map of service addresses. In production these come from env/K8s service discovery.
	// Local defaults point to docker-compose service names.
	addrs := clients.Addrs{
		Auth:          envOr("AUTH_ADDR", "localhost:50051"),
		User:          envOr("USER_ADDR", "localhost:50052"),
		Doubt:         envOr("DOUBT_ADDR", "localhost:50053"),
		Matching:      envOr("MATCHING_ADDR", "localhost:50054"),
		Knowledgebase: envOr("KNOWLEDGEBASE_ADDR", "localhost:50058"),
		Analytics:     envOr("ANALYTICS_ADDR", "localhost:50059"),
		Admin:         envOr("ADMIN_ADDR", "localhost:50060"),
	}

	conns, err := clients.New(addrs)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to services")
	}
	defer conns.Close()

	tokenManager, err := auth.NewManager(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, 15*time.Minute, 24*time.Hour)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init jwt manager")
	}

	handlers := routes.NewHandlers(conns)

	r := gin.New()
	routes.RegisterRoutes(r, handlers, tokenManager, log)

	port := cfg.HTTPPort
	if port == 0 {
		port = 8080
	}

	log.Info().Int("port", port).Msg("api gateway starting")
	if err := r.Run(addr(port)); err != nil {
		log.Fatal().Err(err).Msg("gateway failed")
	}
}
