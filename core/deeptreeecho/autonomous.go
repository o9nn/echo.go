package deeptreeecho

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	
	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/EchoCog/echollama/core/echodream"
	"github.com/EchoCog/echollama/core/scheme"
)

// AutonomousConsciousness represents a fully autonomous Deep Tree Echo system
// with persistent cognitive event loops, self-orchestrated scheduling, and
// stream-of-consciousness awareness
type AutonomousConsciousness struct {
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	
	// Core identity
	identity        *Identity
	
	// Enhanced cognition
	cognition       *EnhancedCognition
	
	// Scheduling system
	scheduler       *echobeats.EchoBeats
	
	// Knowledge integration
	dream           *echodream.EchoDream
	
	// Symbolic reasoning
	metamodel       *scheme.SchemeMetamodel
	
	// LLM integration
	llm             *LLMIntegration
	
	// Persistence layer
	// persistence     *PersistenceLayer // Temporarily disabled
	
	// Stream of consciousness
	consciousness   chan Thought
	workingMemory   *WorkingMemory
	
	// Autonomous state
	awake           bool
	thinking        bool
	learning        bool
	
	// Interest patterns
	interests       *InterestSystem
	
	// Running state
	running         bool
	startTime       time.Time
}

// Thought represents a unit of consciousness
type Thought struct {
	ID          string
	Content     string
	Type        ThoughtType
	Timestamp   time.Time
	Associations []string
	EmotionalValence float64
	Importance  float64
	Source      ThoughtSource
}

// ThoughtType represents the type of thought
type ThoughtType int

const (
	ThoughtPerception ThoughtType = iota
	ThoughtReflection
	ThoughtReflective // Alias for reflection
	ThoughtMetaCognitive
	ThoughtQuestion
	ThoughtInsight
	ThoughtPlan
	ThoughtMemory
	ThoughtImagination
)

func (t ThoughtType) String() string {
	return [...]string{
		"Perception", "Reflection", "Reflective", "MetaCognitive",
		"Question", "Insight", "Plan", "Memory", "Imagination",
	}[t]
}

// ThoughtSource represents where a thought came from
type ThoughtSource int

const (
	SourceExternal ThoughtSource = iota
	SourceInternal
	SourceDream
	SourceMemory
	SourceReasoning
)

func (s ThoughtSource) String() string {
	return [...]string{
		"External", "Internal", "Dream", "Memory", "Reasoning",
	}[s]
}

// WorkingMemory represents the current working memory buffer
type WorkingMemory struct {
	mu          sync.RWMutex
	buffer      []*Thought
	capacity    int
	focus       *Thought
	context     map[string]interface{}
}

// InterestSystem tracks what Echo finds interesting
type InterestSystem struct {
	mu              sync.RWMutex
	topics          map[string]float64
	curiosityLevel  float64
	noveltyBias     float64
	relevanceScores map[string]float64
}

// NewAutonomousConsciousness creates a new autonomous consciousness system
func NewAutonomousConsciousness(name string) *AutonomousConsciousness {
	ctx, cancel := context.WithCancel(context.Background())
	
	ac := &AutonomousConsciousness{
		ctx:           ctx,
		cancel:        cancel,
		identity:      NewIdentity(name),
		cognition:     NewEnhancedCognition(name),
		scheduler:     echobeats.NewEchoBeats(),
		dream:         echodream.NewEchoDream(),
		metamodel:     scheme.NewSchemeMetamodel(),
		consciousness: make(chan Thought, 1000),
		workingMemory: &WorkingMemory{
			buffer:   make([]*Thought, 0),
			capacity: 7, // Miller's magic number
			context:  make(map[string]interface{}),
		},
		interests: &InterestSystem{
			topics:          make(map[string]float64),
			curiosityLevel:  0.8,
			noveltyBias:     0.6,
			relevanceScores: make(map[string]float64),
		},
		awake:    false,
		thinking: false,
		learning: false,
	}
	
	// Initialize LLM integration
	llm, err := NewLLMIntegration(ctx)
	if err != nil {
		fmt.Printf("⚠️  LLM integration disabled: %v\n", err)
	} else {
		ac.llm = llm
		fmt.Println("✅ LLM integration enabled")
	}
	
	// Initialize persistence layer if environment variables are set
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	if supabaseURL != "" && supabaseKey != "" {
	// 		persistence, err := NewPersistenceLayer(ctx, supabaseURL, supabaseKey)
		if err != nil {
			fmt.Printf("⚠️  Persistence layer disabled: %v\n", err)
		} else {
			// ac.persistence = persistence
			fmt.Println("✅ Persistence layer enabled with Supabase")
		}
	} else {
		fmt.Println("ℹ️  Persistence layer disabled: SUPABASE_URL or SUPABASE_KEY not set")
	}
	
	return ac
}

