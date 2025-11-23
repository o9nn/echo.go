package echobeats

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// PhaseManager orchestrates the 12-step cognitive loop
type PhaseManager struct {
	mu           sync.RWMutex
	stepDuration time.Duration
	currentStep  int
	running      bool
	paused       bool
	handlers     map[Term]map[Mode]StepHandler
	ticker       *time.Ticker
	stopChan     chan struct{}
	stepConfigs  []StepConfig
	metrics      CognitiveMetrics
	startTime    time.Time
}

// NewPhaseManager creates a new phase manager
func NewPhaseManager(stepDuration time.Duration) *PhaseManager {
	return &PhaseManager{
		stepDuration: stepDuration,
		handlers:     make(map[Term]map[Mode]StepHandler),
		stopChan:     make(chan struct{}),
		stepConfigs:  buildStepConfigs(),
		metrics: CognitiveMetrics{
			CurrentStep: 0,
		},
	}
}

// RegisterHandler registers a handler for a specific term and mode
func (pm *PhaseManager) RegisterHandler(term Term, mode Mode, handler StepHandler) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.handlers[term] == nil {
		pm.handlers[term] = make(map[Mode]StepHandler)
	}
	pm.handlers[term][mode] = handler
}

// Start begins the cognitive loop
func (pm *PhaseManager) Start() error {
	pm.mu.Lock()
	if pm.running {
		pm.mu.Unlock()
		return fmt.Errorf("phase manager already running")
	}
	pm.running = true
	pm.paused = false
	pm.startTime = time.Now()
	pm.mu.Unlock()

	log.Printf("🎵 PhaseManager: Starting 12-step cognitive loop (step duration: %v)\n", pm.stepDuration)

	pm.ticker = time.NewTicker(pm.stepDuration)
	go pm.runLoop()

	return nil
}

// Stop halts the cognitive loop
func (pm *PhaseManager) Stop() {
	pm.mu.Lock()
	if !pm.running {
		pm.mu.Unlock()
		return
	}
	pm.running = false
	pm.mu.Unlock()

	close(pm.stopChan)
	if pm.ticker != nil {
		pm.ticker.Stop()
	}

	log.Println("🎵 PhaseManager: Stopped")
}

// Pause temporarily pauses the cognitive loop
func (pm *PhaseManager) Pause() {
	pm.mu.Lock()
	pm.paused = true
	pm.mu.Unlock()
	log.Println("⏸️  PhaseManager: Paused")
}

// Resume resumes the cognitive loop
func (pm *PhaseManager) Resume() {
	pm.mu.Lock()
	pm.paused = false
	pm.mu.Unlock()
	log.Println("▶️  PhaseManager: Resumed")
}

// runLoop executes the main cognitive loop
func (pm *PhaseManager) runLoop() {
	for {
		select {
		case <-pm.stopChan:
			return

		case <-pm.ticker.C:
			pm.mu.RLock()
			paused := pm.paused
			pm.mu.RUnlock()

			if paused {
				continue
			}

			pm.executeStep()
		}
	}
}

// executeStep executes the current cognitive step
func (pm *PhaseManager) executeStep() {
	stepStart := time.Now()

	pm.mu.Lock()
	step := pm.currentStep
	pm.currentStep = (pm.currentStep + 1) % 12
	pm.metrics.TotalSteps++
	pm.mu.Unlock()

	// Get step configuration
	config := pm.stepConfigs[step]

	// Update metrics
	pm.mu.Lock()
	if config.Mode == Expressive {
		pm.metrics.ExpressiveSteps++
	} else {
		pm.metrics.ReflectiveSteps++
	}
	pm.metrics.CurrentStep = step
	pm.mu.Unlock()

	// Execute handler if registered
	pm.mu.RLock()
	handlers, hasTermHandlers := pm.handlers[config.Term]
	pm.mu.RUnlock()

	if hasTermHandlers {
		pm.mu.RLock()
		handler, hasHandler := handlers[config.Mode]
		pm.mu.RUnlock()

		if hasHandler {
			if err := handler(step, config.Mode); err != nil {
				log.Printf("⚠️  Step %d handler error: %v\n", step, err)
			}
		}
	}

	// Update timing metrics
	stepDuration := time.Since(stepStart)
	pm.mu.Lock()
	pm.metrics.LastStepTime = time.Now()
	if pm.metrics.AverageStepTime == 0 {
		pm.metrics.AverageStepTime = stepDuration
	} else {
		// Exponential moving average
		pm.metrics.AverageStepTime = time.Duration(
			0.9*float64(pm.metrics.AverageStepTime) + 0.1*float64(stepDuration),
		)
	}
	pm.mu.Unlock()
}

// GetCurrentStep returns the current step number
func (pm *PhaseManager) GetCurrentStep() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.currentStep
}

// GetMetrics returns current phase manager metrics
func (pm *PhaseManager) GetMetrics() CognitiveMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.metrics
}

// IsRunning returns whether the phase manager is running
func (pm *PhaseManager) IsRunning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.running
}

// IsPaused returns whether the phase manager is paused
func (pm *PhaseManager) IsPaused() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.paused
}

// GetStepConfig returns the configuration for a specific step
func (pm *PhaseManager) GetStepConfig(step int) StepConfig {
	if step < 0 || step >= len(pm.stepConfigs) {
		return StepConfig{}
	}
	return pm.stepConfigs[step]
}

// GetUptime returns how long the phase manager has been running
func (pm *PhaseManager) GetUptime() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if !pm.running {
		return 0
	}
	return time.Since(pm.startTime)
}
