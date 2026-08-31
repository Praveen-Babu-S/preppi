---
name: api-endpoint
description: Use when adding, modifying, or removing client-facing REST endpoints in the doubt-resolver API gateway. Trigger on requests like "add an endpoint", "create a new API route", "expose an RPC over REST", "add a route for X". Creates/updates Gin handlers, routes, middleware, and maps to gRPC service calls.
---

# API Endpoint

## Purpose

Add client-facing REST endpoints in the Gin API gateway that map to internal gRPC service calls.
The gateway is the single entry point for web/mobile clients.

## When to Use

Use this skill when the user asks to:
- Add a new REST endpoint
- Expose a gRPC RPC to clients
- Modify/remove an existing endpoint
- Add a route for a new feature

## Gateway Structure

```
api-gateway/
├── main.go                     # Gateway entry point: Gin engine + middleware setup
├── routes/                     # Route registry (all route groups)
│   └── question_routes.go      # Route registration per domain
├── middleware/                 # auth, rate-limit, cors, logging
│   ├── auth.go                 # JWT validation
│   ├── rate_limit.go           # Per-IP rate limiting
│   ├── cors.go                 # CORS headers
│   └── logging.go              # Request logging + request ID
└── handlers/                   # HTTP handlers -> gRPC calls
    └── question_handler.go     # REST handler calling QuestionService
```

## Endpoint Template

### 1. Handler (`api-gateway/handlers/question_handler.go`)

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"

    pb "github.com/<org>/doubt-resolver/proto/question/v1"
)

type QuestionHandler struct {
    questionClient pb.QuestionServiceClient
}

func NewQuestionHandler(conn *grpc.ClientConn) *QuestionHandler {
    return &QuestionHandler{questionClient: pb.NewQuestionServiceClient(conn)}
}

// CreateQuestion handles POST /api/v1/questions
func (h *QuestionHandler) CreateQuestion(c *gin.Context) {
    var req struct {
        Subject     string   `json:"subject" binding:"required"`
        Description string   `json:"description" binding:"required"`
        ImageURLs   []string `json:"image_urls"`
        Urgency     string   `json:"urgency"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    studentID := c.GetString("user_id") // from auth middleware
    resp, err := h.questionClient.CreateQuestion(c, &pb.CreateQuestionRequest{
        StudentId:   studentID,
        Subject:     req.Subject,
        Description: req.Description,
        ImageUrls:   req.ImageURLs,
        Urgency:     req.Urgency,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"id": resp.GetId(), "status": resp.GetStatus()})
}
```

### 2. Route Registration (`api-gateway/routes/question_routes.go`)

```go
package routes

import (
    "github.com/gin-gonic/gin"

    "github.com/<org>/doubt-resolver/api-gateway/handlers"
    "github.com/<org>/doubt-resolver/api-gateway/middleware"
)

func RegisterQuestionRoutes(r *gin.Engine, h *handlers.QuestionHandler) {
    v1 := r.Group("/api/v1")
    v1.Use(middleware.Auth())
    {
        questions := v1.Group("/questions")
        questions.Use(middleware.RateLimit("question", 10, time.Minute))
        {
            questions.POST("", h.CreateQuestion)
            questions.GET("/:id", h.GetQuestion)
            questions.GET("", h.ListQuestions)
            questions.GET("/search", h.SearchQuestions)
        }
    }
}
```

### 3. Route Registry (`api-gateway/routes/router.go`)

```go
package routes

import "github.com/gin-gonic/gin"

func SetupRouter(handlers *handlers.AllHandlers) *gin.Engine {
    r := gin.New()
    r.Use(middleware.Logging(), middleware.CORS(), gin.Recovery())

    RegisterAuthRoutes(r, handlers.Auth)
    RegisterUserRoutes(r, handlers.User)
    RegisterQuestionRoutes(r, handlers.Question)
    // ... register all route groups

    return r
}
```

## Conventions

### REST Naming
- Resource-based, plural nouns: `/api/v1/questions`, `/api/v1/solutions`
- HTTP verbs map to actions: POST (create), GET (read), PUT (update), DELETE (remove)
- Nested resources for sub-resources: `/api/v1/questions/:id/solutions`
- Query params for filtering/pagination: `?page=1&page_size=20&status=open`

### Status Codes
- `200 OK` — successful read/update/delete
- `201 Created` — resource created
- `400 Bad Request` — validation error
- `401 Unauthorized` — missing/invalid JWT
- `403 Forbidden` — wrong role
- `404 Not Found` — resource missing
- `409 Conflict` — duplicate resource
- `429 Too Many Requests` — rate limited
- `500 Internal Server Error` — gRPC call failed

### Error Response Shape
Always return errors consistently:
```json
{ "error": { "code": "VALIDATION_ERROR", "message": "description is required" } }
```

### Middleware Chain
Every endpoint goes through: `RequestID → Logging → CORS → Auth (protected routes) → RateLimit → Handler`

### Auth
- Public: registration, login, health checks — no auth middleware
- Protected: everything else — `middleware.Auth()` extracts user_id and role from JWT
- Role restrictions enforced in middleware or handler (student vs mentor permissions)

### Routing to gRPC
- Gateway maintains gRPC client connections to each service
- Handlers call `service.GetXxx(c, &pb.Request{...})` and map the gRPC response to JSON
- Errors from gRPC are translated to HTTP status codes

### Request IDs
- Accept `X-Request-ID` header if present, else generate one
- Propagate via gRPC metadata to services for tracing

## Rules

- **No business logic in handlers** — handlers only: parse request, call gRPC, format response
- Validation happens in handler (binding) and again minimally in the gRPC service
- Always use bounded rate limits on public endpoints
- Return consistent error shapes
- Keep response payloads minimal — don't expose internal fields

## After Creating

Remind the user to:
- Test with `curl` or the API client
- Run `make build` and `make test` to verify
- Update OpenAPI/Swagger docs if present
