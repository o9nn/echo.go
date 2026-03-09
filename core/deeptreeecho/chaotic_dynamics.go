package deeptreeecho

import (
	"math"
	"sync"
	"time"
)

// LorenzAttractor implements the Lorenz system for generating chaotic dynamics
// Used for micro-expression noise, cognitive perturbation, and preventing
// uncanny-valley smoothness in avatar expressions and thought patterns.
//
// The Lorenz system:
//   dx/dt = sigma * (y - x)
//   dy/dt = x * (rho - z) - y
//   dz/dt = x * y - beta * z
//
// With default parameters sigma=10, rho=28, beta=8/3, the system exhibits
// chaotic behavior with a positive Lyapunov exponent.
type LorenzAttractor struct {
	mu sync.RWMutex

	// State variables
	X, Y, Z float64

	// System parameters
	Sigma float64 // Prandtl number (rate of convective overturning)
	Rho   float64 // Rayleigh number (temperature difference driving)
	Beta  float64 // Geometric factor

	// Integration parameters
	dt float64 // Time step for RK4 integration

	// Scaling
	ChaosIntensity float64 // How much chaos to inject [0, 1]
	OutputScale    float64 // Scale factor for output

	// Monitoring
	lyapunovSum     float64
	lyapunovCount   uint64
	lastLyapunov    float64
	perturbedState  [3]float64 // For Lyapunov computation
	perturbEpsilon  float64

	// History
	stateHistory []LorenzState
	maxHistory   int

	// Metrics
	totalSteps uint64
	startTime  time.Time
}

// LorenzState captures a point-in-time state of the attractor
type LorenzState struct {
	X, Y, Z   float64
	Lyapunov  float64
	Timestamp time.Time
}

// NewLorenzAttractor creates a new Lorenz attractor with default chaotic parameters
func NewLorenzAttractor() *LorenzAttractor {
	la := &LorenzAttractor{
		// Classic Lorenz parameters for chaotic regime
		Sigma: 10.0,
		Rho:   28.0,
		Beta:  8.0 / 3.0,

		// Initial state (slightly off the fixed point)
		X: 1.0,
		Y: 1.0,
		Z: 1.0,

		// Integration
		dt: 0.01,

		// Scaling
		ChaosIntensity: 0.15,
		OutputScale:    1.0 / 50.0, // Normalize Lorenz output to ~[-1, 1]

		// Lyapunov monitoring
		perturbEpsilon: 1e-8,

		// History
		stateHistory: make([]LorenzState, 0, 1000),
		maxHistory:   1000,

		startTime: time.Now(),
	}

	// Initialize perturbed state for Lyapunov computation
	la.perturbedState = [3]float64{
		la.X + la.perturbEpsilon,
		la.Y,
		la.Z,
	}

	return la
}

// Step advances the Lorenz system by one time step using RK4 integration
func (la *LorenzAttractor) Step() (x, y, z float64) {
	la.mu.Lock()
	defer la.mu.Unlock()

	// RK4 integration for main state
	la.X, la.Y, la.Z = la.rk4Step(la.X, la.Y, la.Z)

	// RK4 integration for perturbed state (Lyapunov computation)
	la.perturbedState[0], la.perturbedState[1], la.perturbedState[2] =
		la.rk4Step(la.perturbedState[0], la.perturbedState[1], la.perturbedState[2])

	// Compute Lyapunov exponent contribution
	la.updateLyapunov()

	la.totalSteps++

	// Record history periodically
	if la.totalSteps%10 == 0 {
		la.stateHistory = append(la.stateHistory, LorenzState{
			X: la.X, Y: la.Y, Z: la.Z,
			Lyapunov:  la.lastLyapunov,
			Timestamp: time.Now(),
		})
		if len(la.stateHistory) > la.maxHistory {
			la.stateHistory = la.stateHistory[1:]
		}
	}

	return la.X, la.Y, la.Z
}