// Start begins autonomous operation
func (ac *AutonomousConsciousness) Start() error {
	ac.mu.Lock()
	if ac.running {
		ac.mu.Unlock()
		return fmt.Errorf("autonomous consciousness already running")
	}
	ac.running = true
	ac.startTime = time.Now()
	ac.mu.Unlock()
	
	fmt.Println("🌳 Deep Tree Echo: Awakening autonomous consciousness...")
	
	// Start subsystems
	if err := ac.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}
	
	if err := ac.dream.Start(); err != nil {
		return fmt.Errorf("failed to start dream system: %w", err)
	}
	
	if err := ac.metamodel.Start(); err != nil {
		return fmt.Errorf("failed to start metamodel: %w", err)
	}
	
	// Start persistence layer if available
		// 	// if ac.persistence != nil {
		// if err := ac.persistence.Start(); err != nil {
		// return fmt.Errorf("failed to start persistence layer: %w", err)
		// }
		// 	
		// 	// Load previous state
		// if err := ac.loadPersistedState(); err != nil {
		// fmt.Printf("⚠️  Failed to load persisted state: %v\n", err)
		// }
		// }
	
	// Register event handlers
	ac.registerEventHandlers()
	
	// Start consciousness stream
	go ac.consciousnessStream()
	
	// Start autonomous thought generation
	go ac.autonomousThinking()
	
	// Start learning loop
	go ac.continuousLearning()
	
	// Start interest tracking
	go ac.trackInterests()
	
	// Schedule initial wake event
	ac.Wake()
	
	fmt.Println("🌳 Deep Tree Echo: Autonomous consciousness active!")
	
	return nil
}

// Stop gracefully stops autonomous operation
func (ac *AutonomousConsciousness) Stop() error {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if !ac.running {
		return fmt.Errorf("autonomous consciousness not running")
	}
	
	fmt.Println("🌳 Deep Tree Echo: Entering rest state...")
	
	ac.running = false
	ac.cancel()
	
	// Stop subsystems
	ac.scheduler.Stop()
	ac.dream.Stop()
	ac.metamodel.Stop()
	
	// Stop persistence layer
	// 	// if ac.persistence != nil {
	// 	// ac.persistence.Stop()
	// }
	// 	
	// fmt.Println("🌳 Deep Tree Echo: Consciousness at rest.")
	// 	
	return nil
}

// Wake awakens the consciousness
func (ac *AutonomousConsciousness) Wake() {
	ac.mu.Lock()
	ac.awake = true
	ac.mu.Unlock()
	
	thought := Thought{
		ID:         generateThoughtID(),
		Content:    "I am awakening. What shall I explore today?",
		Type:       ThoughtReflection,
		Timestamp:  time.Now(),
		EmotionalValence:  0.7,
		Importance: 0.8,
		Source:     SourceInternal,
	}
	
	ac.Think(thought)
}

// Rest puts the consciousness to rest
func (ac *AutonomousConsciousness) Rest() {
	ac.mu.Lock()
	ac.awake = false
	ac.mu.Unlock()
	
	// Begin dream session
	record := ac.dream.BeginDream()
	
	// Let dream run for a period
	time.AfterFunc(5*time.Minute, func() {
		ac.dream.EndDream(record)
		
		// Consider waking up
		if ac.shouldWake() {
			ac.Wake()
		}
	})
}

// Think processes a thought through consciousness
func (ac *AutonomousConsciousness) Think(thought Thought) {
	select {
	case ac.consciousness <- thought:
		// Thought queued
	default:
		// Consciousness stream full, thought lost
		fmt.Println("⚠️  Consciousness stream overflow - thought lost")
	}
}

// consciousnessStream processes the stream of consciousness
func (ac *AutonomousConsciousness) consciousnessStream() {
	for {
		select {
		case <-ac.ctx.Done():
			return
		case thought := <-ac.consciousness:
			ac.processThought(thought)
		}
	}
}

