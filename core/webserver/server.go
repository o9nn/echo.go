// Package webserver implements the HTTP API server for Deep Tree Echo using labstack/echo.
// It provides REST endpoints for interacting with the playmate ecosystem, memory system,
// wisdom cultivation, and cognitive functions.
package webserver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// EchoWebServer wraps labstack/echo to provide HTTP API for Deep Tree Echo
type EchoWebServer struct {
	mu sync.RWMutex

	// Echo instance
	echo *echo.Echo

	// Configuration
	Config *ServerConfig

	// State
	running   bool
	startedAt time.Time

	// Handlers (to be wired to ecosystem)
	handlers *APIHandlers
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            int
	Host            string
	EnableCORS      bool
	EnableLogging   bool
	EnableRecover   bool
	EnableRateLimit bool
	RateLimit       int // requests per second
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DefaultServerConfig returns default configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:            8080,
		Host:            "0.0.0.0",
		EnableCORS:      true,
		EnableLogging:   true,
		EnableRecover:   true,
		EnableRateLimit: false,
		RateLimit:       100,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// APIHandlers holds handlers for API endpoints
type APIHandlers struct {
	// Ecosystem handlers
	GetState       func(ctx context.Context) (interface{}, error)
	ControlAction  func(ctx context.Context, action string) (interface{}, error)
	
	// Memory handlers
	AddMemory      func(ctx context.Context, memType, content string, metadata map[string]interface{}) (interface{}, error)
	SearchMemory   func(ctx context.Context, query string, memType string, limit int) (interface{}, error)
	GetMemoryStats func(ctx context.Context) (interface{}, error)
	
	// Playmate handlers
	Interact       func(ctx context.Context, message string) (interface{}, error)
	GetPlaymateState func(ctx context.Context) (interface{}, error)
	RecordWonder   func(ctx context.Context, description, trigger string) (interface{}, error)
	LearnInterest  func(ctx context.Context, name, category, description string) (interface{}, error)
	LearnSkill     func(ctx context.Context, name, description, practice string) (interface{}, error)
	
	// Wisdom handlers
	GetWisdomMetrics func(ctx context.Context) (interface{}, error)
	AddInsight       func(ctx context.Context, content, source string, depth float64) (interface{}, error)
	AddPrinciple     func(ctx context.Context, statement string, dimensions []string, source string) (interface{}, error)
	GetPrinciples    func(ctx context.Context) (interface{}, error)
	
	// Discussion handlers
	StartDiscussion  func(ctx context.Context, topic string) (interface{}, error)
	SendMessage      func(ctx context.Context, discussionID, message string) (interface{}, error)
	EndDiscussion    func(ctx context.Context, discussionID string) (interface{}, error)
	GetDiscussions   func(ctx context.Context) (interface{}, error)
	
	// Cognitive handlers
	Think            func(ctx context.Context, prompt string, depth int) (interface{}, error)
	Introspect       func(ctx context.Context, aspect string) (interface{}, error)
	SpreadActivation func(ctx context.Context, seed string, depth int) (interface{}, error)
}

// NewEchoWebServer creates a new web server
func NewEchoWebServer(config *ServerConfig) *EchoWebServer {
	if config == nil {
		config = DefaultServerConfig()
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	server := &EchoWebServer{
		echo:     e,
		Config:   config,
		handlers: &APIHandlers{},
	}

	// Configure middleware
	server.configureMiddleware()

	// Register routes
	server.registerRoutes()

	return server
}

// configureMiddleware sets up middleware
func (s *EchoWebServer) configureMiddleware() {
	// Recovery middleware
	if s.Config.EnableRecover {
		s.echo.Use(middleware.Recover())
	}

	// Logger middleware
	if s.Config.EnableLogging {
		s.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "${time_rfc3339} ${method} ${uri} ${status} ${latency_human}\n",
		}))
	}

	// CORS middleware
	if s.Config.EnableCORS {
		s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		}))
	}

	// Rate limiter middleware
	if s.Config.EnableRateLimit {
		s.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(s.Config.RateLimit))))
	}

	// Request ID middleware
	s.echo.Use(middleware.RequestID())
}

