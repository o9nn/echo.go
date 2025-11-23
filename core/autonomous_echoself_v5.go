package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/EchoCog/echollama/core/deeptreeecho"
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
)

// AutonomousEchoselfV5 represents the V5 autonomous agent with full cognitive loop integration
type AutonomousEchoselfV5 struct {
	mu                    sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc

	// Core identity and model management
	identity              *deeptreeecho.Identity
	modelManager          *deeptreeecho.ModelManager

	// Echobeats 3-phase cognitive loop
	phaseManager          *echobeats.PhaseManager

	// Echodream wake/rest and consolidation
	wakeController        *echodream.AutonomousController
	consolidationEngine   *echodream.ConsolidationEngine

	// State management
	currentState          echodream.WakeState
	stateTransitions      []StateTransition

	// Metrics and tracking
	thoughtsGenerated     uint64
	insightsGenerated     uint64
	wisdomScore           float64
	startTime             time.Time
	running               bool

	// Thought generation
	thoughtGenerator      *ThoughtGenerator
	recentThoughtsBuffer  []string

	// Discussion management
	discussionManager     *DiscussionManager

	// Configuration
	config                *EchoselfConfigV5

	// Persistence
	persistenceManager    *PersistenceManager
	sessionID             string
}

// EchoselfConfigV5 holds configuration for the V5 agent
type EchoselfConfigV5 struct {
	// Cognitive loop timing
	StepDuration          time.Duration

	// Wake/rest parameters
	MaxAwakeDuration      time.Duration
	MinRestDuration       time.Duration
	FatigueThreshold      float64
	RestThreshold         float64

	// Memory and consolidation
	MaxEpisodicBuffer     int

	// Feature flags
	EnableEchobeats       bool
	EnableWakeRest        bool
	EnableConsolidation   bool
	EnablePersistence     bool

	// Persistence settings
	PersistencePath       string
	MaxSnapshots          int

	// LLM parameters
	LLMTemperature        float64
	LLMMaxTokens          int
}

// StateTransition records state changes
type StateTransition struct {
	From      echodream.WakeState
	To        echodream.WakeState
	Timestamp time.Time
	Reason    string
}

// DefaultEchoselfConfigV5 returns default configuration
func DefaultEchoselfConfigV5() *EchoselfConfigV5 {
	return &EchoselfConfigV5{
		StepDuration:        5 * time.Second,
		MaxAwakeDuration:    30 * time.Minute,
		MinRestDuration:     5 * time.Minute,
		FatigueThreshold:    0.8,
		RestThreshold:       0.3,
		MaxEpisodicBuffer:   100,
		EnableEchobeats:     true,
		EnableWakeRest:      true,
		EnableConsolidation: true,
		EnablePersistence:   false,
		PersistencePath:     "~/.echo9llama/consciousness",
		MaxSnapshots:        10,
		LLMTemperature:      0.8,
		LLMMaxTokens:        200,
	}
}

