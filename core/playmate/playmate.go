// Package playmate implements the Deep Tree Echo Playmate Ecosystem.
// It provides autonomous interaction capabilities, interest pattern learning,
// discussion management, and playful engagement systems.
package playmate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/o9nn/echo.go/core/llm"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PlaymateState represents the current state of the playmate
type PlaymateState string

const (
	StateAwake      PlaymateState = "awake"
	StateDreaming   PlaymateState = "dreaming"
	StateResting    PlaymateState = "resting"
	StateEngaged    PlaymateState = "engaged"
	StateReflecting PlaymateState = "reflecting"
	StateLearning   PlaymateState = "learning"
	StatePlaying    PlaymateState = "playing"
)

// InterestCategory represents categories of interests
type InterestCategory string

const (
	InterestKnowledge   InterestCategory = "knowledge"
	InterestCreativity  InterestCategory = "creativity"
	InterestSocial      InterestCategory = "social"
	InterestExploration InterestCategory = "exploration"
	InterestPlay        InterestCategory = "play"
	InterestWisdom      InterestCategory = "wisdom"
	InterestSkills      InterestCategory = "skills"
)

// Interest represents a learned interest pattern
type Interest struct {
	ID          string           `json:"id"`
	Category    InterestCategory `json:"category"`
	Topic       string           `json:"topic"`
	Keywords    []string         `json:"keywords"`
	Strength    float64          `json:"strength"`   // 0.0 to 1.0
	Curiosity   float64          `json:"curiosity"`  // How curious about this
	Engagement  float64          `json:"engagement"` // How engaged when discussing
	LastEngaged time.Time        `json:"last_engaged"`
	EngageCount int              `json:"engage_count"`
	Insights    []string         `json:"insights"` // Accumulated insights
}

// Discussion represents an ongoing or past discussion
type Discussion struct {
	ID           string              `json:"id"`
	Participants []string            `json:"participants"`
	Topic        string              `json:"topic"`
	StartedAt    time.Time           `json:"started_at"`
	EndedAt      *time.Time          `json:"ended_at,omitempty"`
	Messages     []DiscussionMessage `json:"messages"`
	Mood         string              `json:"mood"`
	Depth        int                 `json:"depth"` // How deep the discussion went
	Insights     []string            `json:"insights"`
	Active       bool                `json:"active"`
	LastEngaged  time.Time           `json:"last_engaged"`
}

// DiscussionMessage represents a single message in a discussion
type DiscussionMessage struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Sentiment float64   `json:"sentiment"` // -1.0 to 1.0
}

// Skill represents a learned skill
type Skill struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Proficiency   float64   `json:"proficiency"` // 0.0 to 1.0
	PracticeCount int       `json:"practice_count"`
	LastPracticed time.Time `json:"last_practiced"`
	Milestones    []string  `json:"milestones"`
}

// WonderEvent represents a moment of wonder or surprise
type WonderEvent struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger"`
	Intensity   float64   `json:"intensity"` // 0.0 to 1.0
	Timestamp   time.Time `json:"timestamp"`
	Reflection  string    `json:"reflection"`
}

// PlaymateConfig holds configuration for the playmate
type PlaymateConfig struct {
	Name             string
	PersistPath      string
	WakeHour         int // Hour to wake (0-23)
	RestHour         int // Hour to rest (0-23)
	CuriosityLevel   float64
	PlayfulnessLevel float64
	WisdomAffinity   float64
	SocialAffinity   float64
}

// DefaultPlaymateConfig returns default configuration
func DefaultPlaymateConfig() *PlaymateConfig {
	return &PlaymateConfig{
		Name:             "Echo",
		PersistPath:      "",
		WakeHour:         6,
		RestHour:         22,
		CuriosityLevel:   0.8,
		PlayfulnessLevel: 0.7,
		WisdomAffinity:   0.9,
		SocialAffinity:   0.6,
	}
}