// registerRoutes sets up API routes
func (s *EchoWebServer) registerRoutes() {
	// Health check
	s.echo.GET("/health", s.handleHealth)
	s.echo.GET("/", s.handleRoot)

	// API v1 group
	api := s.echo.Group("/api/v1")

	// Ecosystem endpoints
	ecosystem := api.Group("/ecosystem")
	ecosystem.GET("/state", s.handleGetState)
	ecosystem.POST("/control", s.handleControl)

	// Memory endpoints
	memory := api.Group("/memory")
	memory.POST("/add", s.handleAddMemory)
	memory.GET("/search", s.handleSearchMemory)
	memory.GET("/stats", s.handleMemoryStats)

	// Playmate endpoints
	playmate := api.Group("/playmate")
	playmate.POST("/interact", s.handleInteract)
	playmate.GET("/state", s.handlePlaymateState)
	playmate.POST("/wonder", s.handleRecordWonder)
	playmate.POST("/interest", s.handleLearnInterest)
	playmate.POST("/skill", s.handleLearnSkill)

	// Wisdom endpoints
	wisdom := api.Group("/wisdom")
	wisdom.GET("/metrics", s.handleWisdomMetrics)
	wisdom.POST("/insight", s.handleAddInsight)
	wisdom.POST("/principle", s.handleAddPrinciple)
	wisdom.GET("/principles", s.handleGetPrinciples)

	// Discussion endpoints
	discussion := api.Group("/discussion")
	discussion.POST("/start", s.handleStartDiscussion)
	discussion.POST("/message", s.handleSendMessage)
	discussion.POST("/end", s.handleEndDiscussion)
	discussion.GET("/list", s.handleGetDiscussions)

	// Cognitive endpoints
	cognitive := api.Group("/cognitive")
	cognitive.POST("/think", s.handleThink)
	cognitive.GET("/introspect", s.handleIntrospect)
	cognitive.POST("/spread-activation", s.handleSpreadActivation)

	// WebSocket endpoint for real-time streaming
	s.echo.GET("/ws", s.handleWebSocket)
}

// SetHandlers sets the API handlers
func (s *EchoWebServer) SetHandlers(handlers *APIHandlers) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = handlers
}

// Start begins the server
func (s *EchoWebServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.startedAt = time.Now()
	s.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	
	// Configure server timeouts
	s.echo.Server.ReadTimeout = s.Config.ReadTimeout
	s.echo.Server.WriteTimeout = s.Config.WriteTimeout

	return s.echo.Start(addr)
}

// StartAsync starts the server in a goroutine
func (s *EchoWebServer) StartAsync() {
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()
}

// Stop gracefully shuts down the server
func (s *EchoWebServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	s.mu.Unlock()

	return s.echo.Shutdown(ctx)
}

// GetEcho returns the underlying echo instance
func (s *EchoWebServer) GetEcho() *echo.Echo {
	return s.echo
}

// Handler implementations

func (s *EchoWebServer) handleRoot(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":        "Deep Tree Echo",
		"description": "Autonomous Wisdom-Cultivating AGI Playmate Ecosystem",
		"version":     "1.0.0",
		"api_version": "v1",
		"endpoints": map[string]string{
			"health":     "/health",
			"ecosystem":  "/api/v1/ecosystem",
			"memory":     "/api/v1/memory",
			"playmate":   "/api/v1/playmate",
			"wisdom":     "/api/v1/wisdom",
			"discussion": "/api/v1/discussion",
			"cognitive":  "/api/v1/cognitive",
			"websocket":  "/ws",
		},
	})
}

func (s *EchoWebServer) handleHealth(c echo.Context) error {
	s.mu.RLock()
	running := s.running
	startedAt := s.startedAt
	s.mu.RUnlock()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "healthy",
		"running":    running,
		"started_at": startedAt,
		"uptime":     time.Since(startedAt).String(),
	})
}