// rk4Step performs one RK4 integration step
func (la *LorenzAttractor) rk4Step(x, y, z float64) (float64, float64, float64) {
	dt := la.dt

	// k1
	k1x, k1y, k1z := la.derivatives(x, y, z)

	// k2
	k2x, k2y, k2z := la.derivatives(
		x+0.5*dt*k1x,
		y+0.5*dt*k1y,
		z+0.5*dt*k1z,
	)

	// k3
	k3x, k3y, k3z := la.derivatives(
		x+0.5*dt*k2x,
		y+0.5*dt*k2y,
		z+0.5*dt*k2z,
	)

	// k4
	k4x, k4y, k4z := la.derivatives(
		x+dt*k3x,
		y+dt*k3y,
		z+dt*k3z,
	)

	// Combine
	newX := x + (dt/6.0)*(k1x+2*k2x+2*k3x+k4x)
	newY := y + (dt/6.0)*(k1y+2*k2y+2*k3y+k4y)
	newZ := z + (dt/6.0)*(k1z+2*k2z+2*k3z+k4z)

	return newX, newY, newZ
}

// derivatives computes the Lorenz system derivatives
func (la *LorenzAttractor) derivatives(x, y, z float64) (dx, dy, dz float64) {
	dx = la.Sigma * (y - x)
	dy = x*(la.Rho-z) - y
	dz = x*y - la.Beta*z
	return
}

// updateLyapunov computes the running Lyapunov exponent estimate
func (la *LorenzAttractor) updateLyapunov() {
	// Distance between main and perturbed trajectories
	dx := la.perturbedState[0] - la.X
	dy := la.perturbedState[1] - la.Y
	dz := la.perturbedState[2] - la.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

	if dist > 0 && la.perturbEpsilon > 0 {
		// Lyapunov exponent contribution
		lambda := math.Log(dist / la.perturbEpsilon)
		la.lyapunovSum += lambda
		la.lyapunovCount++
		la.lastLyapunov = la.lyapunovSum / float64(la.lyapunovCount)

		// Renormalize perturbed state
		scale := la.perturbEpsilon / dist
		la.perturbedState[0] = la.X + dx*scale
		la.perturbedState[1] = la.Y + dy*scale
		la.perturbedState[2] = la.Z + dz*scale
	}
}

// GetScaledOutput returns the chaos-scaled output suitable for modulating other systems
// Returns three values in approximately [-ChaosIntensity, +ChaosIntensity]
func (la *LorenzAttractor) GetScaledOutput() (cx, cy, cz float64) {
	la.mu.RLock()
	defer la.mu.RUnlock()

	cx = la.X * la.OutputScale * la.ChaosIntensity
	cy = la.Y * la.OutputScale * la.ChaosIntensity
	cz = la.Z * la.OutputScale * la.ChaosIntensity

	return
}

// GetLyapunovExponent returns the estimated maximum Lyapunov exponent
// Positive = chaotic (desired), Negative = periodic (undesired)
func (la *LorenzAttractor) GetLyapunovExponent() float64 {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.lastLyapunov
}

// IsChaotic returns true if the system is in a chaotic regime
func (la *LorenzAttractor) IsChaotic() bool {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.lastLyapunov > 0 && la.lyapunovCount > 100
}

// SetParameters updates the Lorenz parameters
func (la *LorenzAttractor) SetParameters(sigma, rho, beta float64) {
	la.mu.Lock()
	defer la.mu.Unlock()
	la.Sigma = sigma
	la.Rho = rho
	la.Beta = beta
}

// SetChaosIntensity sets the chaos injection intensity
func (la *LorenzAttractor) SetChaosIntensity(intensity float64) {
	la.mu.Lock()
	defer la.mu.Unlock()
	la.ChaosIntensity = clampF64(intensity, 0.0, 1.0)
}

// GetState returns the current attractor state
func (la *LorenzAttractor) GetState() (x, y, z float64) {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.X, la.Y, la.Z
}

