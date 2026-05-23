package phymotion

import (
	"testing"
	"time"
)

func TestScoreTrajectoryRewardsFeasibleMotion(t *testing.T) {
	scorer := NewScorer(DefaultScorerConfig())
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	samples := []TrajectorySample{
		{
			Time:                  now,
			CenterOfMass:          Vector3{X: 0, Y: 0, Z: 1},
			LeftFoot:              Vector3{X: -0.1, Y: 0, Z: 0},
			RightFoot:             Vector3{X: 0.1, Y: 0, Z: 0},
			SupportRadius:         0.4,
			GroundReactionForce:   1.0,
			NormalizedJointTorque: 0.4,
			MetabolicEffort:       0.3,
		},
		{
			Time:                  now.Add(500 * time.Millisecond),
			CenterOfMass:          Vector3{X: 0.05, Y: 0.01, Z: 1},
			LeftFoot:              Vector3{X: -0.1, Y: 0, Z: 0},
			RightFoot:             Vector3{X: 0.1, Y: 0, Z: 0},
			SupportRadius:         0.4,
			GroundReactionForce:   1.1,
			NormalizedJointTorque: 0.5,
			MetabolicEffort:       0.4,
		},
		{
			Time:                  now.Add(time.Second),
			CenterOfMass:          Vector3{X: 0.1, Y: 0.02, Z: 1},
			LeftFoot:              Vector3{X: -0.1, Y: 0, Z: 0},
			RightFoot:             Vector3{X: 0.1, Y: 0, Z: 0},
			SupportRadius:         0.4,
			GroundReactionForce:   1.0,
			NormalizedJointTorque: 0.45,
			MetabolicEffort:       0.35,
		},
	}

	score := scorer.ScoreTrajectory(samples)
	if score.Overall < 0.85 {
		t.Fatalf("expected feasible motion score, got %#v", score)
	}
	if len(score.Problems) != 0 {
		t.Fatalf("expected no problems for feasible trajectory, got %#v", score.Problems)
	}
}

func TestScoreTrajectoryFlagsUnsafeMotion(t *testing.T) {
	scorer := NewScorer(DefaultScorerConfig())
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.UTC)
	samples := []TrajectorySample{
		{
			Time:                  now,
			CenterOfMass:          Vector3{X: 0, Y: 0, Z: 1},
			LeftFoot:              Vector3{X: -0.1, Y: 0, Z: 0.2},
			RightFoot:             Vector3{X: 0.1, Y: 0, Z: 0.2},
			SupportRadius:         0.15,
			GroundReactionForce:   4.0,
			NormalizedJointTorque: 1.5,
			MetabolicEffort:       1.4,
		},
		{
			Time:                  now.Add(100 * time.Millisecond),
			CenterOfMass:          Vector3{X: 2, Y: 0, Z: 1},
			LeftFoot:              Vector3{X: -0.1, Y: 0, Z: 0.2},
			RightFoot:             Vector3{X: 0.1, Y: 0, Z: 0.2},
			SupportRadius:         0.15,
			GroundReactionForce:   4.5,
			NormalizedJointTorque: 2.0,
			MetabolicEffort:       1.8,
		},
	}

	score := scorer.ScoreTrajectory(samples)
	if score.Overall >= 0.7 {
		t.Fatalf("expected unsafe motion score below 0.7, got %#v", score)
	}
	if len(score.Problems) == 0 {
		t.Fatalf("expected unsafe trajectory to report problems")
	}
}
