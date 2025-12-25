package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// =============================================================================
// OPPONENT PROCESS SYSTEM
// =============================================================================
//
// Implements the Ordo (Order) and Chao (Chaos) archetypal dynamics for
// wisdom cultivation through dynamic balance. Based on opponent process theory
// from neuroscience and the concept of sophrosyne (wisdom through balance).
//
// Key opponent pairs:
// - Exploration ↔ Exploitation
// - Breadth ↔ Depth
// - Stability ↔ Flexibility
// - Speed ↔ Accuracy
// - Approach ↔ Avoidance
//
// =============================================================================

// Opponent pair names
const (
	ExplorationExploitation = "exploration_exploitation"
	BreadthDepth            = "breadth_depth"
	StabilityFlexibility    = "stability_flexibility"
	SpeedAccuracy           = "speed_accuracy"
	ApproachAvoidance       = "approach_avoidance"
)

// OpponentProcess represents a single opponent process pair
type OpponentProcess struct {
	Name           string
	CurrentBalance float64 // -1.0 to 1.0, negative = first pole, positive = second pole
	History        []BalancePoint
	Stability      float64 // How stable the balance has been
	LastUpdate     time.Time
}

// BalancePoint records a balance state at a point in time
type BalancePoint struct {
	Balance   float64
	Timestamp time.Time
	Context   string
}

// OpponentProcessSystem manages all opponent process pairs
type OpponentProcessSystem struct {
	mu        sync.RWMutex
	processes map[string]*OpponentProcess
}

// NewOpponentProcessSystem creates a new opponent process system
func NewOpponentProcessSystem() *OpponentProcessSystem {
	ops := &OpponentProcessSystem{
		processes: make(map[string]*OpponentProcess),
	}

	// Initialize all opponent pairs at neutral balance
	pairs := []string{
		ExplorationExploitation,
		BreadthDepth,
		StabilityFlexibility,
		SpeedAccuracy,
		ApproachAvoidance,
	}

	for _, pair := range pairs {
		ops.processes[pair] = &OpponentProcess{
			Name:           pair,
			CurrentBalance: 0.0,
			History:        make([]BalancePoint, 0),
			Stability:      0.5,
			LastUpdate:     time.Now(),
		}
	}

	return ops
}

// UpdateBalance updates the balance for a specific opponent pair
func (ops *OpponentProcessSystem) UpdateBalance(pairName string, delta float64, context string) {
	ops.mu.Lock()
	defer ops.mu.Unlock()

	process, exists := ops.processes[pairName]
	if !exists {
		return
	}

	// Apply delta with bounds
	newBalance := process.CurrentBalance + delta
	newBalance = math.Max(-1.0, math.Min(1.0, newBalance))

	// Record history
	process.History = append(process.History, BalancePoint{
		Balance:   newBalance,
		Timestamp: time.Now(),
		Context:   context,
	})

	// Keep only last 100 history points
	if len(process.History) > 100 {
		process.History = process.History[len(process.History)-100:]
	}

	// Update stability based on variance
	process.Stability = ops.calculateStability(process.History)
	process.CurrentBalance = newBalance
	process.LastUpdate = time.Now()
}

// GetCurrentBalance returns the current balance for a pair
func (ops *OpponentProcessSystem) GetCurrentBalance(pairName string) float64 {
	ops.mu.RLock()
	defer ops.mu.RUnlock()

	if process, exists := ops.processes[pairName]; exists {
		return process.CurrentBalance
	}
	return 0.0
}

// GetBalanceStats returns statistics for a pair
func (ops *OpponentProcessSystem) GetBalanceStats(pairName string) map[string]float64 {
	ops.mu.RLock()
	defer ops.mu.RUnlock()

	process, exists := ops.processes[pairName]
	if !exists {
		return nil
	}

	return map[string]float64{
		"current_balance": process.CurrentBalance,
		"stability":       process.Stability,
	}
}

// calculateStability calculates stability from history variance
func (ops *OpponentProcessSystem) calculateStability(history []BalancePoint) float64 {
	if len(history) < 2 {
		return 0.5
	}

	// Calculate variance
	var sum, sumSq float64
	for _, point := range history {
		sum += point.Balance
		sumSq += point.Balance * point.Balance
	}
	n := float64(len(history))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)

	// Convert variance to stability (lower variance = higher stability)
	// Variance ranges from 0 to 1 (since balance is -1 to 1)
	stability := 1.0 - math.Sqrt(variance)
	return math.Max(0.0, math.Min(1.0, stability))
}

// =============================================================================
// EXTENDED IDENTITY WITH OPPONENT PROCESSES
// =============================================================================

// OpponentPattern represents a cognitive pattern for opponent process tracking
type OpponentPattern struct {
	ID       string
	Strength float64
	Tags     []string
	Created  time.Time
}

// EmotionalState represents the current emotional state
type EmotionalState struct {
	Arousal float64 // 0.0 to 1.0
	Valence float64 // -1.0 to 1.0
}

