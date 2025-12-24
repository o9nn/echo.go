package deeptreeecho

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// SYS6 PAYLOAD TYPES
// =============================================================================
//
// This module defines the concrete payload structures that flow through the
// sys6 pipeline. The payloads are designed to carry cognitive content through
// the Clock30, C₈, K₉, φ, and σ components.
//
// Two primary payload types:
//   1. CognitiveToken: Discrete units of cognitive content (thoughts, percepts)
//   2. GraphMessage: Relational structures connecting cognitive elements
//
// =============================================================================

// =============================================================================
// COGNITIVE TOKEN PAYLOAD
// =============================================================================

// CognitiveToken represents a discrete unit of cognitive content that flows
// through the sys6 pipeline. Tokens carry semantic content along with metadata
// about their processing state and transformations.
type CognitiveToken struct {
	// Identity
	ID        string    `json:"id"`
	ParentID  string    `json:"parent_id,omitempty"`  // For derived tokens
	CreatedAt time.Time `json:"created_at"`

	// Content
	Content     TokenContent `json:"content"`
	ContentType TokenType    `json:"content_type"`

	// Semantic properties
	Salience    float64            `json:"salience"`     // 0.0-1.0, how attention-grabbing
	Relevance   float64            `json:"relevance"`    // 0.0-1.0, how goal-relevant
	Valence     float64            `json:"valence"`      // -1.0 to 1.0, emotional tone
	Confidence  float64            `json:"confidence"`   // 0.0-1.0, certainty level
	Tags        []string           `json:"tags"`
	Embeddings  []float64          `json:"embeddings,omitempty"` // Vector representation

	// Pipeline state (where the token is in sys6)
	PipelineState TokenPipelineState `json:"pipeline_state"`

	// Processing history
	Transformations []TokenTransformation `json:"transformations"`

	// C₈ state: Results from 8-way parallel processing
	C8Results [8]*C8TokenResult `json:"c8_results,omitempty"`

	// K₉ state: Results from 9-phase convolution
	K9Results [9]*K9TokenResult `json:"k9_results,omitempty"`

	// φ state: Delay fold integration results
	PhiResult *PhiTokenResult `json:"phi_result,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenContent holds the actual content of a cognitive token
type TokenContent struct {
	// Text representation
	Text string `json:"text"`

	// Structured data (optional)
	Data map[string]interface{} `json:"data,omitempty"`

	// References to other tokens or graph nodes
	References []string `json:"references,omitempty"`

	// Source information
	Source     string `json:"source"`      // "perception", "memory", "inference", "external"
	SourceID   string `json:"source_id,omitempty"`
}

// TokenType categorizes the cognitive content
type TokenType string

const (
	TokenTypePercept    TokenType = "percept"     // Sensory input
	TokenTypeThought    TokenType = "thought"     // Internal cognition
	TokenTypeMemory     TokenType = "memory"      // Retrieved memory
	TokenTypeGoal       TokenType = "goal"        // Goal state
	TokenTypeAction     TokenType = "action"      // Action intention
	TokenTypeEmotion    TokenType = "emotion"     // Emotional state
	TokenTypeQuery      TokenType = "query"       // Question/inquiry
	TokenTypeResponse   TokenType = "response"    // Answer/response
	TokenTypeInsight    TokenType = "insight"     // Derived understanding
	TokenTypeAffordance TokenType = "affordance"  // Perceived opportunity
	TokenTypeSalience   TokenType = "salience"    // Attention marker
)

// TokenPipelineState tracks where a token is in the sys6 pipeline
type TokenPipelineState struct {
	// Clock30 position
	GlobalStep    int `json:"global_step"`    // 1-30
	DyadicPhase   int `json:"dyadic_phase"`   // 1-2
	TriadicPhase  int `json:"triadic_phase"`  // 1-3
	PentadicStage int `json:"pentadic_stage"` // 1-5
	FourStepPhase int `json:"four_step_phase"` // 1-4

	// Stage in σ scheduler
	Stage     string `json:"stage"`      // "perception", "analysis", "planning", "execution", "integration"
	StageStep int    `json:"stage_step"` // 1-6 within stage

	// Processing flags
	C8Processed  bool `json:"c8_processed"`
	K9Processed  bool `json:"k9_processed"`
	PhiProcessed bool `json:"phi_processed"`
	Completed    bool `json:"completed"`
}

// TokenTransformation records a transformation applied to the token
type TokenTransformation struct {
	Timestamp   time.Time `json:"timestamp"`
	Component   string    `json:"component"`   // "c8", "k9", "phi", "sigma"
	Operation   string    `json:"operation"`
	Description string    `json:"description"`
	InputHash   string    `json:"input_hash,omitempty"`
	OutputHash  string    `json:"output_hash,omitempty"`
}

// C8TokenResult holds the result of processing through one C₈ state
type C8TokenResult struct {
	StateID     int                    `json:"state_id"`     // 0-7
	BinaryCode  string                 `json:"binary_code"`  // e.g., "010"
	Perspective string                 `json:"perspective"`  // Cognitive perspective
	Output      string                 `json:"output"`       // LLM-generated analysis
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// K9TokenResult holds the result of processing through one K₉ phase
type K9TokenResult struct {
	PhaseID     int                    `json:"phase_id"`     // 0-8
	GridPos     [2]int                 `json:"grid_pos"`     // [row, col] in 3x3 grid
	Temporal    string                 `json:"temporal"`     // "past", "present", "future"
	Scope       string                 `json:"scope"`        // "universal", "particular", "relational"
	Output      string                 `json:"output"`       // LLM-generated analysis
	Weight      float64                `json:"weight"`       // Convolution weight
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PhiTokenResult holds the result of the delay fold integration
type PhiTokenResult struct {
	Step        int                    `json:"step"`         // 1-4
	StateNum    int                    `json:"state_num"`    // 1, 4, 6, 1
	Dyad        string                 `json:"dyad"`         // "A" or "B"
	Triad       int                    `json:"triad"`        // 1, 2, or 3
	DyadHeld    bool                   `json:"dyad_held"`
	TriadHeld   bool                   `json:"triad_held"`
	Integration string                 `json:"integration"`  // Integrated output
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// GRAPH MESSAGE PAYLOAD
// =============================================================================

// GraphMessage represents a relational structure connecting cognitive elements.
// Graph messages carry information about relationships, dependencies, and
// causal connections between tokens and concepts.
type GraphMessage struct {
	// Identity
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Message type
	MessageType GraphMessageType `json:"message_type"`

	// Graph structure
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`

	// Root node (entry point for processing)
	RootNodeID string `json:"root_node_id"`

	// Semantic properties
	Coherence  float64 `json:"coherence"`   // 0.0-1.0, internal consistency
	Complexity float64 `json:"complexity"`  // 0.0-1.0, structural complexity
	Depth      int     `json:"depth"`       // Maximum path length from root

	// Pipeline state
	PipelineState GraphPipelineState `json:"pipeline_state"`

	// Processing results
	C8GraphResults [8]*C8GraphResult `json:"c8_graph_results,omitempty"`
	K9GraphResults [9]*K9GraphResult `json:"k9_graph_results,omitempty"`
	PhiGraphResult *PhiGraphResult   `json:"phi_graph_result,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphMessageType categorizes the relational structure
type GraphMessageType string

const (
	GraphTypeCausal      GraphMessageType = "causal"       // Cause-effect relationships
	GraphTypeSemantic    GraphMessageType = "semantic"     // Meaning relationships
	GraphTypeTemporal    GraphMessageType = "temporal"     // Time-ordered events
	GraphTypeSpatial     GraphMessageType = "spatial"      // Spatial relationships
	GraphTypeHierarchical GraphMessageType = "hierarchical" // Part-whole relationships
	GraphTypeAssociative GraphMessageType = "associative"  // Loose associations
	GraphTypeInferential GraphMessageType = "inferential"  // Logical inferences
	GraphTypeGoalTree    GraphMessageType = "goal_tree"    // Goal decomposition
)

// GraphNode represents a node in the cognitive graph
type GraphNode struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	NodeType   GraphNodeType          `json:"node_type"`
	Content    string                 `json:"content"`
	TokenRef   string                 `json:"token_ref,omitempty"` // Reference to CognitiveToken
	Activation float64                `json:"activation"`          // 0.0-1.0
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// GraphNodeType categorizes nodes
type GraphNodeType string

const (
	NodeTypeConcept   GraphNodeType = "concept"
	NodeTypeEntity    GraphNodeType = "entity"
	NodeTypeEvent     GraphNodeType = "event"
	NodeTypeState     GraphNodeType = "state"
	NodeTypeAction    GraphNodeType = "action"
	NodeTypeGoal      GraphNodeType = "goal"
	NodeTypeCondition GraphNodeType = "condition"
	NodeTypeResult    GraphNodeType = "result"
)

// GraphEdge represents a directed edge in the cognitive graph
type GraphEdge struct {
	ID         string                 `json:"id"`
	SourceID   string                 `json:"source_id"`
	TargetID   string                 `json:"target_id"`
	EdgeType   GraphEdgeType          `json:"edge_type"`
	Label      string                 `json:"label,omitempty"`
	Weight     float64                `json:"weight"` // 0.0-1.0, relationship strength
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// GraphEdgeType categorizes edges
type GraphEdgeType string

const (
	EdgeTypeCauses     GraphEdgeType = "causes"
	EdgeTypeEnables    GraphEdgeType = "enables"
	EdgeTypeInhibits   GraphEdgeType = "inhibits"
	EdgeTypePrecedes   GraphEdgeType = "precedes"
	EdgeTypeContains   GraphEdgeType = "contains"
	EdgeTypeIsA        GraphEdgeType = "is_a"
	EdgeTypeHasProperty GraphEdgeType = "has_property"
	EdgeTypeRelatesTo  GraphEdgeType = "relates_to"
	EdgeTypeSupports   GraphEdgeType = "supports"
	EdgeTypeConflicts  GraphEdgeType = "conflicts"
)

// GraphPipelineState tracks where a graph message is in the sys6 pipeline
type GraphPipelineState struct {
	GlobalStep    int    `json:"global_step"`
	Stage         string `json:"stage"`
	C8Processed   bool   `json:"c8_processed"`
	K9Processed   bool   `json:"k9_processed"`
	PhiProcessed  bool   `json:"phi_processed"`
	Completed     bool   `json:"completed"`
}

// C8GraphResult holds graph processing results from one C₈ state
type C8GraphResult struct {
	StateID       int                    `json:"state_id"`
	Perspective   string                 `json:"perspective"`
	ModifiedNodes []string               `json:"modified_nodes"`
	ModifiedEdges []string               `json:"modified_edges"`
	NewNodes      []GraphNode            `json:"new_nodes,omitempty"`
	NewEdges      []GraphEdge            `json:"new_edges,omitempty"`
	Analysis      string                 `json:"analysis"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// K9GraphResult holds graph processing results from one K₉ phase
type K9GraphResult struct {
	PhaseID       int                    `json:"phase_id"`
	Temporal      string                 `json:"temporal"`
	Scope         string                 `json:"scope"`
	Subgraph      *GraphMessage          `json:"subgraph,omitempty"` // Extracted subgraph
	Transformations []string             `json:"transformations"`
	Analysis      string                 `json:"analysis"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PhiGraphResult holds the delay fold integration for graphs
type PhiGraphResult struct {
	Step          int                    `json:"step"`
	DyadGraph     *GraphMessage          `json:"dyad_graph,omitempty"`
	TriadGraph    *GraphMessage          `json:"triad_graph,omitempty"`
	IntegratedGraph *GraphMessage        `json:"integrated_graph,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// PAYLOAD ENVELOPE (UNIFIED CONTAINER)
// =============================================================================

// PayloadEnvelope is a unified container that can hold either tokens or graphs
// and tracks their journey through the sys6 pipeline.
type PayloadEnvelope struct {
	mu sync.RWMutex

	// Identity
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Payload type
	PayloadType PayloadType `json:"payload_type"`

	// Content (one of these will be set)
	Token *CognitiveToken `json:"token,omitempty"`
	Graph *GraphMessage   `json:"graph,omitempty"`

	// Batch processing (multiple tokens/graphs)
	TokenBatch []*CognitiveToken `json:"token_batch,omitempty"`
	GraphBatch []*GraphMessage   `json:"graph_batch,omitempty"`

	// Pipeline routing
	Priority    int      `json:"priority"`     // Higher = more urgent
	Destination string   `json:"destination"`  // Target component
	Route       []string `json:"route"`        // Path through pipeline

	// Processing state
	EntryStep   int       `json:"entry_step"`   // Step when entered pipeline
	CurrentStep int       `json:"current_step"` // Current step
	ExitStep    int       `json:"exit_step"`    // Step when completed
	EntryTime   time.Time `json:"entry_time"`
	ExitTime    time.Time `json:"exit_time,omitempty"`

	// Error handling
	Errors []PayloadError `json:"errors,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PayloadType identifies the content type
type PayloadType string

const (
	PayloadTypeToken      PayloadType = "token"
	PayloadTypeGraph      PayloadType = "graph"
	PayloadTypeTokenBatch PayloadType = "token_batch"
	PayloadTypeGraphBatch PayloadType = "graph_batch"
)

// PayloadError records an error during processing
type PayloadError struct {
	Timestamp time.Time `json:"timestamp"`
	Component string    `json:"component"`
	Step      int       `json:"step"`
	Error     string    `json:"error"`
	Recovered bool      `json:"recovered"`
}

// =============================================================================
// PAYLOAD FACTORY
// =============================================================================

// PayloadFactory creates and manages payloads
type PayloadFactory struct {
	mu        sync.Mutex
	idCounter uint64
}

// NewPayloadFactory creates a new payload factory
func NewPayloadFactory() *PayloadFactory {
	return &PayloadFactory{}
}

// generateID creates a unique payload ID
func (pf *PayloadFactory) generateID(prefix string) string {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.idCounter++
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), pf.idCounter)
}

// CreateToken creates a new cognitive token
func (pf *PayloadFactory) CreateToken(content string, contentType TokenType, source string) *CognitiveToken {
	return &CognitiveToken{
		ID:        pf.generateID("tok"),
		CreatedAt: time.Now(),
		Content: TokenContent{
			Text:   content,
			Source: source,
		},
		ContentType: contentType,
		Salience:    0.5,
		Relevance:   0.5,
		Valence:     0.0,
		Confidence:  0.5,
		Tags:        []string{},
		PipelineState: TokenPipelineState{
			GlobalStep: 1,
			Stage:      "perception",
		},
		Transformations: []TokenTransformation{},
		Metadata:        make(map[string]interface{}),
	}
}

// CreateGraph creates a new graph message
func (pf *PayloadFactory) CreateGraph(messageType GraphMessageType) *GraphMessage {
	return &GraphMessage{
		ID:          pf.generateID("grp"),
		CreatedAt:   time.Now(),
		MessageType: messageType,
		Nodes:       []GraphNode{},
		Edges:       []GraphEdge{},
		Coherence:   1.0,
		Complexity:  0.0,
		Depth:       0,
		PipelineState: GraphPipelineState{
			GlobalStep: 1,
			Stage:      "perception",
		},
		Metadata: make(map[string]interface{}),
	}
}

// CreateEnvelope wraps a payload in an envelope
func (pf *PayloadFactory) CreateEnvelope(payload interface{}, priority int) *PayloadEnvelope {
	env := &PayloadEnvelope{
		ID:          pf.generateID("env"),
		CreatedAt:   time.Now(),
		Priority:    priority,
		EntryStep:   1,
		CurrentStep: 1,
		EntryTime:   time.Now(),
		Route:       []string{},
		Metadata:    make(map[string]interface{}),
	}

	switch p := payload.(type) {
	case *CognitiveToken:
		env.PayloadType = PayloadTypeToken
		env.Token = p
	case *GraphMessage:
		env.PayloadType = PayloadTypeGraph
		env.Graph = p
	case []*CognitiveToken:
		env.PayloadType = PayloadTypeTokenBatch
		env.TokenBatch = p
	case []*GraphMessage:
		env.PayloadType = PayloadTypeGraphBatch
		env.GraphBatch = p
	}

	return env
}

// =============================================================================
// PAYLOAD SERIALIZATION
// =============================================================================

// ToJSON serializes a payload envelope to JSON
func (pe *PayloadEnvelope) ToJSON() ([]byte, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return json.MarshalIndent(pe, "", "  ")
}

// FromJSON deserializes a payload envelope from JSON
func PayloadEnvelopeFromJSON(data []byte) (*PayloadEnvelope, error) {
	var pe PayloadEnvelope
	if err := json.Unmarshal(data, &pe); err != nil {
		return nil, err
	}
	return &pe, nil
}

// =============================================================================
// C₈ PERSPECTIVE DEFINITIONS
// =============================================================================

// C8Perspectives defines the cognitive perspective for each of the 8 states
var C8Perspectives = [8]C8Perspective{
	{ID: 0, Binary: "000", Name: "Perception-Action-Learning", 
		Description: "Direct sensory processing with immediate action and learning"},
	{ID: 1, Binary: "001", Name: "Perception-Action-Integration",
		Description: "Direct sensory processing with immediate action and knowledge integration"},
	{ID: 2, Binary: "010", Name: "Perception-Reflection-Learning",
		Description: "Direct sensory processing with reflective analysis and learning"},
	{ID: 3, Binary: "011", Name: "Perception-Reflection-Integration",
		Description: "Direct sensory processing with reflective analysis and integration"},
	{ID: 4, Binary: "100", Name: "Expression-Action-Learning",
		Description: "Generative output with immediate action and learning"},
	{ID: 5, Binary: "101", Name: "Expression-Action-Integration",
		Description: "Generative output with immediate action and integration"},
	{ID: 6, Binary: "110", Name: "Expression-Reflection-Learning",
		Description: "Generative output with reflective analysis and learning"},
	{ID: 7, Binary: "111", Name: "Expression-Reflection-Integration",
		Description: "Generative output with reflective analysis and integration"},
}

// C8Perspective describes a cognitive perspective
type C8Perspective struct {
	ID          int    `json:"id"`
	Binary      string `json:"binary"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// =============================================================================
// K₉ PHASE DEFINITIONS
// =============================================================================

// K9Phases defines the 9 phases in the 3x3 temporal-scope grid
var K9Phases = [9]K9Phase{
	// Row 0: Past
	{ID: 0, Row: 0, Col: 0, Temporal: "past", Scope: "universal",
		Name: "Past-Universal", Description: "Historical patterns and universal principles"},
	{ID: 1, Row: 0, Col: 1, Temporal: "past", Scope: "particular",
		Name: "Past-Particular", Description: "Specific memories and experiences"},
	{ID: 2, Row: 0, Col: 2, Temporal: "past", Scope: "relational",
		Name: "Past-Relational", Description: "Historical connections and relationships"},
	// Row 1: Present
	{ID: 3, Row: 1, Col: 0, Temporal: "present", Scope: "universal",
		Name: "Present-Universal", Description: "Current universal truths and laws"},
	{ID: 4, Row: 1, Col: 1, Temporal: "present", Scope: "particular",
		Name: "Present-Particular", Description: "Current specific situation and context"},
	{ID: 5, Row: 1, Col: 2, Temporal: "present", Scope: "relational",
		Name: "Present-Relational", Description: "Current relationships and dynamics"},
	// Row 2: Future
	{ID: 6, Row: 2, Col: 0, Temporal: "future", Scope: "universal",
		Name: "Future-Universal", Description: "Potential universal outcomes and trends"},
	{ID: 7, Row: 2, Col: 1, Temporal: "future", Scope: "particular",
		Name: "Future-Particular", Description: "Specific anticipated results"},
	{ID: 8, Row: 2, Col: 2, Temporal: "future", Scope: "relational",
		Name: "Future-Relational", Description: "Anticipated relationship changes"},
}

// K9Phase describes a convolution phase
type K9Phase struct {
	ID          int    `json:"id"`
	Row         int    `json:"row"`
	Col         int    `json:"col"`
	Temporal    string `json:"temporal"`
	Scope       string `json:"scope"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// =============================================================================
// STAGE DEFINITIONS
// =============================================================================

// SigmaStages defines the 5 stages in the σ scheduler
var SigmaStages = [5]SigmaStage{
	{ID: 1, Name: "Perception", Steps: [6]int{1, 2, 3, 4, 5, 6},
		Description: "Gather and process sensory input from environment",
		TokenFocus: []TokenType{TokenTypePercept, TokenTypeQuery},
		GraphFocus: []GraphMessageType{GraphTypeSemantic, GraphTypeSpatial}},
	{ID: 2, Name: "Analysis", Steps: [6]int{7, 8, 9, 10, 11, 12},
		Description: "Analyze perceptions and identify patterns",
		TokenFocus: []TokenType{TokenTypeThought, TokenTypeMemory},
		GraphFocus: []GraphMessageType{GraphTypeCausal, GraphTypeInferential}},
	{ID: 3, Name: "Planning", Steps: [6]int{13, 14, 15, 16, 17, 18},
		Description: "Generate and evaluate action plans",
		TokenFocus: []TokenType{TokenTypeGoal, TokenTypeAffordance},
		GraphFocus: []GraphMessageType{GraphTypeGoalTree, GraphTypeTemporal}},
	{ID: 4, Name: "Execution", Steps: [6]int{19, 20, 21, 22, 23, 24},
		Description: "Execute selected actions",
		TokenFocus: []TokenType{TokenTypeAction, TokenTypeResponse},
		GraphFocus: []GraphMessageType{GraphTypeCausal, GraphTypeTemporal}},
	{ID: 5, Name: "Integration", Steps: [6]int{25, 26, 27, 28, 29, 30},
		Description: "Integrate results and update knowledge",
		TokenFocus: []TokenType{TokenTypeInsight, TokenTypeMemory},
		GraphFocus: []GraphMessageType{GraphTypeHierarchical, GraphTypeAssociative}},
}

// SigmaStage describes a processing stage
type SigmaStage struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Steps       [6]int             `json:"steps"`
	Description string             `json:"description"`
	TokenFocus  []TokenType        `json:"token_focus"`
	GraphFocus  []GraphMessageType `json:"graph_focus"`
}
