package echobeats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
)

// GoCronCycleTimer provides precise, configurable timing for the 12-step Echobeats
// cognitive cycle using the gocron scheduler. It replaces ad-hoc time.Ticker usage
// with a proper scheduling system that supports:
//   - Configurable cycle intervals (default 100ms per step = 1.2s per full cycle)
//   - Wake/rest cycle scheduling with fatigue-based transitions
//   - Dream cycle scheduling during rest periods
//   - Goal pursuit scheduling with priority-based timing
//   - Metrics collection on a separate schedule
type GoCronCycleTimer struct {
	mu        sync.RWMutex
	scheduler gocron.Scheduler

	// Cycle configuration
	stepInterval    time.Duration // Time between each of the 12 steps
	dreamInterval   time.Duration // Time between dream cycle checks
	metricsInterval time.Duration // Time between metrics snapshots
	goalInterval    time.Duration // Time between goal pursuit ticks

	// Callbacks
	onBeatStep     func(stepNumber int)       // Called for each of the 12 beat steps
	onCycleComplete func(cycleNumber uint64)   // Called when a full 12-step cycle completes
	onDreamCheck   func()                      // Called to check if dreaming should start/continue
	onGoalTick     func()                      // Called to advance goal pursuit
	onMetrics      func()                      // Called to collect metrics

	// State
	currentStep   int
	cycleCount    uint64
	running       bool

	// Job references for dynamic control
	beatJob       gocron.Job
	dreamJob      gocron.Job
	goalJob       gocron.Job
	metricsJob    gocron.Job
}

// CycleTimerConfig holds configuration for the GoCronCycleTimer
type CycleTimerConfig struct {
	StepInterval    time.Duration // Default: 100ms
	DreamInterval   time.Duration // Default: 5s
	MetricsInterval time.Duration // Default: 10s
	GoalInterval    time.Duration // Default: 2s
}

// DefaultCycleTimerConfig returns sensible defaults for cognitive cycle timing
func DefaultCycleTimerConfig() CycleTimerConfig {
	return CycleTimerConfig{
		StepInterval:    100 * time.Millisecond,
		DreamInterval:   5 * time.Second,
		MetricsInterval: 10 * time.Second,
		GoalInterval:    2 * time.Second,
	}
}

// NewGoCronCycleTimer creates a new gocron-based cycle timer
func NewGoCronCycleTimer(config CycleTimerConfig) (*GoCronCycleTimer, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create gocron scheduler: %w", err)
	}

	return &GoCronCycleTimer{
		scheduler:       s,
		stepInterval:    config.StepInterval,
		dreamInterval:   config.DreamInterval,
		metricsInterval: config.MetricsInterval,
		goalInterval:    config.GoalInterval,
		currentStep:     0,
		cycleCount:      0,
	}, nil
}

// SetBeatStepCallback sets the function called for each beat step (1-12)
func (t *GoCronCycleTimer) SetBeatStepCallback(fn func(stepNumber int)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onBeatStep = fn
}

// SetCycleCompleteCallback sets the function called when a full 12-step cycle completes
func (t *GoCronCycleTimer) SetCycleCompleteCallback(fn func(cycleNumber uint64)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onCycleComplete = fn
}

// SetDreamCheckCallback sets the function called periodically to manage dream cycles
func (t *GoCronCycleTimer) SetDreamCheckCallback(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onDreamCheck = fn
}

// SetGoalTickCallback sets the function called periodically to advance goal pursuit
func (t *GoCronCycleTimer) SetGoalTickCallback(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onGoalTick = fn
}

// SetMetricsCallback sets the function called periodically to collect metrics
func (t *GoCronCycleTimer) SetMetricsCallback(fn func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onMetrics = fn
}

