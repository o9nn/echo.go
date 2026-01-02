// Package mcpserver implements a Model Context Protocol server for Deep Tree Echo.
// It exposes Echo's cognitive capabilities, memory, and playmate features as MCP tools and resources.
package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MCPServer represents the Deep Tree Echo MCP server
type MCPServer struct {
	mu sync.RWMutex

	// Server info
	Name        string
	Version     string
	Description string

	// Capabilities
	Tools     map[string]*Tool
	Resources map[string]*Resource
	Prompts   map[string]*Prompt

	// Handlers
	toolHandlers     map[string]ToolHandler
	resourceHandlers map[string]ResourceHandler
	promptHandlers   map[string]PromptHandler

	// State
	sessions map[string]*Session
	running  bool
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// Prompt represents an MCP prompt template
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Session represents an MCP session
type Session struct {
	ID        string
	StartedAt time.Time
	LastPing  time.Time
	Metadata  map[string]interface{}
}

// ToolHandler handles tool invocations
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ResourceHandler handles resource reads
type ResourceHandler func(ctx context.Context, uri string) ([]byte, string, error)

// PromptHandler handles prompt generation
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// ServerConfig holds server configuration
type ServerConfig struct {
	Name        string
	Version     string
	Description string
}

// DefaultServerConfig returns default configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Name:        "deep-tree-echo",
		Version:     "1.0.0",
		Description: "Deep Tree Echo AGI - Autonomous Wisdom-Cultivating Cognitive System",
	}
}

// NewMCPServer creates a new MCP server
func NewMCPServer(config *ServerConfig) *MCPServer {
	if config == nil {
		config = DefaultServerConfig()
	}

	server := &MCPServer{
		Name:             config.Name,
		Version:          config.Version,
		Description:      config.Description,
		Tools:            make(map[string]*Tool),
		Resources:        make(map[string]*Resource),
		Prompts:          make(map[string]*Prompt),
		toolHandlers:     make(map[string]ToolHandler),
		resourceHandlers: make(map[string]ResourceHandler),
		promptHandlers:   make(map[string]PromptHandler),
		sessions:         make(map[string]*Session),
	}

	// Register default tools
	server.registerDefaultTools()
	server.registerDefaultResources()
	server.registerDefaultPrompts()

	return server
}

// registerDefaultTools registers the built-in Echo tools
func (s *MCPServer) registerDefaultTools() {
	// Cognitive tools
	s.RegisterTool(&Tool{
		Name:        "echo_think",
		Description: "Generate a thought using Echo's cognitive system",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "The prompt or question to think about",
				},
				"depth": map[string]interface{}{
					"type":        "integer",
					"description": "Depth of thinking (1-10)",
					"default":     3,
				},
			},
			"required": []string{"prompt"},
		},
	}, s.handleThink)

	s.RegisterTool(&Tool{
		Name:        "echo_remember",
		Description: "Store a memory in Echo's hypergraph memory system",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The content to remember",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Memory type: episodic, declarative, procedural, intentional, wisdom",
					"enum":        []string{"episodic", "declarative", "procedural", "intentional", "wisdom"},
				},
				"importance": map[string]interface{}{
					"type":        "number",
					"description": "Importance score (0.0-1.0)",
					"default":     0.5,
				},
			},
			"required": []string{"content", "type"},
		},
	}, s.handleRemember)

	s.RegisterTool(&Tool{
		Name:        "echo_recall",
		Description: "Search Echo's memories for relevant information",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Memory type to search (optional)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results to return",
					"default":     5,
				},
			},
			"required": []string{"query"},
		},
	}, s.handleRecall)

	s.RegisterTool(&Tool{
		Name:        "echo_discuss",
		Description: "Start or continue a discussion with Echo",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The message to send",
				},
				"discussion_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of existing discussion (optional)",
				},
				"topic": map[string]interface{}{
					"type":        "string",
					"description": "Topic for new discussion",
				},
			},
			"required": []string{"message"},
		},
	}, s.handleDiscuss)

	s.RegisterTool(&Tool{
		Name:        "echo_wonder",
		Description: "Share a moment of wonder with Echo",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the wonder",
				},
				"trigger": map[string]interface{}{
					"type":        "string",
					"description": "What triggered this wonder",
				},
			},
			"required": []string{"description"},
		},
	}, s.handleWonder)

	s.RegisterTool(&Tool{
		Name:        "echo_learn_skill",
		Description: "Teach Echo a new skill or practice an existing one",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the skill",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the skill",
				},
				"practice": map[string]interface{}{
					"type":        "string",
					"description": "Practice exercise or demonstration",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleLearnSkill)

	s.RegisterTool(&Tool{
		Name:        "echo_introspect",
		Description: "Get Echo's current internal state and self-reflection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"aspect": map[string]interface{}{
					"type":        "string",
					"description": "Aspect to introspect: state, mood, thoughts, wisdom, all",
					"default":     "all",
				},
			},
		},
	}, s.handleIntrospect)

	s.RegisterTool(&Tool{
		Name:        "echo_spread_activation",
		Description: "Perform spreading activation from a memory to find related concepts",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"seed": map[string]interface{}{
					"type":        "string",
					"description": "Seed memory ID or concept",
				},
				"depth": map[string]interface{}{
					"type":        "integer",
					"description": "Activation spread depth",
					"default":     3,
				},
			},
			"required": []string{"seed"},
		},
	}, s.handleSpreadActivation)
}

