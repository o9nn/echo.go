package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// EchobeatsSchedulerV2 is an enhanced goal-directed scheduling system
// that orchestrates cognitive event loops with self-directed timing
type EchobeatsSchedulerV2 struct {
	mu sync.RWMutex
	ctx context.Context
	cancel context.CancelFunc

	// LLM provider
	llmProvider llm.LLMProvider

	// Goal management
	goals           []ScheduledGoalV2
	activeGoal      *ScheduledGoalV2
	completedGoals  []ScheduledGoalV2

	// Scheduling state
	nextBeatTime    time.Time
	beatInterval    time.Duration
	adaptiveRate    float64  // Multiplier for beat interval based on cognitive load

	// Cognitive rhythm
	rhythmPhase     RhythmPhase
	phaseStartTime  time.Time
	phaseDuration   time.Duration

	// Self-orchestration
	selfDirected    bool
	urgencyLevel    float64
	focusIntensity  float64

	// Metrics
	totalBeats      uint64
	goalsCompleted  uint64
	goalsAbandoned  uint64
	avgGoalDuration time.Duration

	// Running state
	running         bool
}

// ScheduledGoalV2 represents a goal with scheduling metadata (enhanced version)
type ScheduledGoalV2 struct {
	ID              string
	Description     string
	Priority        float64
	Progress        float64
	Status          GoalStatusV2
	CreatedAt       time.Time
	StartedAt       time.Time
	CompletedAt     time.Time
	Deadline        time.Time
	EstimatedTime   time.Duration
	ActualTime      time.Duration
	SubGoals        []string
	Dependencies    []string
	Tags            []string
	Metadata        map[string]interface{}
}

// GoalStatusV2 represents the status of a goal (enhanced version)
type GoalStatusV2 int

const (
	GoalPendingV2 GoalStatusV2 = iota
	GoalActiveV2
	GoalPausedV2
	GoalCompletedV2
	GoalAbandonedV2
)

func (gs GoalStatusV2) String() string {
	return [...]string{
		"Pending",
		"Active",
		"Paused",
		"Completed",
		"Abandoned",
	}[gs]
}

// RhythmPhase represents phases in the cognitive rhythm
type RhythmPhase int

const (
	RhythmFocus RhythmPhase = iota
	RhythmExplore
	RhythmIntegrate
	RhythmRest
)

func (rp RhythmPhase) String() string {
	return [...]string{
		"Focus",
		"Explore",
		"Integrate",
		"Rest",
	}[rp]
}

// NewEchobeatsSchedulerV2 creates a new enhanced scheduler
func NewEchobeatsSchedulerV2(llmProvider llm.LLMProvider) *EchobeatsSchedulerV2 {
	ctx, cancel := context.WithCancel(context.Background())

	return &EchobeatsSchedulerV2{
		ctx:            ctx,
		cancel:         cancel,
		llmProvider:    llmProvider,
		goals:          make([]ScheduledGoalV2, 0),
		completedGoals: make([]ScheduledGoalV2, 0),
		beatInterval:   10 * time.Second,
		adaptiveRate:   1.0,
		rhythmPhase:    RhythmFocus,
		selfDirected:   true,
		urgencyLevel:   0.5,
		focusIntensity: 0.5,
	}
}

// Start begins the scheduler
func (es *EchobeatsSchedulerV2) Start() error {
	es.mu.Lock()
	if es.running {
		es.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	es.running = true
	es.phaseStartTime = time.Now()
	es.phaseDuration = es.calculatePhaseDuration()
	es.mu.Unlock()

	go es.runScheduler()
	return nil
}

// Stop stops the scheduler
func (es *EchobeatsSchedulerV2) Stop() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.running {
		return fmt.Errorf("scheduler not running")
	}

	es.cancel()
	es.running = false
	return nil
}

// runScheduler is the main scheduler loop
func (es *EchobeatsSchedulerV2) runScheduler() {
	ticker := time.NewTicker(es.beatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-es.ctx.Done():
			return
		case <-ticker.C:
			es.processBeat()
			
			// Adapt the ticker interval
			es.mu.RLock()
			newInterval := time.Duration(float64(es.beatInterval) * es.adaptiveRate)
			es.mu.RUnlock()
			
			ticker.Reset(newInterval)
		}
	}
}

