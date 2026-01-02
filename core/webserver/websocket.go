// Package webserver - websocket.go provides WebSocket support for real-time
// streaming of Echo's stream-of-consciousness and event notifications.
package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

// WebSocketHub manages WebSocket connections and broadcasts
type WebSocketHub struct {
	mu sync.RWMutex

	// Registered clients
	clients map[*WebSocketClient]bool

	// Broadcast channel
	broadcast chan *WebSocketMessage

	// Register channel
	register chan *WebSocketClient

	// Unregister channel
	unregister chan *WebSocketClient

	// Running state
	running bool
	stopCh  chan struct{}
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	hub  *WebSocketHub
	conn *websocket.Conn
	send chan *WebSocketMessage

	// Client info
	ID        string
	ConnectedAt time.Time
	Subscriptions map[string]bool
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Message types
const (
	WSTypeThought     = "thought"
	WSTypeState       = "state"
	WSTypeWonder      = "wonder"
	WSTypeInsight     = "insight"
	WSTypeDiscussion  = "discussion"
	WSTypeMemory      = "memory"
	WSTypeWisdom      = "wisdom"
	WSTypeHeartbeat   = "heartbeat"
	WSTypeSubscribe   = "subscribe"
	WSTypeUnsubscribe = "unsubscribe"
	WSTypeError       = "error"
)

// Channels
const (
	WSChannelAll          = "all"
	WSChannelThoughts     = "thoughts"
	WSChannelState        = "state"
	WSChannelWonders      = "wonders"
	WSChannelInsights     = "insights"
	WSChannelDiscussions  = "discussions"
	WSChannelMemory       = "memory"
	WSChannelWisdom       = "wisdom"
)

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan *WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		stopCh:     make(chan struct{}),
	}
}

// Run starts the hub's main loop
func (h *WebSocketHub) Run(ctx context.Context) {
	h.mu.Lock()
	h.running = true
	h.mu.Unlock()

	// Heartbeat ticker
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return
		case <-h.stopCh:
			h.shutdown()
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		case <-heartbeat.C:
			h.sendHeartbeat()
		}
	}
}

// Stop stops the hub
func (h *WebSocketHub) Stop() {
	h.mu.Lock()
	if !h.running {
		h.mu.Unlock()
		return
	}
	h.running = false
	h.mu.Unlock()
	close(h.stopCh)
}

// shutdown cleans up all clients
func (h *WebSocketHub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.send)
		delete(h.clients, client)
	}
}

// broadcastMessage sends a message to all subscribed clients
func (h *WebSocketHub) broadcastMessage(message *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		// Check if client is subscribed to this channel
		if client.isSubscribed(message.Channel) {
			select {
			case client.send <- message:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// sendHeartbeat sends heartbeat to all clients
func (h *WebSocketHub) sendHeartbeat() {
	h.Broadcast(&WebSocketMessage{
		Type:      WSTypeHeartbeat,
		Channel:   WSChannelAll,
		Data:      map[string]interface{}{"status": "alive"},
		Timestamp: time.Now(),
	})
}

// Broadcast sends a message to all subscribed clients
func (h *WebSocketHub) Broadcast(message *WebSocketMessage) {
	select {
	case h.broadcast <- message:
	default:
		// Broadcast buffer full
	}
}

// BroadcastThought broadcasts a thought
func (h *WebSocketHub) BroadcastThought(thought string) {
	h.Broadcast(&WebSocketMessage{
		Type:      WSTypeThought,
		Channel:   WSChannelThoughts,
		Data:      map[string]interface{}{"thought": thought},
		Timestamp: time.Now(),
	})
}

// BroadcastState broadcasts state change
func (h *WebSocketHub) BroadcastState(state interface{}) {
	h.Broadcast(&WebSocketMessage{
		Type:      WSTypeState,
		Channel:   WSChannelState,
		Data:      state,
		Timestamp: time.Now(),
	})
}

// BroadcastWonder broadcasts a wonder event
func (h *WebSocketHub) BroadcastWonder(description, trigger string) {
	h.Broadcast(&WebSocketMessage{
		Type:    WSTypeWonder,
		Channel: WSChannelWonders,
		Data: map[string]interface{}{
			"description": description,
			"trigger":     trigger,
		},
		Timestamp: time.Now(),
	})
}

// BroadcastInsight broadcasts an insight
func (h *WebSocketHub) BroadcastInsight(content, source string, depth float64) {
	h.Broadcast(&WebSocketMessage{
		Type:    WSTypeInsight,
		Channel: WSChannelInsights,
		Data: map[string]interface{}{
			"content": content,
			"source":  source,
			"depth":   depth,
		},
		Timestamp: time.Now(),
	})
}

// ClientCount returns the number of connected clients
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// isSubscribed checks if client is subscribed to a channel
func (c *WebSocketClient) isSubscribed(channel string) bool {
	if c.Subscriptions[WSChannelAll] {
		return true
	}
	return c.Subscriptions[channel]
}

// Subscribe subscribes to a channel
func (c *WebSocketClient) Subscribe(channel string) {
	c.Subscriptions[channel] = true
}

// Unsubscribe unsubscribes from a channel
func (c *WebSocketClient) Unsubscribe(channel string) {
	delete(c.Subscriptions, channel)
}

// readPump reads messages from the WebSocket connection
func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg WebSocketMessage
		err := websocket.JSON.Receive(c.conn, &msg)
		if err != nil {
			break
		}

		// Handle subscription messages
		switch msg.Type {
		case WSTypeSubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.Subscribe(channel)
			}
		case WSTypeUnsubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.Unsubscribe(channel)
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *WebSocketClient) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		data, err := json.Marshal(message)
		if err != nil {
			continue
		}

		if err := websocket.Message.Send(c.conn, string(data)); err != nil {
			break
		}
	}
}