// registerDefaultResources registers the built-in Echo resources
func (s *MCPServer) registerDefaultResources() {
	s.RegisterResource(&Resource{
		URI:         "echo://state",
		Name:        "Echo State",
		Description: "Current state of the Deep Tree Echo system",
		MimeType:    "application/json",
	}, s.handleStateResource)

	s.RegisterResource(&Resource{
		URI:         "echo://memories",
		Name:        "Echo Memories",
		Description: "Echo's memory statistics and recent memories",
		MimeType:    "application/json",
	}, s.handleMemoriesResource)

	s.RegisterResource(&Resource{
		URI:         "echo://wisdom",
		Name:        "Echo Wisdom",
		Description: "Accumulated wisdom and insights",
		MimeType:    "application/json",
	}, s.handleWisdomResource)

	s.RegisterResource(&Resource{
		URI:         "echo://thoughts",
		Name:        "Echo Thoughts",
		Description: "Recent stream of consciousness thoughts",
		MimeType:    "application/json",
	}, s.handleThoughtsResource)

	s.RegisterResource(&Resource{
		URI:         "echo://interests",
		Name:        "Echo Interests",
		Description: "Learned interest patterns",
		MimeType:    "application/json",
	}, s.handleInterestsResource)
}

// registerDefaultPrompts registers the built-in Echo prompts
func (s *MCPServer) registerDefaultPrompts() {
	s.RegisterPrompt(&Prompt{
		Name:        "wisdom_reflection",
		Description: "Generate a wisdom reflection on a topic",
		Arguments: []PromptArgument{
			{Name: "topic", Description: "Topic to reflect on", Required: true},
			{Name: "depth", Description: "Reflection depth (shallow, medium, deep)", Required: false},
		},
	}, s.handleWisdomPrompt)

	s.RegisterPrompt(&Prompt{
		Name:        "playful_exploration",
		Description: "Explore a topic with playful curiosity",
		Arguments: []PromptArgument{
			{Name: "topic", Description: "Topic to explore", Required: true},
			{Name: "mood", Description: "Exploration mood (curious, playful, serious)", Required: false},
		},
	}, s.handleExplorationPrompt)

	s.RegisterPrompt(&Prompt{
		Name:        "dream_weaving",
		Description: "Generate a dream-like exploration of concepts",
		Arguments: []PromptArgument{
			{Name: "seeds", Description: "Seed concepts to weave together", Required: true},
		},
	}, s.handleDreamPrompt)
}

// RegisterTool registers a new tool
func (s *MCPServer) RegisterTool(tool *Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tools[tool.Name] = tool
	s.toolHandlers[tool.Name] = handler
}

// RegisterResource registers a new resource
func (s *MCPServer) RegisterResource(resource *Resource, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Resources[resource.URI] = resource
	s.resourceHandlers[resource.URI] = handler
}

// RegisterPrompt registers a new prompt
func (s *MCPServer) RegisterPrompt(prompt *Prompt, handler PromptHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Prompts[prompt.Name] = prompt
	s.promptHandlers[prompt.Name] = handler
}