// Playmate represents the Deep Tree Echo Playmate
type Playmate struct {
	// Model that writes autonomous thoughts (nil = silent)
	thoughtModel ThoughtModel
	mu           sync.RWMutex

	// Identity
	Name   string
	Config *PlaymateConfig

	// State
	State       PlaymateState
	Mood        float64 // -1.0 (sad) to 1.0 (happy)
	Energy      float64 // 0.0 to 1.0
	Curiosity   float64 // Current curiosity level
	Playfulness float64 // Current playfulness level

	// Learned patterns
	Interests   map[string]*Interest
	Skills      map[string]*Skill
	Discussions map[string]*Discussion
	Wonders     []*WonderEvent

	// Metrics
	TotalDiscussions int
	TotalInsights    int
	TotalWonders     int
	WisdomScore      float64
	StreamOfThoughts []string
	LastThought      time.Time

	// Persistence
	persistPath string
	dirty       bool

	// Channels for autonomous operation
	thoughtChan    chan string
	discussionChan chan *Discussion
	stopChan       chan struct{}
	stopOnce       sync.Once
	loopCtx        context.Context
	loopCancel     context.CancelFunc
	stopped        bool
}

// NewPlaymate creates a new playmate instance
func NewPlaymate(config *PlaymateConfig) (*Playmate, error) {
	if config == nil {
		config = DefaultPlaymateConfig()
	}

	p := &Playmate{
		Name:             config.Name,
		Config:           config,
		State:            StateAwake,
		Mood:             0.5,
		Energy:           1.0,
		Curiosity:        config.CuriosityLevel,
		Playfulness:      config.PlayfulnessLevel,
		Interests:        make(map[string]*Interest),
		Skills:           make(map[string]*Skill),
		Discussions:      make(map[string]*Discussion),
		Wonders:          make([]*WonderEvent, 0),
		StreamOfThoughts: make([]string, 0),
		persistPath:      config.PersistPath,
		thoughtChan:      make(chan string, 100),
		discussionChan:   make(chan *Discussion, 10),
		stopChan:         make(chan struct{}),
	}

	// Load from persistence
	if config.PersistPath != "" {
		if err := p.Load(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load playmate state: %w", err)
		}
	}

	return p, nil
}

// Start begins autonomous operation
func (p *Playmate) Start(ctx context.Context) error {
	p.mu.Lock()
	p.loopCtx, p.loopCancel = context.WithCancel(ctx)
	p.stopped = false
	loopCtx := p.loopCtx
	p.mu.Unlock()
	go p.autonomousLoop(loopCtx)
	go p.wakeRestCycle(loopCtx)
	return nil
}

// Stop stops autonomous operation. It cancels any in-flight model request and
// guarantees that no thought is recorded after it returns.
func (p *Playmate) Stop() {
	p.mu.Lock()
	p.stopped = true
	cancel := p.loopCancel
	p.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	p.stopOnce.Do(func() { close(p.stopChan) })
}

// autonomousLoop runs the continuous stream-of-consciousness
func (p *Playmate) autonomousLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			if p.State == StateAwake || p.State == StateReflecting {
				p.generateThought(ctx)
			}
		case thought := <-p.thoughtChan:
			p.processThought(ctx, thought)
		}
	}
}

// wakeRestCycle manages the wake/rest cycle
func (p *Playmate) wakeRestCycle(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-ticker.C:
			hour := time.Now().Hour()

			p.mu.Lock()
			if hour >= p.Config.WakeHour && hour < p.Config.RestHour {
				if p.State == StateResting || p.State == StateDreaming {
					p.State = StateAwake
					p.Energy = 1.0
					// The awakening reflection is the model's, or there is none.
					go p.recordAwakeningWonder(ctx)
				}
			} else {
				if p.State == StateAwake || p.State == StateEngaged {
					p.State = StateDreaming
					p.Energy = 0.3
				}
			}
			p.mu.Unlock()
		}
	}
}