// Start begins the cognitive cycle timer with all configured schedules
func (t *GoCronCycleTimer) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("cycle timer already running")
	}

	// Schedule the 12-step beat cycle
	beatJob, err := t.scheduler.NewJob(
		gocron.DurationJob(t.stepInterval),
		gocron.NewTask(func() {
			t.executeBeatStep()
		}),
		gocron.WithName("echobeats-12step"),
	)
	if err != nil {
		return fmt.Errorf("failed to schedule beat step job: %w", err)
	}
	t.beatJob = beatJob

	// Schedule dream cycle checks
	if t.onDreamCheck != nil {
		dreamJob, err := t.scheduler.NewJob(
			gocron.DurationJob(t.dreamInterval),
			gocron.NewTask(func() {
				t.mu.RLock()
				fn := t.onDreamCheck
				t.mu.RUnlock()
				if fn != nil {
					fn()
				}
			}),
			gocron.WithName("echodream-check"),
		)
		if err != nil {
			return fmt.Errorf("failed to schedule dream check job: %w", err)
		}
		t.dreamJob = dreamJob
	}

	// Schedule goal pursuit ticks
	if t.onGoalTick != nil {
		goalJob, err := t.scheduler.NewJob(
			gocron.DurationJob(t.goalInterval),
			gocron.NewTask(func() {
				t.mu.RLock()
				fn := t.onGoalTick
				t.mu.RUnlock()
				if fn != nil {
					fn()
				}
			}),
			gocron.WithName("goal-pursuit-tick"),
		)
		if err != nil {
			return fmt.Errorf("failed to schedule goal tick job: %w", err)
		}
		t.goalJob = goalJob
	}

	// Schedule metrics collection
	if t.onMetrics != nil {
		metricsJob, err := t.scheduler.NewJob(
			gocron.DurationJob(t.metricsInterval),
			gocron.NewTask(func() {
				t.mu.RLock()
				fn := t.onMetrics
				t.mu.RUnlock()
				if fn != nil {
					fn()
				}
			}),
			gocron.WithName("metrics-collection"),
		)
		if err != nil {
			return fmt.Errorf("failed to schedule metrics job: %w", err)
		}
		t.metricsJob = metricsJob
	}

	// Start the scheduler
	t.scheduler.Start()
	t.running = true

	// Monitor context for shutdown
	go func() {
		<-ctx.Done()
		t.Stop()
	}()

	return nil
}

// executeBeatStep advances the 12-step cognitive cycle by one step
func (t *GoCronCycleTimer) executeBeatStep() {
	t.mu.Lock()
	t.currentStep++
	step := t.currentStep
	if t.currentStep > 12 {
		t.currentStep = 1
		step = 1
		t.cycleCount++
	}
	cycleNum := t.cycleCount
	onStep := t.onBeatStep
	onCycle := t.onCycleComplete
	t.mu.Unlock()

	// Execute the beat step callback
	if onStep != nil {
		onStep(step)
	}

	// If we just completed step 12, fire the cycle complete callback
	if step == 12 && onCycle != nil {
		onCycle(cycleNum)
	}
}

// Stop halts all scheduled cognitive cycles
func (t *GoCronCycleTimer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return
	}

	if err := t.scheduler.Shutdown(); err != nil {
		fmt.Printf("⚠️  GoCronCycleTimer: shutdown error: %v\n", err)
	}
	t.running = false
}

// AdjustStepInterval dynamically changes the beat step interval
// This enables fatigue-based slowdown: as fatigue increases, steps take longer
func (t *GoCronCycleTimer) AdjustStepInterval(newInterval time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		t.stepInterval = newInterval
		return nil
	}

	// Remove old beat job and create new one with updated interval
	if err := t.scheduler.RemoveJob(t.beatJob.ID()); err != nil {
		return fmt.Errorf("failed to remove old beat job: %w", err)
	}

	beatJob, err := t.scheduler.NewJob(
		gocron.DurationJob(newInterval),
		gocron.NewTask(func() {
			t.executeBeatStep()
		}),
		gocron.WithName("echobeats-12step"),
	)
	if err != nil {
		return fmt.Errorf("failed to reschedule beat step job: %w", err)
	}

	t.beatJob = beatJob
	t.stepInterval = newInterval
	return nil
}

// GetState returns the current timer state
func (t *GoCronCycleTimer) GetState() CycleTimerState {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return CycleTimerState{
		CurrentStep:   t.currentStep,
		CycleCount:    t.cycleCount,
		StepInterval:  t.stepInterval,
		Running:       t.running,
	}
}

// CycleTimerState represents the observable state of the cycle timer
type CycleTimerState struct {
	CurrentStep   int           `json:"current_step"`
	CycleCount    uint64        `json:"cycle_count"`
	StepInterval  time.Duration `json:"step_interval"`
	Running       bool          `json:"running"`
}

// StreamForStep returns which cognitive stream is active for a given step number
// Streams are phased 4 steps apart: Perception={1,5,9}, Action={2,6,10}, Simulation={3,7,11}
func StreamForStep(step int) string {
	switch step % 4 {
	case 1:
		return "perception"
	case 2:
		return "action"
	case 3:
		return "simulation"
	case 0:
		return "integration" // Steps 4, 8, 12
	default:
		return "unknown"
	}
}

// PhaseForStep returns which cognitive phase a step belongs to
func PhaseForStep(step int) string {
	switch {
	case step <= 3:
		return "sense"
	case step <= 6:
		return "process"
	case step <= 9:
		return "emit"
	case step <= 12:
		return "integrate"
	default:
		return "unknown"
	}
}
