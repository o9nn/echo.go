// Package webserver - ecosystem_integration.go provides integration between
// the labstack/echo web server and the Deep Tree Echo ecosystem.
package webserver

import (
	"context"
	"fmt"

	"github.com/cogpy/echo9llama/core/deeptreeecho"
	"github.com/cogpy/echo9llama/core/playmate"
	"github.com/cogpy/echo9llama/core/vectormem"
)

// EcosystemWebServer integrates the web server with the Deep Tree Echo ecosystem
type EcosystemWebServer struct {
	Server    *EchoWebServer
	Ecosystem *deeptreeecho.DeepTreeEchoEcosystem
}

// NewEcosystemWebServer creates a new integrated web server
func NewEcosystemWebServer(ecosystem *deeptreeecho.DeepTreeEchoEcosystem, config *ServerConfig) *EcosystemWebServer {
	if config == nil {
		config = DefaultServerConfig()
	}

	server := NewEchoWebServer(config)
	
	ews := &EcosystemWebServer{
		Server:    server,
		Ecosystem: ecosystem,
	}

	// Wire up handlers
	ews.wireHandlers()

	return ews
}

// wireHandlers connects the ecosystem to the API handlers
func (ews *EcosystemWebServer) wireHandlers() {
	handlers := &APIHandlers{
		// Ecosystem handlers
		GetState:      ews.handleGetState,
		ControlAction: ews.handleControlAction,

		// Memory handlers
		AddMemory:      ews.handleAddMemory,
		SearchMemory:   ews.handleSearchMemory,
		GetMemoryStats: ews.handleGetMemoryStats,

		// Playmate handlers
		Interact:         ews.handleInteract,
		GetPlaymateState: ews.handleGetPlaymateState,
		RecordWonder:     ews.handleRecordWonder,
		LearnInterest:    ews.handleLearnInterest,
		LearnSkill:       ews.handleLearnSkill,

		// Wisdom handlers
		GetWisdomMetrics: ews.handleGetWisdomMetrics,
		AddInsight:       ews.handleAddInsight,
		AddPrinciple:     ews.handleAddPrinciple,
		GetPrinciples:    ews.handleGetPrinciples,

		// Discussion handlers
		StartDiscussion: ews.handleStartDiscussion,
		SendMessage:     ews.handleSendMessage,
		EndDiscussion:   ews.handleEndDiscussion,
		GetDiscussions:  ews.handleGetDiscussions,

		// Cognitive handlers
		Think:            ews.handleThink,
		Introspect:       ews.handleIntrospect,
		SpreadActivation: ews.handleSpreadActivation,
	}

	ews.Server.SetHandlers(handlers)
}

// Start begins the web server
func (ews *EcosystemWebServer) Start() error {
	return ews.Server.Start()
}

// StartAsync starts the web server in a goroutine
func (ews *EcosystemWebServer) StartAsync() {
	ews.Server.StartAsync()
}

// Stop gracefully shuts down the web server
func (ews *EcosystemWebServer) Stop(ctx context.Context) error {
	return ews.Server.Stop(ctx)
}

// Handler implementations

func (ews *EcosystemWebServer) handleGetState(ctx context.Context) (interface{}, error) {
	return ews.Ecosystem.GetState(), nil
}