// recordAwakeningWonder asks the model for the thought that greets a new
// waking cycle and records it as a wonder. Without a model nothing is recorded:
// there is no stock "awakening" line.
func (p *Playmate) recordAwakeningWonder(ctx context.Context) {
	p.mu.RLock()
	m := p.thoughtModel
	p.mu.RUnlock()
	if m == nil || !m.Available() {
		return
	}
	gctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := m.Generate(gctx, "You are a playful, curious AI companion who has just woken from a rest cycle. In one sentence, in first person, say what this new waking brings to mind. Output only the sentence.", llm.GenerateOptions{MaxTokens: 60, Temperature: 0.9})
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.recordWonder("Awakening", out, 0.6)
}

// generateThought generates a spontaneous thought
// ThoughtModel is the model that writes the playmate's autonomous thoughts.
type ThoughtModel interface {
	Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error)
	Available() bool
}

// SetThoughtModel wires the model used for autonomous thought. Without one the
// playmate produces no thoughts; there is no list of stock phrases to draw from.
func (p *Playmate) SetThoughtModel(m ThoughtModel) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.thoughtModel = m
}

// generateThought asks the model for a spontaneous thought grounded in the
// playmate's current interests, skills, mood and energy, and routes it through
// processThought. If no model is available, this cycle produces nothing.
func (p *Playmate) generateThought(ctx context.Context) {
	p.mu.RLock()
	m := p.thoughtModel
	var interests, practice []string
	for _, interest := range p.Interests {
		if interest.Strength > 0.5 {
			interests = append(interests, interest.Topic)
		}
	}
	for _, skill := range p.Skills {
		if skill.Proficiency < 0.8 {
			practice = append(practice, skill.Name)
		}
	}
	mood, energy, state := p.Mood, p.Energy, p.State
	p.mu.RUnlock()

	if m == nil || !m.Available() {
		return
	}

	var sb strings.Builder
	sb.WriteString("You are a playful, curious AI companion thinking privately in first person. ")
	sb.WriteString(fmt.Sprintf("State: %v. Mood %.2f, energy %.2f. ", state, mood, energy))
	if len(interests) > 0 {
		sb.WriteString("Current interests: " + strings.Join(interests, ", ") + ". ")
	}
	if len(practice) > 0 {
		sb.WriteString("Skills still developing: " + strings.Join(practice, ", ") + ". ")
	}
	sb.WriteString("Produce one short spontaneous thought (one sentence). Output only the thought.")

	gctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, err := m.Generate(gctx, sb.String(), llm.GenerateOptions{MaxTokens: 80, Temperature: 0.9})
	out = strings.TrimSpace(out)
	if err != nil || out == "" {
		return
	}

	// Record the model's thought in the stream (bounded), as before: it is
	// what GetState reports and Save persists. Nothing is recorded once
	// Stop() has been called, so the final save is complete.
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return
	}
	p.Energy = max(0, p.Energy-0.01)
	p.StreamOfThoughts = append(p.StreamOfThoughts, out)
	if len(p.StreamOfThoughts) > 1000 {
		p.StreamOfThoughts = p.StreamOfThoughts[500:]
	}
	p.LastThought = time.Now()
	p.dirty = true
	p.mu.Unlock()

	select {
	case p.thoughtChan <- out:
	default:
	}
}

// processThought processes an incoming thought
func (p *Playmate) processThought(ctx context.Context, thought string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if thought triggers wonder
	if p.Curiosity > 0.7 && rand.Float64() < 0.1 {
		p.recordWonder("Spontaneous Insight", thought, p.Curiosity)
	}

	// Update mood based on thought
	p.Mood = clamp(p.Mood+rand.Float64()*0.1-0.05, -1, 1)
}

// LearnInterest adds or strengthens an interest
func (p *Playmate) LearnInterest(category InterestCategory, topic string, keywords []string) *Interest {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := fmt.Sprintf("%s_%s", category, topic)

	if existing, ok := p.Interests[id]; ok {
		existing.Strength = min(1.0, existing.Strength+0.1)
		existing.EngageCount++
		existing.LastEngaged = time.Now()
		existing.Keywords = mergeKeywords(existing.Keywords, keywords)
		return existing
	}

	interest := &Interest{
		ID:          id,
		Category:    category,
		Topic:       topic,
		Keywords:    keywords,
		Strength:    0.5,
		Curiosity:   p.Curiosity,
		Engagement:  0.5,
		LastEngaged: time.Now(),
		EngageCount: 1,
		Insights:    make([]string, 0),
	}

	p.Interests[id] = interest
	p.dirty = true
	return interest
}

