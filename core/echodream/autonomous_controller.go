package echodream

import (
	"sync"
	"time"
)

// AutonomousController manages autonomous wake/rest cycles
type AutonomousController struct {
	mu               sync.RWMutex
	maxAwakeDuration time.Duration
	minRestDuration  time.Duration
	fatigueThreshold float64
	restThreshold    float64
	currentFatigue   float64
	awakeStartTime   time.Time
	restStartTime    time.Time
	isAwake          bool
}

// NewAutonomousController creates a new autonomous controller
func NewAutonomousController(
	maxAwakeDuration time.Duration,
	minRestDuration time.Duration,
	fatigueThreshold float64,
	restThreshold float64,
) *AutonomousController {
	return &AutonomousController{
		maxAwakeDuration: maxAwakeDuration,
		minRestDuration:  minRestDuration,
		fatigueThreshold: fatigueThreshold,
		restThreshold:    restThreshold,
		currentFatigue:   0.0,
		awakeStartTime:   time.Now(),
		isAwake:          true,
	}
}

// GetFatigue returns current fatigue level
func (ac *AutonomousController) GetFatigue() float64 {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.currentFatigue
}

// UpdateFatigue updates fatigue level
func (ac *AutonomousController) UpdateFatigue(delta float64) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.currentFatigue += delta
	if ac.currentFatigue > 1.0 {
		ac.currentFatigue = 1.0
	}
	if ac.currentFatigue < 0.0 {
		ac.currentFatigue = 0.0
	}
}

// ResetFatigue resets fatigue to zero
func (ac *AutonomousController) ResetFatigue() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.currentFatigue = 0.0
	ac.awakeStartTime = time.Now()
	ac.isAwake = true
}

// GetRestDuration returns how long system has been resting
func (ac *AutonomousController) GetRestDuration() time.Duration {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if ac.isAwake {
		return 0
	}
	return time.Since(ac.restStartTime)
}

// GetAwakeDuration returns how long system has been awake
func (ac *AutonomousController) GetAwakeDuration() time.Duration {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if !ac.isAwake {
		return 0
	}
	return time.Since(ac.awakeStartTime)
}

// StartRest begins a rest cycle
func (ac *AutonomousController) StartRest() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.isAwake = false
	ac.restStartTime = time.Now()
}

// StartWake begins a wake cycle
func (ac *AutonomousController) StartWake() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.isAwake = true
	ac.awakeStartTime = time.Now()
}

// ShouldWake determines if system should wake up
func (ac *AutonomousController) ShouldWake() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if ac.isAwake {
		return false
	}

	restDuration := time.Since(ac.restStartTime)
	return restDuration >= ac.minRestDuration && ac.currentFatigue < ac.restThreshold
}

// ShouldRest determines if system should rest
func (ac *AutonomousController) ShouldRest() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if !ac.isAwake {
		return false
	}

	awakeDuration := time.Since(ac.awakeStartTime)
	return awakeDuration >= ac.maxAwakeDuration || ac.currentFatigue >= ac.fatigueThreshold
}

// IsAwake returns whether system is currently awake
func (ac *AutonomousController) IsAwake() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.isAwake
}

// GetState returns a summary of the current state
func (ac *AutonomousController) GetState() map[string]interface{} {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	state := map[string]interface{}{
		"is_awake":        ac.isAwake,
		"fatigue":         ac.currentFatigue,
		"should_rest":     false,
		"should_wake":     false,
	}

	if ac.isAwake {
		state["awake_duration"] = time.Since(ac.awakeStartTime).String()
		awakeDuration := time.Since(ac.awakeStartTime)
		state["should_rest"] = awakeDuration >= ac.maxAwakeDuration || ac.currentFatigue >= ac.fatigueThreshold
	} else {
		state["rest_duration"] = time.Since(ac.restStartTime).String()
		restDuration := time.Since(ac.restStartTime)
		state["should_wake"] = restDuration >= ac.minRestDuration && ac.currentFatigue < ac.restThreshold
	}

	return state
}
