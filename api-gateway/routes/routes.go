package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"preppi.com/api-gateway/clients"
	"preppi.com/api-gateway/handlers"
	"preppi.com/api-gateway/middleware"
	"preppi.com/pkg/auth"
)

type Handlers struct {
	Auth *handlers.AuthHandler
	User *handlers.UserHandler
}

func NewHandlers(conns *clients.GrpcConn) *Handlers {
	return &Handlers{
		Auth: handlers.NewAuthHandler(conns.Auth),
		User: handlers.NewUserHandler(conns.User),
	}
}

// Defines HTTP -> gRPC routing. Split into route groups per domain for maintainability.
func RegisterRoutes(r *gin.Engine, h *Handlers, authManager *auth.Manager, log zerolog.Logger) {
	// Global middleware
	r.Use(middleware.Logging(log), middleware.CORS(), gin.Recovery())

	api := r.Group("/api/v1")

	// Public: auth
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/refresh", h.Auth.Refresh)
	}

	// Protected: everything else requires a valid JWT
	authMw := middleware.Auth(authManager)
	rateLimiter := middleware.NewRateLimiter(60, time.Minute)

	protected := api.Group("")
	protected.Use(authMw, middleware.RateLimit(rateLimiter))
	{
		// Users
		users := protected.Group("/users")
		{
			users.POST("/profile", h.User.CreateProfile)
			users.GET("/profile/:user_id", h.User.GetProfile)
			users.GET("/mentors", h.User.GetMentorsBySubject)
		}
	}

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