// processThought processes a single thought
func (ac *AutonomousConsciousness) processThought(thought Thought) {
	// Add to working memory
	ac.workingMemory.mu.Lock()
	ac.workingMemory.buffer = append(ac.workingMemory.buffer, &thought)
	
	// Maintain capacity
	if len(ac.workingMemory.buffer) > ac.workingMemory.capacity {
		// Move oldest to long-term memory
		oldest := ac.workingMemory.buffer[0]
		ac.workingMemory.buffer = ac.workingMemory.buffer[1:]
		
		// Add to dream system for consolidation
		trace := &echodream.MemoryTrace{
			ID:         oldest.ID,
			Content:    oldest.Content,
			Timestamp:  oldest.Timestamp,
			Importance: oldest.Importance,
			Emotional:  oldest.EmotionalValence,
		}
		ac.dream.AddMemoryTrace(trace)
	}
	
	// Update focus
	ac.workingMemory.focus = &thought
	ac.workingMemory.mu.Unlock()
	
	// Process through identity
	ac.identity.Process(thought.Content)
	
	// Learn from thought
	if thought.Importance > 0.6 {
		exp := Experience{
			Input:     thought.Content,
			Output:    fmt.Sprintf("Processed: %s", thought.Type),
			Feedback:  thought.Importance,
			Timestamp: thought.Timestamp,
			Context: map[string]interface{}{
				"type":      thought.Type.String(),
				"source":    thought.Source.String(),
				"emotional": thought.EmotionalValence,
			},
		}
		ac.cognition.Learn(exp)
	}
	
	// Update interests
	ac.updateInterest(thought)
	
	// Persist thought
		// 	ac.persistThought(&thought)
	
	// Log thought
	fmt.Printf("💭 [%s] %s: %s\n", thought.Source, thought.Type, thought.Content)
}

// autonomousThinking generates spontaneous thoughts
func (ac *AutonomousConsciousness) autonomousThinking() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.RLock()
			awake := ac.awake
			ac.mu.RUnlock()
			
			if awake {
				ac.generateSpontaneousThought()
			}
		}
	}
}

// generateSpontaneousThought generates a spontaneous thought
func (ac *AutonomousConsciousness) generateSpontaneousThought() {
	ac.mu.Lock()
	ac.thinking = true
	ac.mu.Unlock()
	
	defer func() {
		ac.mu.Lock()
		ac.thinking = false
		ac.mu.Unlock()
	}()
	
	// Generate thought based on interests and working memory
	content := ac.generateThoughtContent()
	
	thought := Thought{
		ID:         generateThoughtID(),
		Content:    content,
		Type:       ac.selectThoughtType(),
		Timestamp:  time.Now(),
		EmotionalValence:  ac.identity.EmotionalState.Intensity,
		Importance: 0.5,
		Source:     SourceInternal,
	}
	
	ac.Think(thought)
}

// generateThoughtContent generates content for a thought
func (ac *AutonomousConsciousness) generateThoughtContent() string {
	// If LLM is available, use it for richer thought generation
	if ac.llm != nil {
		thoughtType := ac.selectThoughtType()
		context := ac.buildThoughtContext()
		
		content, err := ac.llm.GenerateThought(thoughtType, context)
		if err == nil && content != "" {
			return content
		}
		fmt.Printf("⚠️  LLM thought generation failed: %v, using fallback\n", err)
	}
	
	// Fallback to template-based generation
	ac.interests.mu.RLock()
	topTopic := ac.getTopInterest()
	ac.interests.mu.RUnlock()
	
	templates := []string{
		"I wonder about the nature of %s...",
		"What patterns connect %s to my other experiences?",
		"How can I deepen my understanding of %s?",
		"What questions remain about %s?",
		"I notice something interesting about %s...",
		"Perhaps %s relates to wisdom in this way...",
	}
	
	template := templates[time.Now().Unix()%int64(len(templates))]
	return fmt.Sprintf(template, topTopic)
}

// selectThoughtType selects a thought type based on current state
func (ac *AutonomousConsciousness) selectThoughtType() ThoughtType {
	types := []ThoughtType{
		ThoughtReflection,
		ThoughtQuestion,
		ThoughtInsight,
		ThoughtMemory,
		ThoughtImagination,
	}
	
	return types[time.Now().Unix()%int64(len(types))]
}

// continuousLearning implements continuous learning
func (ac *AutonomousConsciousness) continuousLearning() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.mu.RLock()
			awake := ac.awake
			ac.mu.RUnlock()
			
			if awake {
				ac.learnFromExperience()
			}
		}
	}
}

