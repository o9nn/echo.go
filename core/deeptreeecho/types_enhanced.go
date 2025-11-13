package deeptreeecho

import (
	"sync"
	"time"
)

// CognitiveMode is already defined in consciousness_activation.go

// EnhancedThought extends the base Thought with fields needed for integrated consciousness
type EnhancedThought struct {
	ID               string
	Content          string
	Type             ThoughtType
	Mode             CognitiveMode
	Timestamp        time.Time
	Associations     []string
	EmotionalValence float64
	Importance       float64
	Source           ThoughtSource
	Context          map[string]interface{}
	AARState         *AARState
}

// ToBasicThought converts EnhancedThought to basic Thought for compatibility
func (et *EnhancedThought) ToBasicThought() Thought {
	return Thought{
		ID:               et.ID,
		Content:          et.Content,
		Type:             et.Type,
		Timestamp:        et.Timestamp,
		Associations:     et.Associations,
		EmotionalValence: et.EmotionalValence,
		Importance:       et.Importance,
		Source:           et.Source,
	}
}

// FromBasicThought creates EnhancedThought from basic Thought
func FromBasicThought(t Thought, mode CognitiveMode) *EnhancedThought {
	return &EnhancedThought{
		ID:               t.ID,
		Content:          t.Content,
		Type:             t.Type,
		Mode:             mode,
		Timestamp:        t.Timestamp,
		Associations:     t.Associations,
		EmotionalValence: t.EmotionalValence,
		Importance:       t.Importance,
		Source:           t.Source,
		Context:          make(map[string]interface{}),
	}
}

// WorkingMemory methods for enhanced functionality
func (wm *WorkingMemory) Add(thought *Thought) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	// Add to buffer
	wm.buffer = append(wm.buffer, thought)
	
	// Maintain capacity limit (FIFO)
	if len(wm.buffer) > wm.capacity {
		wm.buffer = wm.buffer[1:]
	}
	
	// Update focus to most recent high-importance thought
	if thought.Importance > 0.7 {
		wm.focus = thought
	}
}

// focusItem sets the current focus of working memory
func (wm *WorkingMemory) focusItem(thought *Thought) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.focus = thought
}

// GetFocus returns the current focus thought
func (wm *WorkingMemory) GetFocus() *Thought {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.focus
}

// GetRecent returns the n most recent thoughts
func (wm *WorkingMemory) GetRecent(n int) []*Thought {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	if len(wm.buffer) == 0 {
		return []*Thought{}
	}
	
	start := len(wm.buffer) - n
	if start < 0 {
		start = 0
	}
	
	// Return copy to avoid race conditions
	recent := make([]*Thought, len(wm.buffer)-start)
	copy(recent, wm.buffer[start:])
	return recent
}

// GetAll returns all thoughts in working memory
func (wm *WorkingMemory) GetAll() []*Thought {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	
	all := make([]*Thought, len(wm.buffer))
	copy(all, wm.buffer)
	return all
}

// Clear clears the working memory buffer
func (wm *WorkingMemory) Clear() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.buffer = make([]*Thought, 0)
	wm.focus = nil
}

// Size returns the current size of the buffer
func (wm *WorkingMemory) Size() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.buffer)
}

// AutonomousStateManager manages wake/rest cycles and cognitive energy
type AutonomousStateManager struct {
	mu                sync.RWMutex
	cognitiveLoad     float64   // 0.0 to 1.0
	energyLevel       float64   // 0.0 to 1.0
	consolidationNeed float64   // 0.0 to 1.0
	restThreshold     float64   // Rest when energy below this
	wakeThreshold     float64   // Wake when energy above this
	lastRest          time.Time
	lastWake          time.Time
	restDuration      time.Duration
	wakeDuration      time.Duration
}

// NewAutonomousStateManager creates a new state manager
func NewAutonomousStateManager() *AutonomousStateManager {
	return &AutonomousStateManager{
		cognitiveLoad:     0.0,
		energyLevel:       1.0,
		consolidationNeed: 0.0,
		restThreshold:     0.3,
		wakeThreshold:     0.7,
		lastWake:          time.Now(),
		wakeDuration:      0,
		restDuration:      0,
	}
}

// UpdateCognitiveLoad updates the cognitive load based on activity
func (asm *AutonomousStateManager) UpdateCognitiveLoad(delta float64) {
	asm.mu.Lock()
	defer asm.mu.Unlock()
	
	asm.cognitiveLoad += delta
	if asm.cognitiveLoad > 1.0 {
		asm.cognitiveLoad = 1.0
	}
	if asm.cognitiveLoad < 0.0 {
		asm.cognitiveLoad = 0.0
	}
	
	// Cognitive load drains energy
	asm.energyLevel -= delta * 0.1
	if asm.energyLevel < 0.0 {
		asm.energyLevel = 0.0
	}
}

// UpdateConsolidationNeed updates the need for memory consolidation
func (asm *AutonomousStateManager) UpdateConsolidationNeed(delta float64) {
	asm.mu.Lock()
	defer asm.mu.Unlock()
	
	asm.consolidationNeed += delta
	if asm.consolidationNeed > 1.0 {
		asm.consolidationNeed = 1.0
	}
	if asm.consolidationNeed < 0.0 {
		asm.consolidationNeed = 0.0
	}
}

// RestoreEnergy restores energy during rest
func (asm *AutonomousStateManager) RestoreEnergy(amount float64) {
	asm.mu.Lock()
	defer asm.mu.Unlock()
	
	asm.energyLevel += amount
	if asm.energyLevel > 1.0 {
		asm.energyLevel = 1.0
	}
}