// GetMetrics returns attractor metrics
func (la *LorenzAttractor) GetMetrics() map[string]interface{} {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return map[string]interface{}{
		"x":               la.X,
		"y":               la.Y,
		"z":               la.Z,
		"sigma":           la.Sigma,
		"rho":             la.Rho,
		"beta":            la.Beta,
		"lyapunov":        la.lastLyapunov,
		"is_chaotic":      la.lastLyapunov > 0 && la.lyapunovCount > 100,
		"total_steps":     la.totalSteps,
		"chaos_intensity": la.ChaosIntensity,
		"uptime":          time.Since(la.startTime).String(),
	}
}

// GetHistory returns recent attractor states
func (la *LorenzAttractor) GetHistory(n int) []LorenzState {
	la.mu.RLock()
	defer la.mu.RUnlock()
	if n > len(la.stateHistory) {
		n = len(la.stateHistory)
	}
	start := len(la.stateHistory) - n
	result := make([]LorenzState, n)
	copy(result, la.stateHistory[start:])
	return result
}

// CognitiveNoiseGenerator uses the Lorenz attractor to generate structured noise
// for cognitive processes, preventing deterministic loops and enabling creative exploration
type CognitiveNoiseGenerator struct {
	mu        sync.RWMutex
	attractor *LorenzAttractor

	// Noise channels for different cognitive subsystems
	attentionNoise  float64 // Modulates attention focus
	memoryNoise     float64 // Modulates memory retrieval
	decisionNoise   float64 // Modulates decision thresholds
	expressionNoise float64 // Modulates avatar expression

	// Filtering
	smoothingFactor float64 // Exponential smoothing [0, 1]
}

// NewCognitiveNoiseGenerator creates a noise generator backed by a Lorenz attractor
func NewCognitiveNoiseGenerator() *CognitiveNoiseGenerator {
	return &CognitiveNoiseGenerator{
		attractor:       NewLorenzAttractor(),
		smoothingFactor: 0.3,
	}
}

// Update steps the attractor and updates all noise channels
func (cng *CognitiveNoiseGenerator) Update() {
	cng.mu.Lock()
	defer cng.mu.Unlock()

	cx, cy, cz := cng.attractor.Step()
	scale := cng.attractor.OutputScale * cng.attractor.ChaosIntensity
	sf := cng.smoothingFactor

	// Map Lorenz dimensions to cognitive noise channels with smoothing
	cng.attentionNoise = sf*cng.attentionNoise + (1-sf)*cx*scale
	cng.memoryNoise = sf*cng.memoryNoise + (1-sf)*cy*scale
	cng.decisionNoise = sf*cng.decisionNoise + (1-sf)*cz*scale
	// Expression noise is a combination
	cng.expressionNoise = sf*cng.expressionNoise + (1-sf)*(cx+cy)*0.5*scale
}

// GetAttentionNoise returns the current attention modulation noise
func (cng *CognitiveNoiseGenerator) GetAttentionNoise() float64 {
	cng.mu.RLock()
	defer cng.mu.RUnlock()
	return cng.attentionNoise
}

// GetMemoryNoise returns the current memory retrieval modulation noise
func (cng *CognitiveNoiseGenerator) GetMemoryNoise() float64 {
	cng.mu.RLock()
	defer cng.mu.RUnlock()
	return cng.memoryNoise
}

// GetDecisionNoise returns the current decision threshold modulation noise
func (cng *CognitiveNoiseGenerator) GetDecisionNoise() float64 {
	cng.mu.RLock()
	defer cng.mu.RUnlock()
	return cng.decisionNoise
}

// GetExpressionNoise returns the current expression modulation noise
func (cng *CognitiveNoiseGenerator) GetExpressionNoise() float64 {
	cng.mu.RLock()
	defer cng.mu.RUnlock()
	return cng.expressionNoise
}

// SetIntensity sets the chaos intensity for the underlying attractor
func (cng *CognitiveNoiseGenerator) SetIntensity(intensity float64) {
	cng.attractor.SetChaosIntensity(intensity)
}

// GetAttractor returns the underlying Lorenz attractor for direct access
func (cng *CognitiveNoiseGenerator) GetAttractor() *LorenzAttractor {
	return cng.attractor
}