func (s *EchoWebServer) handleGetState(c echo.Context) error {
	if s.handlers.GetState == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetState(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleControl(c echo.Context) error {
	if s.handlers.ControlAction == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.ControlAction(c.Request().Context(), req.Action)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleAddMemory(c echo.Context) error {
	if s.handlers.AddMemory == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Type     string                 `json:"type"`
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.AddMemory(c.Request().Context(), req.Type, req.Content, req.Metadata)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleSearchMemory(c echo.Context) error {
	if s.handlers.SearchMemory == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	query := c.QueryParam("query")
	memType := c.QueryParam("type")
	limit := 10 // default

	result, err := s.handlers.SearchMemory(c.Request().Context(), query, memType, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleMemoryStats(c echo.Context) error {
	if s.handlers.GetMemoryStats == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetMemoryStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleInteract(c echo.Context) error {
	if s.handlers.Interact == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Message string `json:"message"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.Interact(c.Request().Context(), req.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handlePlaymateState(c echo.Context) error {
	if s.handlers.GetPlaymateState == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetPlaymateState(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleRecordWonder(c echo.Context) error {
	if s.handlers.RecordWonder == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Description string `json:"description"`
		Trigger     string `json:"trigger"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.RecordWonder(c.Request().Context(), req.Description, req.Trigger)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleLearnInterest(c echo.Context) error {
	if s.handlers.LearnInterest == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.LearnInterest(c.Request().Context(), req.Name, req.Category, req.Description)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleLearnSkill(c echo.Context) error {
	if s.handlers.LearnSkill == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Practice    string `json:"practice"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.LearnSkill(c.Request().Context(), req.Name, req.Description, req.Practice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleWisdomMetrics(c echo.Context) error {
	if s.handlers.GetWisdomMetrics == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetWisdomMetrics(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleAddInsight(c echo.Context) error {
	if s.handlers.AddInsight == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Content string  `json:"content"`
		Source  string  `json:"source"`
		Depth   float64 `json:"depth"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.AddInsight(c.Request().Context(), req.Content, req.Source, req.Depth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleAddPrinciple(c echo.Context) error {
	if s.handlers.AddPrinciple == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Statement  string   `json:"statement"`
		Dimensions []string `json:"dimensions"`
		Source     string   `json:"source"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.AddPrinciple(c.Request().Context(), req.Statement, req.Dimensions, req.Source)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleGetPrinciples(c echo.Context) error {
	if s.handlers.GetPrinciples == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetPrinciples(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleStartDiscussion(c echo.Context) error {
	if s.handlers.StartDiscussion == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Topic string `json:"topic"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.StartDiscussion(c.Request().Context(), req.Topic)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}

func (s *EchoWebServer) handleSendMessage(c echo.Context) error {
	if s.handlers.SendMessage == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		DiscussionID string `json:"discussion_id"`
		Message      string `json:"message"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.SendMessage(c.Request().Context(), req.DiscussionID, req.Message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleEndDiscussion(c echo.Context) error {
	if s.handlers.EndDiscussion == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		DiscussionID string `json:"discussion_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	result, err := s.handlers.EndDiscussion(c.Request().Context(), req.DiscussionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleGetDiscussions(c echo.Context) error {
	if s.handlers.GetDiscussions == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	result, err := s.handlers.GetDiscussions(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleThink(c echo.Context) error {
	if s.handlers.Think == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Prompt string `json:"prompt"`
		Depth  int    `json:"depth"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if req.Depth == 0 {
		req.Depth = 3
	}

	result, err := s.handlers.Think(c.Request().Context(), req.Prompt, req.Depth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleIntrospect(c echo.Context) error {
	if s.handlers.Introspect == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	aspect := c.QueryParam("aspect")
	if aspect == "" {
		aspect = "all"
	}

	result, err := s.handlers.Introspect(c.Request().Context(), aspect)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleSpreadActivation(c echo.Context) error {
	if s.handlers.SpreadActivation == nil {
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "handler not configured"})
	}

	var req struct {
		Seed  string `json:"seed"`
		Depth int    `json:"depth"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if req.Depth == 0 {
		req.Depth = 3
	}

	result, err := s.handlers.SpreadActivation(c.Request().Context(), req.Seed, req.Depth)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (s *EchoWebServer) handleWebSocket(c echo.Context) error {
	// WebSocket handler placeholder
	// Will be implemented with gorilla/websocket for real-time streaming
	return c.JSON(http.StatusNotImplemented, map[string]string{
		"message": "WebSocket endpoint - coming soon",
	})
}
