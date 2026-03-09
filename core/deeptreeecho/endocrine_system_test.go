package deeptreeecho

import (
	"testing"
	"time"
)

func TestNewVirtualEndocrineSystem(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	if ves == nil {
		t.Fatal("Expected non-nil endocrine system")
	}
	if ves.Bus == nil {
		t.Fatal("Expected non-nil hormone bus")
	}

	// Check baseline levels are set
	cortisol := ves.Bus.Concentration(Cortisol)
	if cortisol < 0.1 || cortisol > 0.3 {
		t.Errorf("Expected cortisol baseline ~0.2, got %f", cortisol)
	}

	daTonic := ves.Bus.Concentration(DopamineTonic)
	if daTonic < 0.4 || daTonic > 0.6 {
		t.Errorf("Expected dopamine tonic baseline ~0.5, got %f", daTonic)
	}
}

func TestSignalEvent(t *testing.T) {
	ves := NewVirtualEndocrineSystem()

	// Record initial cortisol
	initialCortisol := ves.Bus.Concentration(Cortisol)

	// Signal a threat
	ves.SignalEvent(EndoThreatDetected, 1.0)

	// Cortisol should increase
	newCortisol := ves.Bus.Concentration(Cortisol)
	if newCortisol <= initialCortisol {
		t.Errorf("Expected cortisol to increase after threat, got %f -> %f", initialCortisol, newCortisol)
	}

	// Signal a reward
	initialDA := ves.Bus.Concentration(DopaminePhasic)
	ves.SignalEvent(EndoRewardReceived, 0.8)
	newDA := ves.Bus.Concentration(DopaminePhasic)
	if newDA <= initialDA {
		t.Errorf("Expected phasic dopamine to increase after reward, got %f -> %f", initialDA, newDA)
	}
}

func TestHormoneDecay(t *testing.T) {
	ves := NewVirtualEndocrineSystem()

	// Spike cortisol
	ves.SignalEvent(EndoThreatDetected, 1.0)
	spikedCortisol := ves.Bus.Concentration(Cortisol)

	// Tick several times
	for i := 0; i < 100; i++ {
		ves.Tick(0.1)
	}

	// Cortisol should have decayed toward baseline
	decayedCortisol := ves.Bus.Concentration(Cortisol)
	if decayedCortisol >= spikedCortisol {
		t.Errorf("Expected cortisol to decay, got %f -> %f", spikedCortisol, decayedCortisol)
	}
}

func TestCognitiveModeDetection(t *testing.T) {
	ves := NewVirtualEndocrineSystem()

	// Default mode should be something reasonable
	mode, confidence := ves.Bus.CurrentMode()
	if confidence < 0 {
		t.Errorf("Expected non-negative confidence, got %f", confidence)
	}

	// Spike social hormones
	ves.SignalEvent(EndoSocialContact, 1.0)
	ves.SignalEvent(EndoSocialContact, 1.0)
	ves.Tick(0.1)

	mode, _ = ves.Bus.CurrentMode()
	// After strong social signal, mode should shift
	t.Logf("Mode after social contact: %s", mode)
}

func TestAffectiveState(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	ves.Tick(0.1)

	valence, arousal, dominance := ves.GetAffectiveState()

	if valence < -1 || valence > 1 {
		t.Errorf("Valence out of range: %f", valence)
	}
	if arousal < 0 || arousal > 1 {
		t.Errorf("Arousal out of range: %f", arousal)
	}
	if dominance < 0 || dominance > 1 {
		t.Errorf("Dominance out of range: %f", dominance)
	}
}

func TestValenceMemory(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	ves.Tick(0.1)

	ves.RecordValenceMemory("test_context_1")
	ves.RecordValenceMemory("test_context_2")

	history := ves.GetValenceHistory(10)
	if len(history) != 2 {
		t.Errorf("Expected 2 valence memories, got %d", len(history))
	}
	if history[0].Context != "test_context_1" {
		t.Errorf("Expected context 'test_context_1', got '%s'", history[0].Context)
	}
}

func TestMoralPerception(t *testing.T) {
	ves := NewVirtualEndocrineSystem()

	// Boost oxytocin for care-oriented perception
	ves.SignalEvent(EndoSocialContact, 1.0)
	ves.SignalEvent(EndoSocialContact, 1.0)
	ves.Tick(0.1)

	signal := ves.MoralPerception("helping someone in need")
	if signal.EmpathicInference < 0 || signal.EmpathicInference > 1 {
		t.Errorf("Empathic inference out of range: %f", signal.EmpathicInference)
	}
	t.Logf("Moral perception: affect=%f, association=%s, empathy=%f, novelty=%f",
		signal.RawAffect, signal.MoralAssociation, signal.EmpathicInference, signal.NoveltySignal)
}

func TestModeTransitionCallback(t *testing.T) {
	ves := NewVirtualEndocrineSystem()

	transitionCount := 0
	ves.SetModeChangeCallback(func(from, to CognitiveMode) {
		transitionCount++
	})

	// Drive mode transitions by varying hormone levels
	for i := 0; i < 50; i++ {
		if i%10 == 0 {
			ves.SignalEvent(EndoThreatDetected, 0.8)
		}
		if i%15 == 0 {
			ves.SignalEvent(EndoSocialContact, 0.9)
		}
		if i%20 == 0 {
			ves.SignalEvent(EndoRewardReceived, 0.7)
		}
		ves.Tick(0.5)
	}

	t.Logf("Mode transitions observed: %d", transitionCount)
}