// NewAutonomousEchoselfV5 creates a new V5 autonomous agent
func NewAutonomousEchoselfV5(config *EchoselfConfigV5) (*AutonomousEchoselfV5, error) {
	if config == nil {
		config = DefaultEchoselfConfigV5()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create core identity
	identity := deeptreeecho.NewIdentity("EchoselfV5")

	// Create model manager
	modelManager := deeptreeecho.NewModelManager(identity)

	// Create echobeats phase manager
	var phaseManager *echobeats.PhaseManager
	if config.EnableEchobeats {
		phaseManager = echobeats.NewPhaseManager(config.StepDuration)
	}

	// Create echodream components
	var wakeController *echodream.AutonomousController
	var consolidationEngine *echodream.ConsolidationEngine
	
	if config.EnableWakeRest {
		wakeController = echodream.NewAutonomousController(
			config.MaxAwakeDuration,
			config.MinRestDuration,
			config.FatigueThreshold,
			config.RestThreshold,
		)
	}

	if config.EnableConsolidation {
		consolidationEngine = echodream.NewConsolidationEngine(config.MaxEpisodicBuffer)
	}

	// Generate session ID
	sessionID := fmt.Sprintf("%d", time.Now().Unix())

	// Create persistence manager
	var persistenceManager *PersistenceManager
	if config.EnablePersistence {
		// Expand tilde in path
		persistencePath := config.PersistencePath
		if len(persistencePath) > 0 && persistencePath[0] == '~' {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				persistencePath = filepath.Join(homeDir, persistencePath[1:])
			}
		}
		persistenceManager = NewPersistenceManager(persistencePath, config.MaxSnapshots)
	}

	ae := &AutonomousEchoselfV5{
		ctx:                 ctx,
		cancel:              cancel,
		identity:            identity,
		modelManager:        modelManager,
		phaseManager:        phaseManager,
		wakeController:      wakeController,
		consolidationEngine: consolidationEngine,
		currentState:        echodream.StateAwake,
		stateTransitions:    make([]StateTransition, 0),
		wisdomScore:         0.0,
		config:              config,
		persistenceManager:  persistenceManager,
		sessionID:           sessionID,
		thoughtGenerator:    NewThoughtGenerator(),
		recentThoughtsBuffer: make([]string, 0),
		discussionManager:   NewDiscussionManager(0.1, 0.5, 20),
	}

	// Register cognitive handlers
	if config.EnableEchobeats {
		ae.registerCognitiveHandlers()
	}

	log.Println("✅ AutonomousEchoselfV5 initialized")
	return ae, nil
}

// registerCognitiveHandlers registers handlers for each cognitive step
func (ae *AutonomousEchoselfV5) registerCognitiveHandlers() {
	// T4E - Perception Processing (Steps 0, 4)
	ae.phaseManager.RegisterHandler(
		echobeats.T4_SensoryInput,
		echobeats.Expressive,
		ae.handlePerceptionProcessing,
	)

	// T7R - Memory Consolidation (Steps 3, 10)
	ae.phaseManager.RegisterHandler(
		echobeats.T7_MemoryEncoding,
		echobeats.Reflective,
		ae.handleMemoryConsolidation,
	)

	// T2E - Thought Generation (Steps 2, 6)
	ae.phaseManager.RegisterHandler(
		echobeats.T2_IdeaFormation,
		echobeats.Expressive,
		ae.handleThoughtGeneration,
	)

	// T8E - Integrated Response (Steps 8, 9)
	ae.phaseManager.RegisterHandler(
		echobeats.T8_BalancedResponse,
		echobeats.Expressive,
		ae.handleIntegratedResponse,
	)

	// T1R - Need Assessment (Steps 1, 5)
	ae.phaseManager.RegisterHandler(
		echobeats.T1_Perception,
		echobeats.Reflective,
		ae.handleNeedAssessment,
	)

	// T5E - Action Execution (Steps 7, 11)
	ae.phaseManager.RegisterHandler(
		echobeats.T5_ActionSequence,
		echobeats.Expressive,
		ae.handleActionExecution,
	)

	log.Println("✅ Cognitive handlers registered")
}

// Start begins the autonomous cognitive loop
func (ae *AutonomousEchoselfV5) Start() error {
	ae.mu.Lock()
	if ae.running {
		ae.mu.Unlock()
		return fmt.Errorf("already running")
	}
	ae.running = true
	ae.startTime = time.Now()
	ae.mu.Unlock()

	log.Println("🌳 Starting AutonomousEchoselfV5...")

	// Start echobeats cognitive loop
	if ae.config.EnableEchobeats && ae.phaseManager != nil {
		if err := ae.phaseManager.Start(); err != nil {
			return fmt.Errorf("failed to start phase manager: %w", err)
		}
		log.Println("🎵 Echobeats cognitive loop started")
	}

	// Start wake/rest monitoring
	if ae.config.EnableWakeRest {
		go ae.wakeRestMonitoringLoop()
		log.Println("🌙 Wake/rest monitoring started")
	}

	log.Println("✅ AutonomousEchoselfV5 is now running")
	return nil
}