// RelevanceDecision represents a decision from relevance realization
type RelevanceDecision struct {
	ExplorationWeight float64 // 0.0 to 1.0, higher = more exploration
	ScopePreference   string  // "breadth" or "depth"
	AdaptationRate    float64 // 0.0 to 1.0, higher = more flexible
	Confidence        float64 // 0.0 to 1.0, confidence threshold
	Context           string
	Timestamp         time.Time
}

// ExtendedIdentity extends Identity with opponent process capabilities
type ExtendedIdentity struct {
	*Identity
	mu               sync.RWMutex
	Patterns         map[string]*OpponentPattern
	Coherence        float64
	Iterations       uint64
	EmotionalState   *EmotionalState
	OpponentProcesses *OpponentProcessSystem
	wisdomHistory    []float64
}

// NewExtendedIdentity creates a new extended identity
func NewExtendedIdentity(name string) *ExtendedIdentity {
	return &ExtendedIdentity{
		Identity:          NewIdentity(name, []string{}),
		Patterns:          make(map[string]*OpponentPattern),
		Coherence:         0.5,
		Iterations:        0,
		EmotionalState:    &EmotionalState{Arousal: 0.5, Valence: 0.0},
		OpponentProcesses: NewOpponentProcessSystem(),
		wisdomHistory:     make([]float64, 0),
	}
}

// OptimizeRelevanceRealization performs relevance realization optimization
// This is the core decision-making function that balances opponent processes
func (ei *ExtendedIdentity) OptimizeRelevanceRealization(context string) *RelevanceDecision {
	ei.mu.Lock()
	defer ei.mu.Unlock()

	ei.Iterations++

	// Calculate base exploration weight based on state
	explorationWeight := ei.calculateExplorationWeight()

	// Calculate scope preference
	scopePreference := ei.calculateScopePreference()

	// Calculate adaptation rate
	adaptationRate := ei.calculateAdaptationRate()

	// Calculate confidence threshold
	confidence := ei.calculateConfidenceThreshold()

	// Update opponent processes based on decision
	ei.updateOpponentProcesses(explorationWeight, scopePreference, adaptationRate, confidence, context)

	return &RelevanceDecision{
		ExplorationWeight: explorationWeight,
		ScopePreference:   scopePreference,
		AdaptationRate:    adaptationRate,
		Confidence:        confidence,
		Context:           context,
		Timestamp:         time.Now(),
	}
}

// calculateExplorationWeight determines exploration vs exploitation balance
func (ei *ExtendedIdentity) calculateExplorationWeight() float64 {
	// Factors favoring exploration (Chao):
	// - Few patterns (need to discover)
	// - High coherence (risk of stagnation)
	// - Early iterations (need to explore)

	// Factors favoring exploitation (Ordo):
	// - Many patterns (need to consolidate)
	// - Low coherence (need to integrate)
	// - Many iterations (time to exploit)

	patternCount := float64(len(ei.Patterns))
	
	// Normalize pattern count (0-100 patterns maps to 0-1)
	patternFactor := math.Min(1.0, patternCount/100.0)
	
	// High coherence suggests stagnation risk → explore more
	coherenceFactor := ei.Coherence
	
	// Early iterations → explore more
	iterationFactor := 1.0 - math.Min(1.0, float64(ei.Iterations)/1000.0)

	// Combine factors
	// Few patterns + high coherence + early iterations = high exploration
	// Many patterns + low coherence + many iterations = low exploration (exploitation)
	exploration := (1.0-patternFactor)*0.3 + coherenceFactor*0.4 + iterationFactor*0.3

	// Apply emotional modulation
	if ei.EmotionalState != nil {
		// High arousal increases exploration (novelty seeking)
		exploration += (ei.EmotionalState.Arousal - 0.5) * 0.2
	}

	return math.Max(0.0, math.Min(1.0, exploration))
}

// calculateScopePreference determines breadth vs depth preference
func (ei *ExtendedIdentity) calculateScopePreference() string {
	// Breadth (Chao): Few patterns, need to explore widely
	// Depth (Ordo): Many patterns, need to consolidate deeply

	patternCount := float64(len(ei.Patterns))
	
	// Threshold for switching from breadth to depth
	if patternCount > 30 && ei.Coherence < 0.7 {
		return "depth"
	}
	return "breadth"
}

// calculateAdaptationRate determines stability vs flexibility
func (ei *ExtendedIdentity) calculateAdaptationRate() float64 {
	// High adaptation (Chao): Early stage, need flexibility
	// Low adaptation (Ordo): Mature stage, need stability

	// More iterations → lower adaptation (more stable)
	iterationFactor := 1.0 - math.Min(1.0, float64(ei.Iterations)/2000.0)
	
	// Higher coherence → lower adaptation (more stable)
	coherenceFactor := 1.0 - ei.Coherence

	adaptation := iterationFactor*0.5 + coherenceFactor*0.5

	// Apply emotional modulation
	if ei.EmotionalState != nil {
		// High arousal increases adaptation (more reactive)
		adaptation += (ei.EmotionalState.Arousal - 0.5) * 0.2
	}

	return math.Max(0.0, math.Min(1.0, adaptation))
}