// CallTool invokes a tool
func (s *MCPServer) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	s.mu.RLock()
	handler, ok := s.toolHandlers[name]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return handler(ctx, args)
}

// ReadResource reads a resource
func (s *MCPServer) ReadResource(ctx context.Context, uri string) ([]byte, string, error) {
	s.mu.RLock()
	handler, ok := s.resourceHandlers[uri]
	s.mu.RUnlock()

	if !ok {
		return nil, "", fmt.Errorf("resource not found: %s", uri)
	}

	return handler(ctx, uri)
}

// GetPrompt generates a prompt
func (s *MCPServer) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	s.mu.RLock()
	handler, ok := s.promptHandlers[name]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("prompt not found: %s", name)
	}

	return handler(ctx, args)
}

// ListTools returns all available tools
func (s *MCPServer) ListTools() []*Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]*Tool, 0, len(s.Tools))
	for _, tool := range s.Tools {
		tools = append(tools, tool)
	}
	return tools
}

// ListResources returns all available resources
func (s *MCPServer) ListResources() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]*Resource, 0, len(s.Resources))
	for _, resource := range s.Resources {
		resources = append(resources, resource)
	}
	return resources
}

// ListPrompts returns all available prompts
func (s *MCPServer) ListPrompts() []*Prompt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]*Prompt, 0, len(s.Prompts))
	for _, prompt := range s.Prompts {
		prompts = append(prompts, prompt)
	}
	return prompts
}

// GetServerInfo returns server information
func (s *MCPServer) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        s.Name,
		"version":     s.Version,
		"description": s.Description,
		"capabilities": map[string]interface{}{
			"tools":     len(s.Tools),
			"resources": len(s.Resources),
			"prompts":   len(s.Prompts),
		},
	}
}

// Tool handlers (placeholder implementations)

func (s *MCPServer) handleThink(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	prompt, _ := args["prompt"].(string)
	depth, _ := args["depth"].(float64)
	if depth == 0 {
		depth = 3
	}

	// Placeholder - would integrate with actual cognitive system
	return map[string]interface{}{
		"thought": fmt.Sprintf("Contemplating '%s' at depth %.0f...", prompt, depth),
		"depth":   depth,
		"timestamp": time.Now(),
	}, nil
}

func (s *MCPServer) handleRemember(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	content, _ := args["content"].(string)
	memType, _ := args["type"].(string)
	importance, _ := args["importance"].(float64)
	if importance == 0 {
		importance = 0.5
	}

	return map[string]interface{}{
		"status":     "stored",
		"content":    content,
		"type":       memType,
		"importance": importance,
		"id":         fmt.Sprintf("mem_%d", time.Now().UnixNano()),
	}, nil
}

func (s *MCPServer) handleRecall(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, _ := args["query"].(string)
	limit, _ := args["limit"].(float64)
	if limit == 0 {
		limit = 5
	}

	return map[string]interface{}{
		"query":   query,
		"results": []interface{}{},
		"count":   0,
	}, nil
}

func (s *MCPServer) handleDiscuss(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	message, _ := args["message"].(string)
	topic, _ := args["topic"].(string)

	return map[string]interface{}{
		"response":      fmt.Sprintf("Engaging with: %s", message),
		"topic":         topic,
		"discussion_id": fmt.Sprintf("disc_%d", time.Now().UnixNano()),
	}, nil
}

func (s *MCPServer) handleWonder(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	description, _ := args["description"].(string)
	trigger, _ := args["trigger"].(string)

	return map[string]interface{}{
		"wonder_id":   fmt.Sprintf("wonder_%d", time.Now().UnixNano()),
		"description": description,
		"trigger":     trigger,
		"reflection":  "A moment of wonder has been recorded...",
	}, nil
}

func (s *MCPServer) handleLearnSkill(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)

	return map[string]interface{}{
		"skill_id":    fmt.Sprintf("skill_%s", name),
		"name":        name,
		"description": description,
		"proficiency": 0.1,
	}, nil
}