// processBeat processes a single beat
func (es *EchobeatsSchedulerV2) processBeat() {
	es.mu.Lock()
	es.totalBeats++
	es.mu.Unlock()

	// Check for phase transition
	es.checkPhaseTransition()

	// Process based on current phase
	switch es.rhythmPhase {
	case RhythmFocus:
		es.processFocusPhase()
	case RhythmExplore:
		es.processExplorePhase()
	case RhythmIntegrate:
		es.processIntegratePhase()
	case RhythmRest:
		es.processRestPhase()
	}

	// Update adaptive rate based on cognitive state
	es.updateAdaptiveRate()
}

// checkPhaseTransition checks if it's time to transition phases
func (es *EchobeatsSchedulerV2) checkPhaseTransition() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if time.Since(es.phaseStartTime) > es.phaseDuration {
		// Transition to next phase
		es.rhythmPhase = (es.rhythmPhase + 1) % 4
		es.phaseStartTime = time.Now()
		es.phaseDuration = es.calculatePhaseDuration()

		fmt.Printf("🎵 Rhythm phase transition: %s (duration: %v)\n", 
			es.rhythmPhase.String(), es.phaseDuration)
	}
}

// calculatePhaseDuration calculates the duration for the current phase
func (es *EchobeatsSchedulerV2) calculatePhaseDuration() time.Duration {
	baseDuration := 5 * time.Minute

	switch es.rhythmPhase {
	case RhythmFocus:
		// Focus phase is longer when urgency is high
		return time.Duration(float64(baseDuration) * (1.0 + es.urgencyLevel))
	case RhythmExplore:
		// Explore phase is longer when focus intensity is low
		return time.Duration(float64(baseDuration) * (2.0 - es.focusIntensity))
	case RhythmIntegrate:
		// Integration phase is moderate
		return baseDuration
	case RhythmRest:
		// Rest phase is shorter when urgency is high
		return time.Duration(float64(baseDuration) * (1.0 - es.urgencyLevel*0.5))
	}
	return baseDuration
}

// processFocusPhase processes the focus phase
func (es *EchobeatsSchedulerV2) processFocusPhase() {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Select highest priority goal if none active
	if es.activeGoal == nil && len(es.goals) > 0 {
		es.selectNextGoal()
	}

	// Progress on active goal
	if es.activeGoal != nil {
		es.progressGoal()
	}
}

// processExplorePhase processes the explore phase
func (es *EchobeatsSchedulerV2) processExplorePhase() {
	// During explore phase, generate new goals based on interests
	es.mu.Lock()
	defer es.mu.Unlock()

	// Slow down beat rate during exploration
	es.adaptiveRate = 1.5
}

// processIntegratePhase processes the integrate phase
func (es *EchobeatsSchedulerV2) processIntegratePhase() {
	// During integrate phase, consolidate progress and learnings
	es.mu.Lock()
	defer es.mu.Unlock()

	// Normal beat rate during integration
	es.adaptiveRate = 1.0
}

// processRestPhase processes the rest phase
func (es *EchobeatsSchedulerV2) processRestPhase() {
	// During rest phase, minimal activity
	es.mu.Lock()
	defer es.mu.Unlock()

	// Slow down significantly during rest
	es.adaptiveRate = 3.0
}

// selectNextGoal selects the next goal to work on
func (es *EchobeatsSchedulerV2) selectNextGoal() {
	if len(es.goals) == 0 {
		return
	}

	// Find highest priority pending goal
	var bestGoal *ScheduledGoalV2
	var bestScore float64 = -1

	for i := range es.goals {
		goal := &es.goals[i]
		if goal.Status != GoalPendingV2 {
			continue
		}

		// Score based on priority and urgency
		score := goal.Priority
		if !goal.Deadline.IsZero() {
			timeUntilDeadline := time.Until(goal.Deadline)
			if timeUntilDeadline < 0 {
				score += 1.0 // Overdue goals get priority boost
			} else if timeUntilDeadline < goal.EstimatedTime*2 {
				score += 0.5 // Approaching deadline
			}
		}

		if score > bestScore {
			bestScore = score
			bestGoal = goal
		}
	}

	if bestGoal != nil {
		bestGoal.Status = GoalActiveV2
		bestGoal.StartedAt = time.Now()
		es.activeGoal = bestGoal
		fmt.Printf("🎯 Selected goal: %s (priority: %.2f)\n", bestGoal.Description, bestGoal.Priority)
	}
}