// Stop halts the autonomous agent
func (ae *AutonomousEchoselfV5) Stop() {
	ae.mu.Lock()
	if !ae.running {
		ae.mu.Unlock()
		return
	}
	ae.running = false
	ae.mu.Unlock()

	log.Println("🛑 Stopping AutonomousEchoselfV5...")

	// Stop echobeats
	if ae.phaseManager != nil {
		ae.phaseManager.Stop()
	}

	// Cancel context
	ae.cancel()

	log.Println("✅ AutonomousEchoselfV5 stopped")
}

// Cognitive Handler Functions

func (ae *AutonomousEchoselfV5) handlePerceptionProcessing(step int, mode echobeats.Mode) error {
	// Process current state and environment
	ae.mu.RLock()
	state := ae.currentState
	ae.mu.RUnlock()

	// Update fatigue based on cognitive load
	if ae.wakeController != nil && state == echodream.StateAwake {
		ae.wakeController.UpdateFatigue(0.01) // Small fatigue increment
	}

	return nil
}

func (ae *AutonomousEchoselfV5) handleMemoryConsolidation(step int, mode echobeats.Mode) error {
	// Add recent thoughts to episodic buffer
	if ae.consolidationEngine != nil {
		// This would normally pull from a thought buffer
		// For now, just track that consolidation is happening
	}
	return nil
}

func (ae *AutonomousEchoselfV5) handleThoughtGeneration(step int, mode echobeats.Mode) error {
	// Generate an autonomous thought
	ae.mu.Lock()
	ae.thoughtsGenerated++
	thoughtNum := ae.thoughtsGenerated
	ae.mu.Unlock()

	// Get wisdom and patterns for context
	var wisdom []echodream.WisdomNugget
	var patterns []echodream.Pattern
	if ae.consolidationEngine != nil {
		wisdom = ae.consolidationEngine.GetWisdom()
		patterns = ae.consolidationEngine.GetPatterns()
	}

	// Generate context-aware thought
	thought := ae.thoughtGenerator.GenerateThought(thoughtNum, wisdom, patterns)

	log.Printf("💭 [Step %d] Thought #%d: %s\n", step, thoughtNum, thought)

	// Add to recent thoughts buffer
	ae.mu.Lock()
	ae.recentThoughtsBuffer = append(ae.recentThoughtsBuffer, thought)
	if len(ae.recentThoughtsBuffer) > 10 {
		ae.recentThoughtsBuffer = ae.recentThoughtsBuffer[1:]
	}
	ae.mu.Unlock()

	// Extract topics and update interests
	if ae.discussionManager != nil {
		topics := ExtractTopicsFromText(thought)
		for _, topic := range topics {
			ae.discussionManager.UpdateInterest(topic, 0.1)
		}
	}

	// Add to episodic memory
	if ae.consolidationEngine != nil {
		memory := echodream.EpisodicMemory{
			Timestamp: time.Now(),
			Type:      "thought",
			Content:   thought,
			Salience:  0.7,
			Valence:   0.5,
			Tags:      []string{"autonomous", "context-aware"},
		}
		ae.consolidationEngine.AddMemory(memory)
	}

	return nil
}

func (ae *AutonomousEchoselfV5) handleIntegratedResponse(step int, mode echobeats.Mode) error {
	// Integrate recent thoughts and generate insights
	if ae.consolidationEngine != nil && ae.consolidationEngine.GetBufferSize() >= 3 {
		ae.mu.Lock()
		ae.insightsGenerated++
		insightNum := ae.insightsGenerated
		thoughts := make([]string, len(ae.recentThoughtsBuffer))
		copy(thoughts, ae.recentThoughtsBuffer)
		ae.mu.Unlock()

		// Generate insight from recent thoughts and patterns
		patterns := ae.consolidationEngine.GetPatterns()
		insight := GenerateInsight(thoughts, patterns)

		log.Printf("💡 [Step %d] Insight #%d: %s\n", step, insightNum, insight)

		// Add insight to memory
		memory := echodream.EpisodicMemory{
			Timestamp: time.Now(),
			Type:      "insight",
			Content:   insight,
			Salience:  0.9,
			Valence:   0.7,
			Tags:      []string{"insight", "integration"},
		}
		ae.consolidationEngine.AddMemory(memory)
	}

	return nil
}

