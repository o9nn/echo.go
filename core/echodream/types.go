package echodream

import "time"

// WakeState represents the current wake/rest state
type WakeState int

const (
	StateWaking WakeState = iota
	StateAwake
	StateTiring
	StateResting
	StateDreaming
)

func (ws WakeState) String() string {
	return [...]string{"waking", "awake", "tiring", "resting", "dreaming"}[ws]
}

// EpisodicMemory represents a memory of an experience
type EpisodicMemory struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"` // "thought", "insight", "action", "observation"
	Content   string                 `json:"content"`
	Context   map[string]interface{} `json:"context"`
	Salience  float64                `json:"salience"`  // How important/notable
	Valence   float64                `json:"valence"`   // Emotional tone (-1 to 1)
	Tags      []string               `json:"tags"`
}

// Pattern represents a detected pattern across memories
type Pattern struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "recurring_theme", "connection", "contradiction"
	Description string                 `json:"description"`
	Memories    []string               `json:"memories"` // IDs of related memories
	Strength    float64                `json:"strength"` // How strong the pattern is
	Abstraction string                 `json:"abstraction"` // Higher-level insight
	Context     map[string]interface{} `json:"context"`
	DetectedAt  time.Time              `json:"detected_at"`
}

// WisdomNugget represents synthesized wisdom
type WisdomNugget struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Patterns    []string               `json:"patterns"` // IDs of source patterns
	Abstraction int                    `json:"abstraction"` // Level of abstraction (1-5)
	Confidence  float64                `json:"confidence"` // How confident in this wisdom
	Applicability string               `json:"applicability"` // When/where this applies
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time              `json:"created_at"`
	Refined     int                    `json:"refined"` // How many times refined
}

// DreamCycleMetrics tracks metrics for dream cycles
type DreamCycleMetrics struct {
	TotalCycles          uint64
	MemoriesConsolidated uint64
	PatternsDetected     uint64
	WisdomExtracted      uint64
	AverageCycleDuration time.Duration
	LastCycleTime        time.Time
}

// ConsolidationResult represents the result of memory consolidation
type ConsolidationResult struct {
	MemoriesProcessed int
	PatternsFound     []Pattern
	WisdomGenerated   []WisdomNugget
	Duration          time.Duration
	Success           bool
	Error             error
}
