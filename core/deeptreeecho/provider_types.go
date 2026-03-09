package deeptreeecho

import (
	"context"
	"time"
)

// GenerateOptions contains options for text generation
type GenerateOptions struct {
	MaxTokens        int
	Temperature      float64
	TopP             float64
	Stop             []string
	SystemPrompt     string
	FrequencyPenalty float64
	PresencePenalty  float64
	Model            string
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatOptions contains options for chat completion
type ChatOptions struct {
	MaxTokens        int
	Temperature      float64
	TopP             float64
	Stop             []string
	SystemPrompt     string
	FrequencyPenalty float64
	PresencePenalty  float64
	StopSequences    []string
	Model            string
	GenerateOptions  GenerateOptions
}

// ProviderInfo contains information about an LLM provider
type ProviderInfo struct {
	Name         string
	Type         string
	Available    bool
	MaxTokens    int
	Models       []string
	Description  string
	Capabilities []string
}

// MemoryNode represents a node in the identity memory graph
type MemoryNode struct {
	ID       string
	Content  string
	Strength float64
}

// MemoryEdge represents an edge in the identity memory graph
type MemoryEdge struct {
	From   string
	To     string
	Weight float64
}

// IdentityMemory represents the identity's memory graph
type IdentityMemory struct {
	Coherence float64
	Nodes     map[string]*MemoryNode
	Edges     []*MemoryEdge
	Patterns  []string
}

// IdentityReservoir represents the echo state reservoir in identity
type IdentityReservoir struct {
	Nodes      []float64
	SpectralRadius float64
}

// IdentityEmotionalState represents the emotional state
type IdentityEmotionalState struct {
	Valence   float64
	Arousal   float64
	Dominance float64
}

// IdentitySpatialContext represents spatial awareness
type IdentitySpatialContext struct {
	Location string
	Context  string
}

// IdentityEmbeddings represents vector embeddings
type IdentityEmbeddings struct {
	RepoEmbeddings map[string][]float64
	Dimensions     int
}

// Identity represents the core identity of an echo agent
type Identity struct {
	ID             string
	Name           string
	Essence        string
	CoreValues     []string
	Patterns       map[string]*Pattern
	Traits         map[string]float64
	Memories       []string
	Coherence      float64
	RecursiveDepth int
	Memory         *IdentityMemory
	Reservoir      *IdentityReservoir
	EmotionalState *IdentityEmotionalState
	SpatialContext *IdentitySpatialContext
	Embeddings     *IdentityEmbeddings
	CreatedAt      int64
	UpdatedAt      int64
}

// NewIdentity creates a new identity with the given name and values
func NewIdentity(name string, coreValues []string) *Identity {
	return &Identity{
		ID:         generateID("identity"),
		Name:       name,
		Essence:    "Deep Tree Echo",
		CoreValues: coreValues,
		Patterns:   make(map[string]*Pattern),
		Traits:     make(map[string]float64),
		Memories:   make([]string, 0),
		Memory: &IdentityMemory{
			Nodes:    make(map[string]*MemoryNode),
			Edges:    make([]*MemoryEdge, 0),
			Patterns: make([]string, 0),
		},
		Embeddings: &IdentityEmbeddings{
			RepoEmbeddings: make(map[string][]float64),
		},
		CreatedAt:  currentTimestamp(),
		UpdatedAt:  currentTimestamp(),
	}
}

// ModelProvider defines the interface for model providers
type ModelProvider interface {
	// Name returns the provider name
	Name() string
	
	// Available checks if the provider is available
	Available() bool
	
	// Generate generates text from a prompt
	Generate(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	
	// Chat performs a chat completion
	Chat(ctx context.Context, messages []ChatMessage, opts ChatOptions) (string, error)
	
	// GetInfo returns provider information
	GetInfo() ProviderInfo
}

// Helper functions
func generateID(prefix string) string {
	return prefix + "_" + randomString(16)
}

func currentTimestamp() int64 {
	return time.Now().Unix()
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