func (ews *EcosystemWebServer) handleControlAction(ctx context.Context, action string) (interface{}, error) {
	switch action {
	case "start":
		if err := ews.Ecosystem.Start(ctx); err != nil {
			return nil, err
		}
		return map[string]string{"status": "started"}, nil
	case "stop":
		ews.Ecosystem.Stop()
		return map[string]string{"status": "stopped"}, nil
	case "dream":
		ews.Ecosystem.EnterDreamState()
		return map[string]string{"status": "dreaming"}, nil
	case "reflect":
		ews.Ecosystem.EnterReflectionState()
		return map[string]string{"status": "reflecting"}, nil
	case "wake":
		ews.Ecosystem.Wake()
		return map[string]string{"status": "awake"}, nil
	case "save":
		if err := ews.Ecosystem.SaveAll(); err != nil {
			return nil, err
		}
		return map[string]string{"status": "saved"}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (ews *EcosystemWebServer) handleAddMemory(ctx context.Context, memType, content string, metadata map[string]interface{}) (interface{}, error) {
	var mt vectormem.MemoryType
	switch memType {
	case "episodic":
		mt = vectormem.EpisodicMemory
	case "declarative":
		mt = vectormem.DeclarativeMemory
	case "procedural":
		mt = vectormem.ProceduralMemory
	case "intentional":
		mt = vectormem.IntentionalMemory
	case "wisdom":
		mt = vectormem.WisdomMemory
	default:
		mt = vectormem.DeclarativeMemory
	}

	mem, err := ews.Ecosystem.Memory.Add(ctx, mt, content, metadata)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":      mem.ID,
		"type":    memType,
		"content": content,
		"status":  "stored",
	}, nil
}

func (ews *EcosystemWebServer) handleSearchMemory(ctx context.Context, query string, memType string, limit int) (interface{}, error) {
	var mt vectormem.MemoryType
	switch memType {
	case "episodic":
		mt = vectormem.EpisodicMemory
	case "declarative":
		mt = vectormem.DeclarativeMemory
	case "procedural":
		mt = vectormem.ProceduralMemory
	case "intentional":
		mt = vectormem.IntentionalMemory
	case "wisdom":
		mt = vectormem.WisdomMemory
	default:
		mt = "" // Search all
	}

	results, err := ews.Ecosystem.Memory.Query(ctx, query, mt, limit)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"query":   query,
		"type":    memType,
		"results": results,
		"count":   len(results),
	}, nil
}

func (ews *EcosystemWebServer) handleGetMemoryStats(ctx context.Context) (interface{}, error) {
	return ews.Ecosystem.Memory.GetStats(), nil
}

func (ews *EcosystemWebServer) handleInteract(ctx context.Context, message string) (interface{}, error) {
	response, err := ews.Ecosystem.Interact(ctx, message)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"input":    message,
		"response": response,
	}, nil
}

func (ews *EcosystemWebServer) handleGetPlaymateState(ctx context.Context) (interface{}, error) {
	return ews.Ecosystem.Playmate.GetState(), nil
}

func (ews *EcosystemWebServer) handleRecordWonder(ctx context.Context, description, trigger string) (interface{}, error) {
	ews.Ecosystem.RecordWonder(description, trigger)
	return map[string]interface{}{
		"description": description,
		"trigger":     trigger,
		"status":      "recorded",
	}, nil
}

func (ews *EcosystemWebServer) handleLearnInterest(ctx context.Context, name, category, description string) (interface{}, error) {
	interest := ews.Ecosystem.Playmate.LearnInterest(playmate.InterestCategory(category), name, []string{description})
	return map[string]interface{}{
		"id":          interest.ID,
		"name":        interest.Topic,
		"category":    interest.Category,
		"description": interest.Keywords[0],
		"status":      "learned",
	}, nil
}

func (ews *EcosystemWebServer) handleLearnSkill(ctx context.Context, name, description, practice string) (interface{}, error) {
	skill := ews.Ecosystem.Playmate.LearnSkill(name, description)
	return map[string]interface{}{
		"id":          skill.ID,
		"name":        skill.Name,
		"description": skill.Description,
		"proficiency": skill.Proficiency,
		"status":      "learned",
	}, nil
}

func (ews *EcosystemWebServer) handleGetWisdomMetrics(ctx context.Context) (interface{}, error) {
	return ews.Ecosystem.Wisdom.GetMetrics(), nil
}

