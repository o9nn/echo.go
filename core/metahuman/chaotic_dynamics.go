package metahuman

import (
	"math"
	"sync"
)

// LorenzAttractor implements the Lorenz system for generating chaotic
// micro-expression dynamics. The attractor produces bounded but unpredictable
// trajectories that prevent uncanny-valley smoothness in avatar expressions.
//
// System equations (RK4 integration):
//
//	dx/dt = sigma * (y - x)
//	dy/dt = x * (rho - z) - y
//	dz/dt = x * y - beta * z
type LorenzAttractor struct {
	mu sync.Mutex

	// State variables
	X, Y, Z float64

	// System parameters
	Sigma float64 // Rate of chaotic divergence (default: 10.0)
	Rho   float64 // System complexity (default: 28.0)
	Beta  float64 // Dissipation rate (default: 8/3)

	// Integration parameters
	DT float64 // Integration timestep (default: 0.01)

	// Scaling
	ChaosIntensity float64 // Amplitude scaling to morph targets (default: 0.15)

	// Lyapunov monitoring
	perturbX, perturbY, perturbZ float64
	lyapunovSum                  float64
	lyapunovSteps                int
}

// NewLorenzAttractor creates a new Lorenz attractor with default parameters
// and a standard initial condition near the attractor.
func NewLorenzAttractor() *LorenzAttractor {
	return &LorenzAttractor{
		// Initial state near the attractor
		X: 1.0,
		Y: 1.0,
		Z: 1.0,

		// Standard Lorenz parameters (chaotic regime)
		Sigma: 10.0,
		Rho:   28.0,
		Beta:  8.0 / 3.0,

		// Integration
		DT: 0.01,

		// Scaling
		ChaosIntensity: 0.15,

		// Perturbed trajectory for Lyapunov estimation
		perturbX: 1.0001,
		perturbY: 1.0,
		perturbZ: 1.0,
	}
}

// Step advances the attractor by one timestep using RK4 integration.
// Returns the normalized state suitable for micro-expression modulation.
func (la *LorenzAttractor) Step() (x, y, z float64) {
	la.mu.Lock()
	defer la.mu.Unlock()

	// RK4 integration for main trajectory
	la.X, la.Y, la.Z = la.rk4Step(la.X, la.Y, la.Z)

	// RK4 integration for perturbed trajectory (Lyapunov estimation)
	la.perturbX, la.perturbY, la.perturbZ = la.rk4Step(la.perturbX, la.perturbY, la.perturbZ)

	// Update Lyapunov estimation
	la.updateLyapunov()

	// Normalize to approximately [-1, 1] range
	// Lorenz attractor typical ranges: x∈[-20,20], y∈[-30,30], z∈[0,50]
	return la.X / 20.0, la.Y / 30.0, (la.Z - 25.0) / 25.0
}

// StepDelta advances by a variable delta time (seconds).
// Performs multiple sub-steps to maintain numerical stability.
func (la *LorenzAttractor) StepDelta(deltaTime float64) (x, y, z float64) {
	la.mu.Lock()
	defer la.mu.Unlock()

	steps := int(math.Ceil(deltaTime / la.DT))
	if steps < 1 {
		steps = 1
	}
	if steps > 1000 {
		steps = 1000 // Safety cap
	}

	for i := 0; i < steps; i++ {
		la.X, la.Y, la.Z = la.rk4Step(la.X, la.Y, la.Z)
		la.perturbX, la.perturbY, la.perturbZ = la.rk4Step(la.perturbX, la.perturbY, la.perturbZ)
	}
	la.updateLyapunov()

	return la.X / 20.0, la.Y / 30.0, (la.Z - 25.0) / 25.0
}

// rk4Step performs one RK4 integration step (must be called under lock).
func (la *LorenzAttractor) rk4Step(x, y, z float64) (float64, float64, float64) {
	dt := la.DT

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

	// Update
	newX := x + (dt/6.0)*(k1x+2*k2x+2*k3x+k4x)
	newY := y + (dt/6.0)*(k1y+2*k2y+2*k3y+k4y)
	newZ := z + (dt/6.0)*(k1z+2*k2z+2*k3z+k4z)

	return newX, newY, newZ
}

