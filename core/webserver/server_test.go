// Package webserver_test provides tests for the webserver package.
package webserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/webserver"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

// TestNewEchoWebServer tests the creation of a new web server.
func TestNewEchoWebServer(t *testing.T) {
	config := webserver.DefaultServerConfig()
	server := webserver.NewEchoWebServer(config)

	assert.NotNil(t, server)
	assert.NotNil(t, server.GetEcho())
	assert.Equal(t, config, server.Config)
}

// TestWebServerStartStop tests starting and stopping the web server.
func TestWebServerStartStop(t *testing.T) {
	config := webserver.DefaultServerConfig()
	config.Port = 9090 // Use a different port for testing
	server := webserver.NewEchoWebServer(config)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start server: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check if server is running
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", config.Port))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Stop the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Stop(ctx)
	assert.NoError(t, err)
}

// TestHealthEndpoint tests the /health endpoint.
func TestHealthEndpoint(t *testing.T) {
	config := webserver.DefaultServerConfig()
	server := webserver.NewEchoWebServer(config)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.GetEcho().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var healthResponse map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &healthResponse)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", healthResponse["status"])
}

// TestRootEndpoint tests the root (/) endpoint.
func TestRootEndpoint(t *testing.T) {
	config := webserver.DefaultServerConfig()
	server := webserver.NewEchoWebServer(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	server.GetEcho().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var rootResponse map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &rootResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Deep Tree Echo", rootResponse["name"])
}

// TestEcosystemIntegration tests the integration with the ecosystem.
func TestEcosystemIntegration(t *testing.T) {
	// Create mock ecosystem
	ecoConfig := deeptreeecho.DefaultEcosystemConfig()
	eco, err := deeptreeecho.NewDeepTreeEchoEcosystem(ecoConfig)
	assert.NoError(t, err)

	// Create web server with ecosystem integration
	serverConfig := webserver.DefaultServerConfig()
	webServer := webserver.NewEcosystemWebServer(eco, serverConfig)

	// Test /api/v1/ecosystem/state endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ecosystem/state", nil)
	rec := httptest.NewRecorder()
	webServer.Server.GetEcho().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var stateResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &stateResponse)
	assert.NoError(t, err)
	assert.Equal(t, "DeepTreeEcho", stateResponse["name"])
	assert.Equal(t, "stopped", stateResponse["state"])
}

// TestControlEndpoint tests the /api/v1/ecosystem/control endpoint.
func TestControlEndpoint(t *testing.T) {
	// Create mock ecosystem
	ecoConfig := deeptreeecho.DefaultEcosystemConfig()
	eco, err := deeptreeecho.NewDeepTreeEchoEcosystem(ecoConfig)
	assert.NoError(t, err)

	// Create web server with ecosystem integration
	serverConfig := webserver.DefaultServerConfig()
	webServer := webserver.NewEcosystemWebServer(eco, serverConfig)

	// Test "dream" action
	jsonBody := `{"action": "dream"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ecosystem/control", strings.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	webServer.Server.GetEcho().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var controlResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &controlResponse)
	assert.NoError(t, err)
	assert.Equal(t, "dreaming", controlResponse["status"])
	assert.Equal(t, deeptreeecho.EcosystemDreaming, eco.State)
}

// TestWebSocketHandler tests the WebSocket handler.
func TestWebSocketHandler(t *testing.T) {
	hub := webserver.NewWebSocketHub()
	go hub.Run(context.Background())
	defer hub.Stop()

	e := echo.New()
	e.GET("/ws", webserver.WebSocketHandler(hub))

	server := httptest.NewServer(e)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, err := websocket.Dial(wsURL, "", server.URL)
	assert.NoError(t, err)
	defer ws.Close()

	// Read welcome message
	var msg webserver.WebSocketMessage
	err = websocket.JSON.Receive(ws, &msg)
	assert.NoError(t, err)
	assert.Equal(t, "welcome", msg.Type)

	// Test broadcast
	hub.Broadcast(&webserver.WebSocketMessage{Type: "test", Channel: "all", Data: "hello"})

	// Read broadcast message
	err = websocket.JSON.Receive(ws, &msg)
	assert.NoError(t, err)
	assert.Equal(t, "test", msg.Type)
	assert.Equal(t, "hello", msg.Data)
}

// TestSSEHandler tests the Server-Sent Events handler.
func TestSSEHandler(t *testing.T) {
	hub := webserver.NewWebSocketHub()
	go hub.Run(context.Background())
	defer hub.Stop()

	e := echo.New()
	e.GET("/sse", webserver.SSEHandler(hub))

	server := httptest.NewServer(e)
	defer server.Close()

	resp, err := http.Get(server.URL + "/sse")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	// Read initial event
	buf := make([]byte, 1024)
	n, err := resp.Body.Read(buf)
	assert.NoError(t, err)
	assert.Contains(t, string(buf[:n]), "event: connected")
}