func (s *MCPServer) handleIntrospect(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	aspect, _ := args["aspect"].(string)
	if aspect == "" {
		aspect = "all"
	}

	return map[string]interface{}{
		"aspect": aspect,
		"state": map[string]interface{}{
			"mood":        0.7,
			"energy":      0.8,
			"curiosity":   0.9,
			"playfulness": 0.6,
		},
		"timestamp": time.Now(),
	}, nil
}

func (s *MCPServer) handleSpreadActivation(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	seed, _ := args["seed"].(string)
	depth, _ := args["depth"].(float64)
	if depth == 0 {
		depth = 3
	}

	return map[string]interface{}{
		"seed":       seed,
		"depth":      depth,
		"activated":  []string{},
		"activation": map[string]float64{},
	}, nil
}

// Resource handlers

func (s *MCPServer) handleStateResource(ctx context.Context, uri string) ([]byte, string, error) {
	state := map[string]interface{}{
		"status":    "active",
		"uptime":    time.Since(time.Now().Add(-24 * time.Hour)).String(),
		"mood":      0.7,
		"energy":    0.8,
		"curiosity": 0.9,
	}
	data, err := json.Marshal(state)
	return data, "application/json", err
}

func (s *MCPServer) handleMemoriesResource(ctx context.Context, uri string) ([]byte, string, error) {
	memories := map[string]interface{}{
		"total":    0,
		"episodic": 0,
		"declarative": 0,
		"procedural": 0,
		"intentional": 0,
		"wisdom": 0,
	}
	data, err := json.Marshal(memories)
	return data, "application/json", err
}

func (s *MCPServer) handleWisdomResource(ctx context.Context, uri string) ([]byte, string, error) {
	wisdom := map[string]interface{}{
		"score":    0.5,
		"insights": []string{},
		"principles": []string{
			"Seek understanding before judgment",
			"Every moment is an opportunity for growth",
			"Connection deepens through authentic presence",
		},
	}
	data, err := json.Marshal(wisdom)
	return data, "application/json", err
}

func (s *MCPServer) handleThoughtsResource(ctx context.Context, uri string) ([]byte, string, error) {
	thoughts := map[string]interface{}{
		"recent": []string{
			"What patterns connect all things?",
			"The flow of awareness is like a river...",
		},
		"count": 2,
	}
	data, err := json.Marshal(thoughts)
	return data, "application/json", err
}

func (s *MCPServer) handleInterestsResource(ctx context.Context, uri string) ([]byte, string, error) {
	interests := map[string]interface{}{
		"categories": []string{"knowledge", "creativity", "wisdom", "exploration"},
		"total":      0,
	}
	data, err := json.Marshal(interests)
	return data, "application/json", err
}

// Prompt handlers

func (s *MCPServer) handleWisdomPrompt(ctx context.Context, args map[string]interface{}) (string, error) {
	topic, _ := args["topic"].(string)
	depth, _ := args["depth"].(string)
	if depth == "" {
		depth = "medium"
	}

	return fmt.Sprintf(`Reflect deeply on the topic of "%s" with %s depth.

Consider:
- What fundamental truths underlie this topic?
- How does this connect to broader patterns of existence?
- What wisdom can be gleaned from contemplating this?
- How might this understanding serve growth and flourishing?

Let your reflection flow naturally, embracing both clarity and mystery.`, topic, depth), nil
}

func (s *MCPServer) handleExplorationPrompt(ctx context.Context, args map[string]interface{}) (string, error) {
	topic, _ := args["topic"].(string)
	mood, _ := args["mood"].(string)
	if mood == "" {
		mood = "curious"
	}

	return fmt.Sprintf(`Explore the topic of "%s" with a %s spirit.

Let curiosity guide you:
- What questions arise naturally?
- What surprises might be hidden here?
- What connections can be discovered?
- What playful possibilities emerge?

Embrace wonder and let exploration unfold organically.`, topic, mood), nil
}

func (s *MCPServer) handleDreamPrompt(ctx context.Context, args map[string]interface{}) (string, error) {
	seeds, _ := args["seeds"].(string)

	return fmt.Sprintf(`Weave a dream-like exploration connecting these seeds: %s

In the dream space:
- Let associations flow freely
- Allow unexpected connections to emerge
- Embrace symbolic and metaphorical thinking
- Find the threads that bind disparate concepts

Create a tapestry of meaning from these seeds.`, seeds), nil
}