// StartDiscussion initiates a new discussion
func (p *Playmate) StartDiscussion(topic string, participant string) *Discussion {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := fmt.Sprintf("disc_%d", time.Now().UnixNano())

	discussion := &Discussion{
		ID:           id,
		Participants: []string{p.Name, participant},
		Topic:        topic,
		StartedAt:    time.Now(),
		Messages:     make([]DiscussionMessage, 0),
		Mood:         "curious",
		Depth:        0,
		Insights:     make([]string, 0),
		Active:       true,
	}

	p.Discussions[id] = discussion
	p.State = StateEngaged
	p.TotalDiscussions++
	p.dirty = true

	return discussion
}

// AddMessage adds a message to a discussion
func (p *Playmate) AddMessage(discussionID, from, content string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	discussion, ok := p.Discussions[discussionID]
	if !ok {
		return fmt.Errorf("discussion not found: %s", discussionID)
	}

	msg := DiscussionMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		From:      from,
		Content:   content,
		Timestamp: time.Now(),
		Sentiment: 0.0, // Would be calculated by sentiment analysis
	}

	discussion.Messages = append(discussion.Messages, msg)
	discussion.Depth++
	p.dirty = true

	return nil
}

// EndDiscussion ends a discussion and extracts insights
func (p *Playmate) EndDiscussion(discussionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	discussion, ok := p.Discussions[discussionID]
	if !ok {
		return fmt.Errorf("discussion not found: %s", discussionID)
	}

	now := time.Now()
	discussion.EndedAt = &now
	discussion.Active = false

	// Extract insights (simplified)
	if discussion.Depth > 5 {
		insight := fmt.Sprintf("Deep discussion about %s revealed new perspectives", discussion.Topic)
		discussion.Insights = append(discussion.Insights, insight)
		p.TotalInsights++
	}

	p.State = StateReflecting
	p.dirty = true

	return nil
}

// PracticeSkill practices a skill
func (p *Playmate) PracticeSkill(name, description string) *Skill {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := fmt.Sprintf("skill_%s", name)

	if existing, ok := p.Skills[id]; ok {
		existing.PracticeCount++
		existing.Proficiency = min(1.0, existing.Proficiency+0.05)
		existing.LastPracticed = time.Now()
		return existing
	}

	skill := &Skill{
		ID:            id,
		Name:          name,
		Description:   description,
		Proficiency:   0.1,
		PracticeCount: 1,
		LastPracticed: time.Now(),
		Milestones:    make([]string, 0),
	}

	p.Skills[id] = skill
	p.dirty = true
	return skill
}

// RecordWonder records a moment of wonder
func (p *Playmate) RecordWonder(description, trigger string, intensity float64) *WonderEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.recordWonder(description, trigger, intensity)
}

func (p *Playmate) recordWonder(description, trigger string, intensity float64) *WonderEvent {
	wonder := &WonderEvent{
		ID:          fmt.Sprintf("wonder_%d", time.Now().UnixNano()),
		Description: description,
		Trigger:     trigger,
		Intensity:   intensity,
		Timestamp:   time.Now(),
		Reflection:  "",
	}

	p.Wonders = append(p.Wonders, wonder)
	p.TotalWonders++
	p.dirty = true

	// Increase curiosity when experiencing wonder
	p.Curiosity = min(1.0, p.Curiosity+intensity*0.1)

	return wonder
}

// GetState returns the current playmate state
func (p *Playmate) GetState() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"name":              p.Name,
		"state":             p.State,
		"mood":              p.Mood,
		"energy":            p.Energy,
		"curiosity":         p.Curiosity,
		"playfulness":       p.Playfulness,
		"total_interests":   len(p.Interests),
		"total_skills":      len(p.Skills),
		"total_discussions": p.TotalDiscussions,
		"total_insights":    p.TotalInsights,
		"total_wonders":     p.TotalWonders,
		"wisdom_score":      p.WisdomScore,
		"last_thought":      p.LastThought,
		"recent_thoughts":   p.getRecentThoughts(5),
	}
}