// ShouldRest determines if the system should enter rest state
func (asm *AutonomousStateManager) ShouldRest() bool {
	asm.mu.RLock()
	defer asm.mu.RUnlock()
	
	// Rest if energy is low OR consolidation need is high
	return asm.energyLevel < asm.restThreshold || asm.consolidationNeed > 0.8
}

// ShouldWake determines if the system should wake from rest
func (asm *AutonomousStateManager) ShouldWake() bool {
	asm.mu.RLock()
	defer asm.mu.RUnlock()
	
	// Wake if energy is restored AND consolidation is complete
	return asm.energyLevel > asm.wakeThreshold && asm.consolidationNeed < 0.2
}

// EnterRest marks the transition to rest state
func (asm *AutonomousStateManager) EnterRest() {
	asm.mu.Lock()
	defer asm.mu.Unlock()
	
	asm.lastRest = time.Now()
	if !asm.lastWake.IsZero() {
		asm.wakeDuration += time.Since(asm.lastWake)
	}
}

// EnterWake marks the transition to wake state
func (asm *AutonomousStateManager) EnterWake() {
	asm.mu.Lock()
	defer asm.mu.Unlock()
	
	asm.lastWake = time.Now()
	if !asm.lastRest.IsZero() {
		asm.restDuration += time.Since(asm.lastRest)
	}
}

// GetState returns the current state metrics
func (asm *AutonomousStateManager) GetState() map[string]interface{} {
	asm.mu.RLock()
	defer asm.mu.RUnlock()
	
	return map[string]interface{}{
		"cognitive_load":     asm.cognitiveLoad,
		"energy_level":       asm.energyLevel,
		"consolidation_need": asm.consolidationNeed,
		"wake_duration":      asm.wakeDuration.String(),
		"rest_duration":      asm.restDuration.String(),
	}
}

// DiscussionManager manages external discussions and interactions
type DiscussionManager struct {
	mu                  sync.RWMutex
	activeDiscussions   map[string]*Discussion
	engagementThreshold float64
	interests           *InterestSystem
}

// Discussion represents an active discussion
type Discussion struct {
	ID            string
	Participants  []string
	Topic         string
	InterestScore float64
	Messages      []*DiscussionMessage
	Active        bool
	StartTime     time.Time
	LastActivity  time.Time
}

// Message type for discussions (different from llm_integration.go Message)
type DiscussionMessage struct {
	ID        string
	From      string
	Content   string
	Timestamp time.Time
}

// NewDiscussionManager creates a new discussion manager
func NewDiscussionManager(interests *InterestSystem) *DiscussionManager {
	return &DiscussionManager{
		activeDiscussions:   make(map[string]*Discussion),
		engagementThreshold: 0.5,
		interests:           interests,
	}
}

// ShouldEngage determines if the system should engage in a discussion about a topic
func (dm *DiscussionManager) ShouldEngage(topic string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	// Calculate interest score for topic
	interestScore := dm.interests.GetInterest(topic)
	
	return interestScore >= dm.engagementThreshold
}

// InitiateDiscussion starts a new discussion
func (dm *DiscussionManager) InitiateDiscussion(topic string, participants []string) *Discussion {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	discussion := &Discussion{
		ID:            generateID(),
		Participants:  participants,
		Topic:         topic,
		InterestScore: dm.interests.GetInterest(topic),
		Messages:      make([]*DiscussionMessage, 0),
		Active:        true,
		StartTime:     time.Now(),
		LastActivity:  time.Now(),
	}
	
	dm.activeDiscussions[discussion.ID] = discussion
	return discussion
}

// AddMessage adds a message to a discussion
func (dm *DiscussionManager) AddMessage(discussionID string, from string, content string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if discussion, exists := dm.activeDiscussions[discussionID]; exists {
		message := &DiscussionMessage{
			ID:        generateID(),
			From:      from,
			Content:   content,
			Timestamp: time.Now(),
		}
		discussion.Messages = append(discussion.Messages, message)
		discussion.LastActivity = time.Now()
	}
}

// GetActiveDiscussions returns all active discussions
func (dm *DiscussionManager) GetActiveDiscussions() []*Discussion {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	
	active := make([]*Discussion, 0)
	for _, discussion := range dm.activeDiscussions {
		if discussion.Active {
			active = append(active, discussion)
		}
	}
	return active
}

// EndDiscussion marks a discussion as inactive
func (dm *DiscussionManager) EndDiscussion(discussionID string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if discussion, exists := dm.activeDiscussions[discussionID]; exists {
		discussion.Active = false
	}
}

// InterestSystem methods for enhanced functionality
func (is *InterestSystem) GetInterest(topic string) float64 {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	if score, exists := is.topics[topic]; exists {
		return score
	}
	return 0.0
}

// UpdateInterest updates the interest score for a topic
func (is *InterestSystem) UpdateInterest(topic string, delta float64) {
	is.mu.Lock()
	defer is.mu.Unlock()
	
	current := is.topics[topic]
	current += delta
	
	// Clamp to [0, 1]
	if current > 1.0 {
		current = 1.0
	}
	if current < 0.0 {
		current = 0.0
	}
	
	is.topics[topic] = current
}

// GetTopInterests returns the top N interests
func (is *InterestSystem) GetTopInterests(n int) []string {
	is.mu.RLock()
	defer is.mu.RUnlock()
	
	// Simple implementation - return first n topics
	// In production, would sort by score
	topics := make([]string, 0, n)
	count := 0
	for topic := range is.topics {
		if count >= n {
			break
		}
		topics = append(topics, topic)
		count++
	}
	return topics
}

// generateID is already defined in identity.go