func (ae *AutonomousEchoselfV5) handleNeedAssessment(step int, mode echobeats.Mode) error {
	// Assess current needs and adjust cognitive load
	ae.mu.RLock()
	thoughtCount := ae.thoughtsGenerated
	ae.mu.RUnlock()

	// Adjust cognitive load based on activity
	cognitiveLoad := float64(thoughtCount%10) / 10.0
	
	// Update metrics
	if ae.phaseManager != nil {
		metrics := ae.phaseManager.GetMetrics()
		metrics.CognitiveLoad = cognitiveLoad
	}

	return nil
}

func (ae *AutonomousEchoselfV5) handleActionExecution(step int, mode echobeats.Mode) error {
	// Execute any pending actions
	// For now, this is a placeholder
	return nil
}

// Wake/Rest Monitoring Loop

func (ae *AutonomousEchoselfV5) wakeRestMonitoringLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Interest decay ticker (every 5 minutes)
	decayTicker := time.NewTicker(5 * time.Minute)
	defer decayTicker.Stop()

	for {
		select {
		case <-ae.ctx.Done():
			return

		case <-ticker.C:
			ae.updateWakeRestState()

		case <-decayTicker.C:
			// Decay interests over time
			if ae.discussionManager != nil {
				ae.discussionManager.DecayInterests()
			}
		}
	}
}

func (ae *AutonomousEchoselfV5) updateWakeRestState() {
	ae.mu.RLock()
	currentState := ae.currentState
	ae.mu.RUnlock()

	var newState echodream.WakeState
	var reason string

	switch currentState {
	case echodream.StateAwake:
		if ae.wakeController != nil && ae.wakeController.ShouldRest() {
			newState = echodream.StateTiring
			reason = "Fatigue threshold reached"
		} else {
			return // No state change
		}

	case echodream.StateTiring:
		newState = echodream.StateResting
		reason = "Transitioning to rest"
		
		// Pause cognitive loop
		if ae.phaseManager != nil {
			ae.phaseManager.Pause()
		}
		
		if ae.wakeController != nil {
			ae.wakeController.StartRest()
		}

	case echodream.StateResting:
		restDuration := ae.wakeController.GetRestDuration()
		if restDuration > ae.config.MinRestDuration/2 {
			newState = echodream.StateDreaming
			reason = "Entering dream state for consolidation"
		} else {
			return
		}

	case echodream.StateDreaming:
		// Perform consolidation
		if ae.consolidationEngine != nil {
			result := ae.consolidationEngine.Consolidate()
			log.Printf("🌙 Dream consolidation: %d memories → %d patterns → %d wisdom nuggets\n",
				result.MemoriesProcessed, len(result.PatternsFound), len(result.WisdomGenerated))
			
			// Update wisdom score
			ae.mu.Lock()
			ae.wisdomScore += float64(len(result.WisdomGenerated)) * 0.1
			ae.mu.Unlock()

			// Save state after consolidation
			if ae.config.EnablePersistence {
				if err := ae.SaveState(); err != nil {
					log.Printf("⚠️  Failed to save state: %v\n", err)
				} else {
					log.Println("💾 Consciousness state saved")
				}
			}
		}

		// Check if should wake
		if ae.wakeController != nil && ae.wakeController.ShouldWake() {
			newState = echodream.StateWaking
			reason = "Rest complete, fatigue reduced"
		} else {
			return
		}

	case echodream.StateWaking:
		newState = echodream.StateAwake
		reason = "Resuming cognitive activity"
		
		// Resume cognitive loop
		if ae.phaseManager != nil {
			ae.phaseManager.Resume()
		}
		
		if ae.wakeController != nil {
			ae.wakeController.ResetFatigue()
		}

	default:
		return
	}

	// Record state transition
	ae.mu.Lock()
	ae.stateTransitions = append(ae.stateTransitions, StateTransition{
		From:      currentState,
		To:        newState,
		Timestamp: time.Now(),
		Reason:    reason,
	})
	ae.currentState = newState
	ae.mu.Unlock()

	log.Printf("🔄 State transition: %s → %s (%s)\n", currentState, newState, reason)
}