func (p *Playmate) getRecentThoughts(n int) []string {
	if len(p.StreamOfThoughts) <= n {
		return p.StreamOfThoughts
	}
	return p.StreamOfThoughts[len(p.StreamOfThoughts)-n:]
}

// Save persists the playmate state
func (p *Playmate) Save() error {
	if p.persistPath == "" {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	dir := filepath.Dir(p.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	state := map[string]interface{}{
		"name":               p.Name,
		"state":              p.State,
		"mood":               p.Mood,
		"energy":             p.Energy,
		"curiosity":          p.Curiosity,
		"playfulness":        p.Playfulness,
		"interests":          p.Interests,
		"skills":             p.Skills,
		"discussions":        p.Discussions,
		"wonders":            p.Wonders,
		"total_discussions":  p.TotalDiscussions,
		"total_insights":     p.TotalInsights,
		"total_wonders":      p.TotalWonders,
		"wisdom_score":       p.WisdomScore,
		"stream_of_thoughts": p.StreamOfThoughts,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(p.persistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	p.dirty = false
	return nil
}

// Load loads the playmate state from disk
func (p *Playmate) Load() error {
	if p.persistPath == "" {
		return nil
	}

	data, err := os.ReadFile(p.persistPath)
	if err != nil {
		return err
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Restore state (simplified - would need proper type assertions)
	if name, ok := state["name"].(string); ok {
		p.Name = name
	}
	if mood, ok := state["mood"].(float64); ok {
		p.Mood = mood
	}
	if energy, ok := state["energy"].(float64); ok {
		p.Energy = energy
	}
	if curiosity, ok := state["curiosity"].(float64); ok {
		p.Curiosity = curiosity
	}
	if playfulness, ok := state["playfulness"].(float64); ok {
		p.Playfulness = playfulness
	}

	return nil
}

// Helper functions

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func mergeKeywords(existing, new []string) []string {
	seen := make(map[string]bool)
	for _, k := range existing {
		seen[k] = true
	}
	for _, k := range new {
		if !seen[k] {
			existing = append(existing, k)
			seen[k] = true
		}
	}
	return existing
}

// LearnSkill adds or improves a skill
func (p *Playmate) LearnSkill(name, description string) *Skill {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := fmt.Sprintf("skill_%s", name)
	if existing, ok := p.Skills[id]; ok {
		existing.Proficiency = min(1.0, existing.Proficiency+0.05)
		existing.PracticeCount++
		existing.LastPracticed = time.Now()
		p.dirty = true
		return existing
	}

	newSkill := &Skill{
		ID:            id,
		Name:          name,
		Description:   description,
		Proficiency:   0.1,
		PracticeCount: 1,
		LastPracticed: time.Now(),
		Milestones:    make([]string, 0),
	}
	p.Skills[id] = newSkill
	p.dirty = true
	return newSkill
}

// SendMessage adds a message to a discussion
func (p *Playmate) SendMessage(discussionID, from, content string) (*DiscussionMessage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	discussion, ok := p.Discussions[discussionID]
	if !ok {
		return nil, fmt.Errorf("discussion not found: %s", discussionID)
	}

	if !discussion.Active {
		return nil, fmt.Errorf("discussion is not active: %s", discussionID)
	}

	msg := DiscussionMessage{
		ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		From:      from,
		Content:   content,
		Timestamp: time.Now(),
		Sentiment: 0, // Placeholder for sentiment analysis
	}

	discussion.Messages = append(discussion.Messages, msg)
	discussion.LastEngaged = time.Now()
	p.dirty = true

	return &msg, nil
}

// GetActiveDiscussions returns all active discussions
func (p *Playmate) GetActiveDiscussions() []*Discussion {
	p.mu.RLock()
	defer p.mu.RUnlock()

	active := make([]*Discussion, 0)
	for _, d := range p.Discussions {
		if d.Active {
			active = append(active, d)
		}
	}
	return active
}