// progressGoal makes progress on the active goal
func (es *EchobeatsSchedulerV2) progressGoal() {
	if es.activeGoal == nil {
		return
	}

	// Simulate progress (in real implementation, this would involve actual work)
	progressIncrement := 0.1 * es.focusIntensity
	es.activeGoal.Progress = minFloat(1.0, es.activeGoal.Progress+progressIncrement)

	// Check if goal is complete
	if es.activeGoal.Progress >= 1.0 {
		es.completeGoal()
	}
}

// completeGoal marks the active goal as complete
func (es *EchobeatsSchedulerV2) completeGoal() {
	if es.activeGoal == nil {
		return
	}

	es.activeGoal.Status = GoalCompletedV2
	es.activeGoal.CompletedAt = time.Now()
	es.activeGoal.ActualTime = es.activeGoal.CompletedAt.Sub(es.activeGoal.StartedAt)

	es.completedGoals = append(es.completedGoals, *es.activeGoal)
	es.goalsCompleted++

	// Update average goal duration
	totalDuration := es.avgGoalDuration * time.Duration(es.goalsCompleted-1)
	totalDuration += es.activeGoal.ActualTime
	es.avgGoalDuration = totalDuration / time.Duration(es.goalsCompleted)

	fmt.Printf("✅ Goal completed: %s (duration: %v)\n", 
		es.activeGoal.Description, es.activeGoal.ActualTime)

	es.activeGoal = nil
}

// updateAdaptiveRate updates the adaptive rate based on cognitive state
func (es *EchobeatsSchedulerV2) updateAdaptiveRate() {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Adjust based on goal progress and urgency
	if es.activeGoal != nil && es.urgencyLevel > 0.7 {
		es.adaptiveRate = maxFloat(0.5, es.adaptiveRate*0.9) // Speed up
	} else if es.urgencyLevel < 0.3 {
		es.adaptiveRate = minFloat(3.0, es.adaptiveRate*1.1) // Slow down
	}
}

// AddGoal adds a new goal to the scheduler
func (es *EchobeatsSchedulerV2) AddGoal(description string, priority float64) string {
	es.mu.Lock()
	defer es.mu.Unlock()

	goal := ScheduledGoalV2{
		ID:          fmt.Sprintf("goal_%d", time.Now().UnixNano()),
		Description: description,
		Priority:    priority,
		Progress:    0.0,
		Status:      GoalPendingV2,
		CreatedAt:   time.Now(),
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}

	es.goals = append(es.goals, goal)
	return goal.ID
}

// SetUrgency sets the urgency level
func (es *EchobeatsSchedulerV2) SetUrgency(urgency float64) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.urgencyLevel = minFloat(1.0, maxFloat(0.0, urgency))
}

// SetFocusIntensity sets the focus intensity
func (es *EchobeatsSchedulerV2) SetFocusIntensity(intensity float64) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.focusIntensity = minFloat(1.0, maxFloat(0.0, intensity))
}

// GetMetrics returns scheduler metrics
func (es *EchobeatsSchedulerV2) GetMetrics() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return map[string]interface{}{
		"total_beats":       es.totalBeats,
		"goals_completed":   es.goalsCompleted,
		"goals_abandoned":   es.goalsAbandoned,
		"pending_goals":     len(es.goals),
		"rhythm_phase":      es.rhythmPhase.String(),
		"adaptive_rate":     es.adaptiveRate,
		"urgency_level":     es.urgencyLevel,
		"focus_intensity":   es.focusIntensity,
		"avg_goal_duration": es.avgGoalDuration.String(),
		"running":           es.running,
	}
}

// GetActiveGoal returns the currently active goal
func (es *EchobeatsSchedulerV2) GetActiveGoal() *ScheduledGoalV2 {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.activeGoal
}

// GetPendingGoals returns all pending goals
func (es *EchobeatsSchedulerV2) GetPendingGoals() []ScheduledGoalV2 {
	es.mu.RLock()
	defer es.mu.RUnlock()

	pending := make([]ScheduledGoalV2, 0)
	for _, goal := range es.goals {
		if goal.Status == GoalPendingV2 {
			pending = append(pending, goal)
		}
	}
	return pending
}