// WebSocketHandler creates an echo handler for WebSocket connections
func WebSocketHandler(hub *WebSocketHub) echo.HandlerFunc {
	return func(c echo.Context) error {
		websocket.Handler(func(ws *websocket.Conn) {
			client := &WebSocketClient{
				hub:           hub,
				conn:          ws,
				send:          make(chan *WebSocketMessage, 256),
				ID:            fmt.Sprintf("client_%d", time.Now().UnixNano()),
				ConnectedAt:   time.Now(),
				Subscriptions: map[string]bool{WSChannelAll: true}, // Subscribe to all by default
			}

			hub.register <- client

			// Send welcome message
			client.send <- &WebSocketMessage{
				Type:    "welcome",
				Channel: WSChannelAll,
				Data: map[string]interface{}{
					"client_id": client.ID,
					"message":   "Welcome to Deep Tree Echo WebSocket stream",
					"channels":  []string{WSChannelAll, WSChannelThoughts, WSChannelState, WSChannelWonders, WSChannelInsights, WSChannelDiscussions, WSChannelMemory, WSChannelWisdom},
				},
				Timestamp: time.Now(),
			}

			// Start read and write pumps
			go client.writePump()
			client.readPump()
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	}
}

// AddWebSocketToServer adds WebSocket support to an EchoWebServer
func AddWebSocketToServer(server *EchoWebServer, hub *WebSocketHub) {
	server.echo.GET("/ws", WebSocketHandler(hub))
	server.echo.GET("/ws/stream", WebSocketHandler(hub))
}

// StreamingConfig holds configuration for streaming
type StreamingConfig struct {
	BufferSize     int
	HeartbeatInterval time.Duration
	WriteTimeout   time.Duration
}

// DefaultStreamingConfig returns default streaming configuration
func DefaultStreamingConfig() *StreamingConfig {
	return &StreamingConfig{
		BufferSize:        256,
		HeartbeatInterval: 30 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
}

// SSEHandler creates a Server-Sent Events handler for streaming
func SSEHandler(hub *WebSocketHub) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "text/event-stream")
		c.Response().Header().Set("Cache-Control", "no-cache")
		c.Response().Header().Set("Connection", "keep-alive")
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")

		// Create a channel for this client
		messageChan := make(chan *WebSocketMessage, 256)

		// Create a pseudo-client for SSE
		client := &WebSocketClient{
			hub:           hub,
			send:          messageChan,
			ID:            fmt.Sprintf("sse_%d", time.Now().UnixNano()),
			ConnectedAt:   time.Now(),
			Subscriptions: map[string]bool{WSChannelAll: true},
		}

		hub.register <- client
		defer func() {
			hub.unregister <- client
		}()

		// Send initial event
		fmt.Fprintf(c.Response(), "event: connected\ndata: {\"client_id\":\"%s\"}\n\n", client.ID)
		c.Response().Flush()

		// Stream events
		for {
			select {
			case <-c.Request().Context().Done():
				return nil
			case msg := <-messageChan:
				data, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.Response(), "event: %s\ndata: %s\n\n", msg.Type, string(data))
				c.Response().Flush()
			}
		}
	}
}

// AddSSEToServer adds Server-Sent Events support to an EchoWebServer
func AddSSEToServer(server *EchoWebServer, hub *WebSocketHub) {
	server.echo.GET("/sse", SSEHandler(hub))
	server.echo.GET("/events", SSEHandler(hub))
}

// StreamEndpoint represents a streaming endpoint configuration
type StreamEndpoint struct {
	Path        string
	Description string
	Handler     echo.HandlerFunc
}

// GetStreamingEndpoints returns all available streaming endpoints
func GetStreamingEndpoints(hub *WebSocketHub) []StreamEndpoint {
	return []StreamEndpoint{
		{
			Path:        "/ws",
			Description: "WebSocket endpoint for bidirectional real-time communication",
			Handler:     WebSocketHandler(hub),
		},
		{
			Path:        "/ws/stream",
			Description: "WebSocket stream endpoint (alias)",
			Handler:     WebSocketHandler(hub),
		},
		{
			Path:        "/sse",
			Description: "Server-Sent Events endpoint for unidirectional streaming",
			Handler:     SSEHandler(hub),
		},
		{
			Path:        "/events",
			Description: "Events endpoint (SSE alias)",
			Handler:     SSEHandler(hub),
		},
	}
}

// StreamingInfo returns information about streaming capabilities
func StreamingInfo() map[string]interface{} {
	return map[string]interface{}{
		"websocket": map[string]interface{}{
			"endpoints": []string{"/ws", "/ws/stream"},
			"protocol":  "ws://",
			"channels": []string{
				WSChannelAll,
				WSChannelThoughts,
				WSChannelState,
				WSChannelWonders,
				WSChannelInsights,
				WSChannelDiscussions,
				WSChannelMemory,
				WSChannelWisdom,
			},
			"message_types": []string{
				WSTypeThought,
				WSTypeState,
				WSTypeWonder,
				WSTypeInsight,
				WSTypeDiscussion,
				WSTypeMemory,
				WSTypeWisdom,
				WSTypeHeartbeat,
				WSTypeSubscribe,
				WSTypeUnsubscribe,
			},
		},
		"sse": map[string]interface{}{
			"endpoints": []string{"/sse", "/events"},
			"protocol":  "text/event-stream",
		},
	}
}

// HandleStreamingInfo returns an echo handler for streaming info
func HandleStreamingInfo() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, StreamingInfo())
	}
}