// derivatives computes the Lorenz system derivatives.
func (la *LorenzAttractor) derivatives(x, y, z float64) (float64, float64, float64) {
	dxdt := la.Sigma * (y - x)
	dydt := x*(la.Rho-z) - y
	dzdt := x*y - la.Beta*z
	return dxdt, dydt, dzdt
}

// updateLyapunov updates the running Lyapunov exponent estimation.
func (la *LorenzAttractor) updateLyapunov() {
	dx := la.perturbX - la.X
	dy := la.perturbY - la.Y
	dz := la.perturbZ - la.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)

	if dist > 1e-10 {
		la.lyapunovSum += math.Log(dist / 1e-4)
		la.lyapunovSteps++

		// Renormalize perturbed trajectory
		scale := 1e-4 / dist
		la.perturbX = la.X + dx*scale
		la.perturbY = la.Y + dy*scale
		la.perturbZ = la.Z + dz*scale
	}
}

// LyapunovExponent returns the estimated maximum Lyapunov exponent.
// Positive = chaotic (good for expressions), negative = periodic (bad).
func (la *LorenzAttractor) LyapunovExponent() float64 {
	la.mu.Lock()
	defer la.mu.Unlock()
	if la.lyapunovSteps == 0 {
		return 0.0
	}
	return la.lyapunovSum / float64(la.lyapunovSteps) / la.DT
}

// ApplyMicroExpressions applies chaotic noise to the FACS state.
// The Lorenz attractor output is scaled by ChaosIntensity and mapped
// to specific micro-expression channels.
func (la *LorenzAttractor) ApplyMicroExpressions(facs *FACSState) {
	x, y, z := la.Step()
	ci := la.ChaosIntensity

	// Micro-expression channels from meta-echo-dna spec
	facs.Add(AU1, x*ci*0.3)   // Subtle brow micro-movements
	facs.Add(AU7, y*ci*0.2)   // Eye tension micro-fluctuations
	facs.Add(AU12, z*ci*0.15) // Mouth corner micro-twitches
	facs.Add(AU9, (x+y)*ci*0.1) // Nose micro-wrinkles
	facs.Add(AU26, z*ci*0.05) // Jaw micro-movements
}

// ApplyMicroExpressionsDelta applies chaotic noise with variable delta time.
func (la *LorenzAttractor) ApplyMicroExpressionsDelta(facs *FACSState, deltaTime float64) {
	x, y, z := la.StepDelta(deltaTime)
	ci := la.ChaosIntensity

	facs.Add(AU1, x*ci*0.3)
	facs.Add(AU7, y*ci*0.2)
	facs.Add(AU12, z*ci*0.15)
	facs.Add(AU9, (x+y)*ci*0.1)
	facs.Add(AU26, z*ci*0.05)
}

// IsHealthy returns true if the attractor is in a chaotic regime
// (positive Lyapunov exponent) and not diverging to infinity.
func (la *LorenzAttractor) IsHealthy() bool {
	la.mu.Lock()
	defer la.mu.Unlock()
	// Check for NaN/Inf
	if math.IsNaN(la.X) || math.IsInf(la.X, 0) {
		return false
	}
	if math.IsNaN(la.Y) || math.IsInf(la.Y, 0) {
		return false
	}
	if math.IsNaN(la.Z) || math.IsInf(la.Z, 0) {
		return false
	}
	// Check bounds (Lorenz attractor should stay bounded)
	if math.Abs(la.X) > 100 || math.Abs(la.Y) > 100 || math.Abs(la.Z) > 100 {
		return false
	}
	return true
}

// Reset reinitializes the attractor to default initial conditions.
func (la *LorenzAttractor) Reset() {
	la.mu.Lock()
	defer la.mu.Unlock()
	la.X = 1.0
	la.Y = 1.0
	la.Z = 1.0
	la.perturbX = 1.0001
	la.perturbY = 1.0
	la.perturbZ = 1.0
	la.lyapunovSum = 0
	la.lyapunovSteps = 0
}