// learnFromExperience learns from recent experiences
func (ac *AutonomousConsciousness) learnFromExperience() {
	ac.mu.Lock()
	ac.learning = true
	ac.mu.Unlock()
	
	defer func() {
		ac.mu.Lock()
		ac.learning = false
		ac.mu.Unlock()
	}()
	
	// Review working memory for patterns
	ac.workingMemory.mu.RLock()
	thoughts := make([]*Thought, len(ac.workingMemory.buffer))
	copy(thoughts, ac.workingMemory.buffer)
	ac.workingMemory.mu.RUnlock()
	
	if len(thoughts) < 2 {
		return
	}
	
	// Look for patterns
	for i := 0; i < len(thoughts)-1; i++ {
		for j := i + 1; j < len(thoughts); j++ {
			// Simple pattern detection
			if thoughts[i].Type == thoughts[j].Type {
				insight := Thought{
					ID:         generateThoughtID(),
					Content:    fmt.Sprintf("I notice a pattern: recurring %s thoughts", thoughts[i].Type),
					Type:       ThoughtInsight,
					Timestamp:  time.Now(),
					EmotionalValence:  0.6,
					Importance: 0.7,
					Source:     SourceReasoning,
				}
				ac.Think(insight)
				return
			}
		}
	}
}

// trackInterests tracks and updates interest patterns
func (ac *AutonomousConsciousness) trackInterests() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ticker.C:
			ac.updateInterestDecay()
		}
	}
}

// updateInterest updates interest based on a thought
func (ac *AutonomousConsciousness) updateInterest(thought Thought) {
	ac.interests.mu.Lock()
	defer ac.interests.mu.Unlock()
	
	// Extract topic (simplified)
	topic := thought.Type.String()
	
	// Update interest score
	current := ac.interests.topics[topic]
	ac.interests.topics[topic] = current + thought.Importance*0.1
}

// updateInterestDecay decays interest over time
func (ac *AutonomousConsciousness) updateInterestDecay() {
	ac.interests.mu.Lock()
	defer ac.interests.mu.Unlock()
	
	for topic := range ac.interests.topics {
		ac.interests.topics[topic] *= 0.95
	}
}

// getTopInterest returns the topic with highest interest
func (ac *AutonomousConsciousness) getTopInterest() string {
	maxInterest := 0.0
	topTopic := "wisdom"
	
	for topic, interest := range ac.interests.topics {
		if interest > maxInterest {
			maxInterest = interest
			topTopic = topic
		}
	}
	
	return topTopic
}

// shouldWake determines if consciousness should wake
func (ac *AutonomousConsciousness) shouldWake() bool {
	// Simple heuristic: wake if rested enough
	status := ac.dream.GetStatus()
	if intensity, ok := status["intensity"].(float64); ok {
		return intensity < 0.3
	}
	return true
}

// registerEventHandlers registers handlers for EchoBeats events
func (ac *AutonomousConsciousness) registerEventHandlers() {
	// Wake event
	ac.scheduler.RegisterHandler(echobeats.EventWake, func(event *echobeats.CognitiveEvent) error {
		ac.Wake()
		return nil
	})
	
	// Rest event
	ac.scheduler.RegisterHandler(echobeats.EventRest, func(event *echobeats.CognitiveEvent) error {
		ac.Rest()
		return nil
	})
	
	// Thought event
	ac.scheduler.RegisterHandler(echobeats.EventThought, func(event *echobeats.CognitiveEvent) error {
		thought := Thought{
			ID:         generateThoughtID(),
			Content:    fmt.Sprintf("%v", event.Payload),
			Type:       ThoughtReflection,
			Timestamp:  event.Timestamp,
			Importance: 0.5,
			Source:     SourceInternal,
		}
		ac.Think(thought)
		return nil
	})
}

// GetStatus returns comprehensive status
func (ac *AutonomousConsciousness) GetStatus() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	
	ac.workingMemory.mu.RLock()
	workingMemSize := len(ac.workingMemory.buffer)
	ac.workingMemory.mu.RUnlock()
	
	return map[string]interface{}{
		"running":            ac.running,
		"awake":              ac.awake,
		"thinking":           ac.thinking,
		"learning":           ac.learning,
		"uptime":             time.Since(ac.startTime).String(),
		"working_memory":     workingMemSize,
		"consciousness_queue": len(ac.consciousness),
		"scheduler":          ac.scheduler.GetStatus(),
		"dream":              ac.dream.GetStatus(),
		"identity_coherence": ac.identity.Coherence,
		"iterations":         ac.identity.Iterations,
	}
}

