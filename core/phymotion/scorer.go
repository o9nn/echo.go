package phymotion

import (
	"math"
	"sort"
	"time"
)

// Vector3 is a small dependency-free spatial vector for embodied cognition scoring.
type Vector3 struct {
	X float64
	Y float64
	Z float64
}

// TrajectorySample records one moment in an embodied movement attempt.
type TrajectorySample struct {
	Time                  time.Time
	CenterOfMass          Vector3
	LeftFoot              Vector3
	RightFoot             Vector3
	SupportRadius         float64
	GroundReactionForce   float64
	NormalizedJointTorque float64
	MetabolicEffort       float64
}

// ScoreBreakdown exposes the three PhyMotion-inspired axes used by Echo.
type ScoreBreakdown struct {
	KinematicContinuity float64
	ContactBalance      float64
	DynamicPlausibility float64
}

// Score summarizes whether a movement is physically feasible enough to practice.
type Score struct {
	Overall   float64
	Breakdown ScoreBreakdown
	Problems  []string
}

// ScorerConfig controls physical feasibility thresholds.
type ScorerConfig struct {
	MaxSpeed             float64
	MaxAcceleration      float64
	MaxGroundReaction    float64
	MaxNormalizedTorque  float64
	MaxMetabolicEffort   float64
	FootContactTolerance float64
	MinimumSupportRadius float64
}

// DefaultScorerConfig returns conservative, dimensionless defaults suitable for
// comparative scoring of avatar or robotics trajectories.
func DefaultScorerConfig() ScorerConfig {
	return ScorerConfig{
		MaxSpeed:             6.0,
		MaxAcceleration:      30.0,
		MaxGroundReaction:    3.0,
		MaxNormalizedTorque:  1.0,
		MaxMetabolicEffort:   1.0,
		FootContactTolerance: 0.08,
		MinimumSupportRadius: 0.10,
	}
}

// Scorer evaluates movement trajectories for grounded embodied cognition.
type Scorer struct {
	config ScorerConfig
}

// NewScorer creates a PhyMotion-style scorer with validated thresholds.
func NewScorer(config ScorerConfig) Scorer {
	def := DefaultScorerConfig()
	if config.MaxSpeed <= 0 {
		config.MaxSpeed = def.MaxSpeed
	}
	if config.MaxAcceleration <= 0 {
		config.MaxAcceleration = def.MaxAcceleration
	}
	if config.MaxGroundReaction <= 0 {
		config.MaxGroundReaction = def.MaxGroundReaction
	}
	if config.MaxNormalizedTorque <= 0 {
		config.MaxNormalizedTorque = def.MaxNormalizedTorque
	}
	if config.MaxMetabolicEffort <= 0 {
		config.MaxMetabolicEffort = def.MaxMetabolicEffort
	}
	if config.FootContactTolerance <= 0 {
		config.FootContactTolerance = def.FootContactTolerance
	}
	if config.MinimumSupportRadius <= 0 {
		config.MinimumSupportRadius = def.MinimumSupportRadius
	}
	return Scorer{config: config}
}

// ScoreTrajectory computes a normalized feasibility score in [0, 1].
func (s Scorer) ScoreTrajectory(samples []TrajectorySample) Score {
	if len(samples) < 2 {
		return Score{Overall: 0, Problems: []string{"trajectory requires at least two samples"}}
	}

	ordered := append([]TrajectorySample(nil), samples...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Time.Before(ordered[j].Time)
	})

	kinematic, kinematicProblems := s.kinematicContinuity(ordered)
	contact, contactProblems := s.contactBalance(ordered)
	dynamic, dynamicProblems := s.dynamicPlausibility(ordered)
	problems := append(append(kinematicProblems, contactProblems...), dynamicProblems...)

	overall := clamp(0.40*kinematic+0.35*contact+0.25*dynamic, 0, 1)
	return Score{
		Overall: overall,
		Breakdown: ScoreBreakdown{
			KinematicContinuity: kinematic,
			ContactBalance:      contact,
			DynamicPlausibility: dynamic,
		},
		Problems: problems,
	}
}