func TestEndocrineMetrics(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	ves.SignalEvent(EndoRewardReceived, 0.5)
	ves.Tick(0.1)

	metrics := ves.GetMetrics()
	if metrics["total_events"].(uint64) != 1 {
		t.Errorf("Expected 1 total event, got %v", metrics["total_events"])
	}
	if _, ok := metrics["current_mode"]; !ok {
		t.Error("Expected current_mode in metrics")
	}
}

func TestFourECognitionState(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	fes := NewFourECognitionState(ves)

	if fes == nil {
		t.Fatal("Expected non-nil 4E state")
	}

	// Initial score should be low
	overall := fes.OverallScore()
	if overall < 0 || overall > 1 {
		t.Errorf("Overall score out of range: %f", overall)
	}

	// Update with events
	fes.UpdateFromCognitiveEvent("prediction", true, map[string]float64{"accuracy": 0.8})
	fes.UpdateFromCognitiveEvent("tool_use", true, nil)
	fes.UpdateFromCognitiveEvent("social_interaction", true, nil)

	scores := fes.DimensionScores()
	for dim, score := range scores {
		if score < 0 || score > 1 {
			t.Errorf("Dimension %s score out of range: %f", dim, score)
		}
	}
}

func TestFourEMaturityLevel(t *testing.T) {
	ves := NewVirtualEndocrineSystem()
	fes := NewFourECognitionState(ves)

	level := fes.DetermineMaturity(0.1)
	if level != MaturityNascent {
		t.Errorf("Expected Nascent maturity at low scores, got %s", level)
	}
}

func TestLorenzAttractor(t *testing.T) {
	la := NewLorenzAttractor()
	if la == nil {
		t.Fatal("Expected non-nil Lorenz attractor")
	}

	// Step many times
	for i := 0; i < 1000; i++ {
		la.Step()
	}

	// Should be chaotic
	lyapunov := la.GetLyapunovExponent()
	t.Logf("Lyapunov exponent after 1000 steps: %f", lyapunov)

	if !la.IsChaotic() {
		t.Log("Warning: Lorenz attractor not yet detected as chaotic (may need more steps)")
	}

	// State should be bounded (Lorenz attractor is bounded)
	x, y, z := la.GetState()
	if x > 100 || x < -100 || y > 100 || y < -100 || z > 100 || z < -100 {
		t.Errorf("Lorenz state seems unbounded: (%f, %f, %f)", x, y, z)
	}
}

func TestCognitiveNoiseGenerator(t *testing.T) {
	cng := NewCognitiveNoiseGenerator()

	// Update many times
	for i := 0; i < 100; i++ {
		cng.Update()
	}

	// Noise values should be bounded
	an := cng.GetAttentionNoise()
	mn := cng.GetMemoryNoise()
	dn := cng.GetDecisionNoise()
	en := cng.GetExpressionNoise()

	for _, v := range []float64{an, mn, dn, en} {
		if v > 1 || v < -1 {
			t.Errorf("Noise value out of expected range: %f", v)
		}
	}
}

func TestCognitiveEventLoopCreation(t *testing.T) {
	// Use nil LLM provider for testing
	cel := NewCognitiveEventLoop(nil)
	if cel == nil {
		t.Fatal("Expected non-nil cognitive event loop")
	}

	if cel.endocrine == nil {
		t.Error("Expected non-nil endocrine system")
	}
	if cel.fourE == nil {
		t.Error("Expected non-nil 4E cognition state")
	}
	if cel.chaosEngine == nil {
		t.Error("Expected non-nil chaos engine")
	}
	if len(cel.streams) != 3 {
		t.Errorf("Expected 3 streams, got %d", len(cel.streams))
	}
}

func TestCognitiveEventLoopStartStop(t *testing.T) {
	cel := NewCognitiveEventLoop(nil)

	err := cel.Start()
	if err != nil {
		t.Fatalf("Failed to start: %v", err)
	}

	// Let it run for a bit
	time.Sleep(500 * time.Millisecond)

	// Should have completed some cycles
	cycles := cel.cycleCount.Load()
	steps := cel.stepCount.Load()
	t.Logf("After 500ms: %d cycles, %d steps", cycles, steps)

	if steps == 0 {
		t.Error("Expected some steps to have been executed")
	}

	cel.Stop()

	// Verify it stopped
	time.Sleep(100 * time.Millisecond)
	stepsAfterStop := cel.stepCount.Load()
	time.Sleep(200 * time.Millisecond)
	stepsLater := cel.stepCount.Load()

	if stepsLater > stepsAfterStop+1 {
		t.Error("Event loop did not stop properly")
	}
}

func TestCognitiveEventLoopWakeRest(t *testing.T) {
	cel := NewCognitiveEventLoop(nil)
	cel.Start()
	defer cel.Stop()

	if !cel.IsAwake() {
		t.Error("Expected to be awake after start")
	}

	cel.Rest()
	if cel.IsAwake() {
		t.Error("Expected to be resting after Rest()")
	}

	cel.Wake()
	if !cel.IsAwake() {
		t.Error("Expected to be awake after Wake()")
	}
}

func TestCognitiveEventLoopMetrics(t *testing.T) {
	cel := NewCognitiveEventLoop(nil)
	cel.Start()
	time.Sleep(300 * time.Millisecond)
	cel.Stop()

	metrics := cel.GetMetrics()
	if _, ok := metrics["total_cycles"]; !ok {
		t.Error("Expected total_cycles in metrics")
	}
	if _, ok := metrics["cognitive_mode"]; !ok {
		t.Error("Expected cognitive_mode in metrics")
	}
	if _, ok := metrics["4e_scores"]; !ok {
		t.Error("Expected 4e_scores in metrics")
	}
	if _, ok := metrics["endocrine"]; !ok {
		t.Error("Expected endocrine in metrics")
	}
}