// calculateConfidenceThreshold determines speed vs accuracy tradeoff
func (ei *ExtendedIdentity) calculateConfidenceThreshold() float64 {
	// High confidence threshold (Ordo): Favor accuracy
	// Low confidence threshold (Chao): Favor speed

	// More patterns → higher confidence threshold (more accurate)
	patternFactor := math.Min(1.0, float64(len(ei.Patterns))/50.0)
	
	// Higher coherence → higher confidence threshold
	coherenceFactor := ei.Coherence

	confidence := patternFactor*0.4 + coherenceFactor*0.4 + 0.2

	// Apply emotional modulation
	if ei.EmotionalState != nil {
		// High arousal decreases confidence threshold (favor speed)
		confidence -= (ei.EmotionalState.Arousal - 0.5) * 0.3
	}

	return math.Max(0.1, math.Min(0.9, confidence))
}

// updateOpponentProcesses updates all opponent process balances
func (ei *ExtendedIdentity) updateOpponentProcesses(exploration float64, scope string, adaptation float64, confidence float64, context string) {
	// Exploration-Exploitation: exploration weight maps to balance
	// Negative = exploitation, Positive = exploration
	explorationBalance := (exploration - 0.5) * 2.0
	ei.OpponentProcesses.UpdateBalance(ExplorationExploitation, explorationBalance*0.1, context)

	// Breadth-Depth: scope preference maps to balance
	// Negative = depth, Positive = breadth
	var scopeBalance float64
	if scope == "breadth" {
		scopeBalance = 0.5
	} else {
		scopeBalance = -0.5
	}
	ei.OpponentProcesses.UpdateBalance(BreadthDepth, scopeBalance*0.1, context)

	// Stability-Flexibility: adaptation rate maps to balance
	// Negative = stability, Positive = flexibility
	stabilityBalance := (adaptation - 0.5) * 2.0
	ei.OpponentProcesses.UpdateBalance(StabilityFlexibility, stabilityBalance*0.1, context)

	// Speed-Accuracy: confidence threshold maps to balance
	// Negative = speed, Positive = accuracy
	accuracyBalance := (confidence - 0.5) * 2.0
	ei.OpponentProcesses.UpdateBalance(SpeedAccuracy, accuracyBalance*0.1, context)

	// Approach-Avoidance: valence maps to balance
	if ei.EmotionalState != nil {
		ei.OpponentProcesses.UpdateBalance(ApproachAvoidance, ei.EmotionalState.Valence*0.1, context)
	}
}

// GetWisdomScore calculates the current wisdom score
// Wisdom = balance stability across all opponent processes
func (ei *ExtendedIdentity) GetWisdomScore() float64 {
	ei.mu.RLock()
	defer ei.mu.RUnlock()

	pairs := []string{
		ExplorationExploitation,
		BreadthDepth,
		StabilityFlexibility,
		SpeedAccuracy,
		ApproachAvoidance,
	}

	var totalStability float64
	for _, pair := range pairs {
		stats := ei.OpponentProcesses.GetBalanceStats(pair)
		if stats != nil {
			totalStability += stats["stability"]
		}
	}

	// Average stability across all pairs
	avgStability := totalStability / float64(len(pairs))

	// Factor in coherence and iteration maturity
	maturityFactor := math.Min(1.0, float64(ei.Iterations)/1000.0)
	coherenceFactor := ei.Coherence

	// Wisdom = stability * coherence * maturity
	wisdom := avgStability * 0.4 + coherenceFactor * 0.3 + maturityFactor * 0.3

	// Record wisdom history
	ei.wisdomHistory = append(ei.wisdomHistory, wisdom)
	if len(ei.wisdomHistory) > 100 {
		ei.wisdomHistory = ei.wisdomHistory[len(ei.wisdomHistory)-100:]
	}

	return wisdom
}

// GetWisdomTrend returns the trend of wisdom over recent history
func (ei *ExtendedIdentity) GetWisdomTrend() float64 {
	ei.mu.RLock()
	defer ei.mu.RUnlock()

	if len(ei.wisdomHistory) < 2 {
		return 0.0
	}

	// Calculate trend as difference between recent and older averages
	mid := len(ei.wisdomHistory) / 2
	
	var oldSum, newSum float64
	for i := 0; i < mid; i++ {
		oldSum += ei.wisdomHistory[i]
	}
	for i := mid; i < len(ei.wisdomHistory); i++ {
		newSum += ei.wisdomHistory[i]
	}

	oldAvg := oldSum / float64(mid)
	newAvg := newSum / float64(len(ei.wisdomHistory)-mid)

	return newAvg - oldAvg
}