func (s Scorer) kinematicContinuity(samples []TrajectorySample) (float64, []string) {
	var problems []string
	var speedPenalty, accelerationPenalty float64
	var previousVelocity Vector3
	validSteps := 0

	for i := 1; i < len(samples); i++ {
		dt := secondsBetween(samples[i-1].Time, samples[i].Time)
		if dt <= 0 {
			problems = append(problems, "non-increasing trajectory timestamps")
			continue
		}
		velocity := div(sub(samples[i].CenterOfMass, samples[i-1].CenterOfMass), dt)
		speed := norm(velocity)
		speedPenalty += excessRatio(speed, s.config.MaxSpeed)
		if speed > s.config.MaxSpeed {
			problems = append(problems, "center-of-mass speed exceeds configured limit")
		}
		if validSteps > 0 {
			acceleration := norm(div(sub(velocity, previousVelocity), dt))
			accelerationPenalty += excessRatio(acceleration, s.config.MaxAcceleration)
			if acceleration > s.config.MaxAcceleration {
				problems = append(problems, "center-of-mass acceleration exceeds configured limit")
			}
		}
		previousVelocity = velocity
		validSteps++
	}

	if validSteps == 0 {
		return 0, problems
	}
	averagePenalty := (speedPenalty + accelerationPenalty) / float64(validSteps+maxInt(1, validSteps-1))
	return clamp(1-averagePenalty, 0, 1), dedupe(problems)
}

func (s Scorer) contactBalance(samples []TrajectorySample) (float64, []string) {
	var problems []string
	var totalPenalty float64
	for _, sample := range samples {
		supportRadius := sample.SupportRadius
		if supportRadius <= 0 {
			supportRadius = s.config.MinimumSupportRadius
		}
		midFoot := mul(add(sample.LeftFoot, sample.RightFoot), 0.5)
		horizontalOffset := math.Hypot(sample.CenterOfMass.X-midFoot.X, sample.CenterOfMass.Y-midFoot.Y)
		balancePenalty := excessRatio(horizontalOffset, supportRadius)
		leftFloat := excessRatio(math.Abs(sample.LeftFoot.Z), s.config.FootContactTolerance)
		rightFloat := excessRatio(math.Abs(sample.RightFoot.Z), s.config.FootContactTolerance)
		stepPenalty := clamp(0.70*balancePenalty+0.15*leftFloat+0.15*rightFloat, 0, 1)
		totalPenalty += stepPenalty
		if balancePenalty > 0 {
			problems = append(problems, "center of mass leaves support polygon")
		}
		if leftFloat > 0 || rightFloat > 0 {
			problems = append(problems, "foot contact exceeds tolerance")
		}
	}
	return clamp(1-totalPenalty/float64(len(samples)), 0, 1), dedupe(problems)
}

func (s Scorer) dynamicPlausibility(samples []TrajectorySample) (float64, []string) {
	var problems []string
	var totalPenalty float64
	for _, sample := range samples {
		forcePenalty := excessRatio(sample.GroundReactionForce, s.config.MaxGroundReaction)
		torquePenalty := excessRatio(sample.NormalizedJointTorque, s.config.MaxNormalizedTorque)
		metabolicPenalty := excessRatio(sample.MetabolicEffort, s.config.MaxMetabolicEffort)
		totalPenalty += clamp(0.35*forcePenalty+0.40*torquePenalty+0.25*metabolicPenalty, 0, 1)
		if forcePenalty > 0 {
			problems = append(problems, "ground reaction force exceeds configured limit")
		}
		if torquePenalty > 0 {
			problems = append(problems, "joint torque exceeds configured limit")
		}
		if metabolicPenalty > 0 {
			problems = append(problems, "metabolic effort exceeds configured limit")
		}
	}
	return clamp(1-totalPenalty/float64(len(samples)), 0, 1), dedupe(problems)
}

func secondsBetween(a, b time.Time) float64 {
	if a.IsZero() || b.IsZero() {
		return 1
	}
	return b.Sub(a).Seconds()
}

func add(a, b Vector3) Vector3         { return Vector3{X: a.X + b.X, Y: a.Y + b.Y, Z: a.Z + b.Z} }
func sub(a, b Vector3) Vector3         { return Vector3{X: a.X - b.X, Y: a.Y - b.Y, Z: a.Z - b.Z} }
func mul(a Vector3, k float64) Vector3 { return Vector3{X: a.X * k, Y: a.Y * k, Z: a.Z * k} }
func div(a Vector3, k float64) Vector3 {
	if k == 0 {
		return Vector3{}
	}
	return Vector3{X: a.X / k, Y: a.Y / k, Z: a.Z / k}
}
func norm(a Vector3) float64 { return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z) }

func excessRatio(value, limit float64) float64 {
	if limit <= 0 || value <= limit {
		return 0
	}
	return clamp((value-limit)/limit, 0, 1)
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