func (ews *EcosystemWebServer) handleAddInsight(ctx context.Context, content, source string, depth float64) (interface{}, error) {
	insight := ews.Ecosystem.Wisdom.AddInsight(ctx, content, source, depth)
	return map[string]interface{}{
		"id":      insight.ID,
		"content": insight.Content,
		"source":  source,
		"depth":   insight.Depth,
		"status":  "added",
	}, nil
}

func (ews *EcosystemWebServer) handleAddPrinciple(ctx context.Context, statement string, dimensions []string, source string) (interface{}, error) {
	// Convert string dimensions to WisdomDimension
	dims := make([]playmate.WisdomDimension, len(dimensions))
	for i, d := range dimensions {
		dims[i] = playmate.WisdomDimension(d)
	}

	principle := ews.Ecosystem.Wisdom.AddPrinciple(statement, dims, source)
	return map[string]interface{}{
		"id":         principle.ID,
		"statement":  principle.Statement,
		"dimensions": principle.Dimensions,
		"source":     principle.Source,
		"status":     "added",
	}, nil
}

func (ews *EcosystemWebServer) handleGetPrinciples(ctx context.Context) (interface{}, error) {
	principles := ews.Ecosystem.Wisdom.GetPrinciples()
	return map[string]interface{}{
		"principles": principles,
		"count":      len(principles),
	}, nil
}

func (ews *EcosystemWebServer) handleStartDiscussion(ctx context.Context, topic string) (interface{}, error) {
	discussion := ews.Ecosystem.Playmate.StartDiscussion(topic, "api_user")
	return map[string]interface{}{
		"id":     discussion.ID,
		"topic":  discussion.Topic,
		"status": "started",
	}, nil
}

func (ews *EcosystemWebServer) handleSendMessage(ctx context.Context, discussionID, message string) (interface{}, error) {
	response, err := ews.Ecosystem.Playmate.SendMessage(discussionID, "api_user", message)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"discussion_id": discussionID,
		"input":         message,
		"response":      response,
	}, nil
}

func (ews *EcosystemWebServer) handleEndDiscussion(ctx context.Context, discussionID string) (interface{}, error) {
	if err := ews.Ecosystem.Playmate.EndDiscussion(discussionID); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"discussion_id": discussionID,
		"status":        "ended",
	}, nil
}

func (ews *EcosystemWebServer) handleGetDiscussions(ctx context.Context) (interface{}, error) {
	discussions := ews.Ecosystem.Playmate.GetActiveDiscussions()
	return map[string]interface{}{
		"discussions": discussions,
		"count":       len(discussions),
	}, nil
}

func (ews *EcosystemWebServer) handleThink(ctx context.Context, prompt string, depth int) (interface{}, error) {
	// Placeholder - would integrate with LLM system
	return map[string]interface{}{
		"prompt":    prompt,
		"depth":     depth,
		"thought":   fmt.Sprintf("Contemplating '%s' at depth %d...", prompt, depth),
		"timestamp": "now",
	}, nil
}

func (ews *EcosystemWebServer) handleIntrospect(ctx context.Context, aspect string) (interface{}, error) {
	state := ews.Ecosystem.GetState()
	
	switch aspect {
	case "state":
		return map[string]interface{}{
			"aspect": "state",
			"data":   state["state"],
		}, nil
	case "wisdom":
		return map[string]interface{}{
			"aspect": "wisdom",
			"data":   state["wisdom_metrics"],
		}, nil
	case "playmate":
		return map[string]interface{}{
			"aspect": "playmate",
			"data":   state["playmate_state"],
		}, nil
	case "memory":
		return map[string]interface{}{
			"aspect": "memory",
			"data":   state["memory_stats"],
		}, nil
	default:
		return state, nil
	}
}

func (ews *EcosystemWebServer) handleSpreadActivation(ctx context.Context, seed string, depth int) (interface{}, error) {
	results, err := ews.Ecosystem.Memory.SpreadActivation(ctx, seed, depth, 0.1)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"seed":      seed,
		"depth":     depth,
		"activated": results,
		"count":     len(results),
	}, nil
}