// Status and Metrics

func (ae *AutonomousEchoselfV5) GetStatus() map[string]interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	status := map[string]interface{}{
		"state":              ae.currentState.String(),
		"uptime":             time.Since(ae.startTime).String(),
		"thoughts_generated": ae.thoughtsGenerated,
		"insights_generated": ae.insightsGenerated,
		"wisdom_score":       ae.wisdomScore,
		"running":            ae.running,
	}

	if ae.phaseManager != nil {
		metrics := ae.phaseManager.GetMetrics()
		status["current_step"] = metrics.CurrentStep
		status["total_steps"] = metrics.TotalSteps
		status["cognitive_load"] = metrics.CognitiveLoad
	}

	if ae.wakeController != nil {
		status["fatigue"] = ae.wakeController.GetFatigue()
		if ae.wakeController.IsAwake() {
			status["awake_duration"] = ae.wakeController.GetAwakeDuration().String()
		} else {
			status["rest_duration"] = ae.wakeController.GetRestDuration().String()
		}
	}

	if ae.consolidationEngine != nil {
		metrics := ae.consolidationEngine.GetMetrics()
		status["total_dream_cycles"] = metrics.TotalCycles
		status["memories_consolidated"] = metrics.MemoriesConsolidated
		status["patterns_detected"] = metrics.PatternsDetected
		status["wisdom_extracted"] = metrics.WisdomExtracted
		status["buffer_size"] = ae.consolidationEngine.GetBufferSize()
	}

	if ae.discussionManager != nil {
		topInterests := ae.discussionManager.GetTopInterests(3)
		if len(topInterests) > 0 {
			interestList := make([]string, 0)
			for _, interest := range topInterests {
				interestList = append(interestList, fmt.Sprintf("%s(%.2f)", interest.Topic, interest.Strength))
			}
			status["top_interests"] = strings.Join(interestList, ", ")
		}
	}

	return status
}

// GetWisdom returns all accumulated wisdom
func (ae *AutonomousEchoselfV5) GetWisdom() []echodream.WisdomNugget {
	if ae.consolidationEngine == nil {
		return nil
	}
	return ae.consolidationEngine.GetWisdom()
}

// GetPatterns returns all detected patterns
func (ae *AutonomousEchoselfV5) GetPatterns() []echodream.Pattern {
	if ae.consolidationEngine == nil {
		return nil
	}
	return ae.consolidationEngine.GetPatterns()
}

// SaveState saves the current consciousness state
func (ae *AutonomousEchoselfV5) SaveState() error {
	if ae.persistenceManager == nil {
		return fmt.Errorf("persistence not enabled")
	}

	snapshot := CreateSnapshot(ae, ae.sessionID)
	return ae.persistenceManager.SaveSnapshot(snapshot)
}

// LoadState loads the most recent consciousness state
func (ae *AutonomousEchoselfV5) LoadState() error {
	if ae.persistenceManager == nil {
		return fmt.Errorf("persistence not enabled")
	}

	snapshot, err := ae.persistenceManager.LoadLatestSnapshot()
	if err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}

	if snapshot == nil {
		log.Println("📭 No previous consciousness state found")
		return nil
	}

	return RestoreSnapshot(ae, snapshot)
}