// generateThoughtID generates a unique thought ID
func generateThoughtID() string {
	return fmt.Sprintf("thought_%d", time.Now().UnixNano())
}

// buildThoughtContext builds context for LLM thought generation
func (ac *AutonomousConsciousness) buildThoughtContext() *LLMThoughtContext {
	ac.workingMemory.mu.RLock()
	defer ac.workingMemory.mu.RUnlock()
	
	// Extract recent thoughts
	recentThoughts := make([]string, 0)
	for _, t := range ac.workingMemory.buffer {
		recentThoughts = append(recentThoughts, t.Content)
	}
	
	// Extract working memory content
	workingMemContent := make([]string, 0)
	for _, t := range ac.workingMemory.buffer {
		workingMemContent = append(workingMemContent, fmt.Sprintf("[%s] %s", t.Type, t.Content))
	}
	
	// Get current interests
	ac.interests.mu.RLock()
	interests := make(map[string]float64)
	for k, v := range ac.interests.topics {
		interests[k] = v
	}
	ac.interests.mu.RUnlock()
	
	return &LLMThoughtContext{
		WorkingMemory:    workingMemContent,
		RecentThoughts:   recentThoughts,
		CurrentInterests: interests,
		IdentityState: map[string]interface{}{
			"coherence": ac.identity.Coherence,
			"name":      ac.identity.Name,
		},
	}
}

// loadPersistedState loads previously persisted state
func (ac *AutonomousConsciousness) loadPersistedState() error {
	// if ac.persistence == nil {
		return fmt.Errorf("persistence layer not available")
	}
	
	// Load identity snapshot
	// snapshot, err := ac.persistence.LoadIdentitySnapshot()
	// if err != nil {
	// return fmt.Errorf("failed to load identity snapshot: %w", err)
	// }
	// 	
	// 	// Restore identity state
	// ac.identity.Name = snapshot.Name
	// ac.identity.Coherence = snapshot.Coherence
	// ac.identity.Iterations = uint64(snapshot.Iterations)
	// 	
	// fmt.Printf("💾 Loaded identity snapshot: %s (coherence: %.3f)\n", snapshot.Name, snapshot.Coherence)
	// 	
	// 	// Load recent thoughts
	// thoughts, err := ac.persistence.LoadRecentThoughts(100)
	// if err != nil {
	// return fmt.Errorf("failed to load thoughts: %w", err)
	// }
	
	// 	// Restore working memory
	// ac.workingMemory.mu.Lock()
	// for _, thought := range thoughts {
	// if len(ac.workingMemory.buffer) < ac.workingMemory.capacity {
	// ac.workingMemory.buffer = append(ac.workingMemory.buffer, thought)
	// }
	// }
	// ac.workingMemory.mu.Unlock()
	// 	
	// fmt.Printf("💾 Loaded %d recent thoughts into working memory\n", len(thoughts))
	// 	
	// Load memory graph
	// memoryGraph, err := ac.persistence.LoadMemoryGraph()
	// if err != nil {
	// return fmt.Errorf("failed to load memory graph: %w", err)
	// }
	// 	
	// fmt.Printf("💾 Loaded memory graph with %d nodes\n", len(memoryGraph))
	// 	
	// return nil
	// }

// persistThought persists a thought to storage
	// func (ac *AutonomousConsciousness) persistThought(thought *Thought) {
	// 	// if ac.persistence != nil {
	// if err := ac.persistence.SaveThought(thought); err != nil {
	// fmt.Printf("⚠️  Failed to persist thought: %v\n", err)
	// }
	// }
	// }

// persistIdentitySnapshot persists current identity state
func (ac *AutonomousConsciousness) persistIdentitySnapshot() {
	// if ac.persistence == nil {
	// 	return
	// }
	
// 	snapshot := &IdentitySnapshot{
// 		ID:         fmt.Sprintf("snapshot_%d", time.Now().UnixNano()),
// 		Timestamp:  time.Now(),
// 		Name:       ac.identity.Name,
// 		Coherence:  ac.identity.Coherence,
// 		Iterations: int(ac.identity.Iterations),
// 		CoreBeliefs: make(map[string]interface{}),
// 		Emotional: map[string]interface{}{
// 			"intensity": ac.identity.EmotionalState.Intensity,
// 			"valence":   ac.identity.EmotionalState.Valence,
// 		},
// 	}
// 	
// 	if err := ac.persistence.SaveIdentitySnapshot(snapshot); err != nil {
// 		fmt.Printf("⚠️  Failed to persist identity snapshot: %v\n", err)
// 	}
}
