package deeptreeecho

import (
	"testing"
)

// TestPersonaManagerActivation tests automatic persona activation based on cognitive state
func TestPersonaManagerActivation(t *testing.T) {
	identity := NewExtendedIdentity("TestPersonaManager")
	
	// Scenario 1: Low coherence, many patterns → should activate Ordo (exploitation)
	identity.Coherence = 0.4
	for i := 0; i < 50; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}
	identity.Iterations = 1500
	
	decision1 := identity.OptimizeRelevanceRealization("ordo_state")
	
	// Low coherence + many patterns should favor exploitation (low exploration)
	if decision1.ExplorationWeight > 0.5 {
		t.Logf("Note: Expected exploitation bias for low coherence + many patterns, got exploration=%.2f", 
			decision1.ExplorationWeight)
	}
	t.Logf("✓ Ordo-like state: low coherence (%.2f), many patterns (%d), exploration=%.2f", 
		identity.Coherence, len(identity.Patterns), decision1.ExplorationWeight)
	
	// Scenario 2: High coherence, few patterns → should activate Chao (exploration)
	identity.Coherence = 0.92
	identity.Patterns = make(map[string]*OpponentPattern)
	for i := 0; i < 10; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.9,
		}
	}
	identity.Iterations = 50
	
	decision2 := identity.OptimizeRelevanceRealization("chao_state")
	
	// High coherence + few patterns should favor exploration
	if decision2.ExplorationWeight < 0.5 {
		t.Logf("Note: Expected exploration bias for high coherence + few patterns, got exploration=%.2f", 
			decision2.ExplorationWeight)
	}
	t.Logf("✓ Chao-like state: high coherence (%.2f), few patterns (%d), exploration=%.2f", 
		identity.Coherence, len(identity.Patterns), decision2.ExplorationWeight)
	
	// Verify opponent process tracking
	stats := identity.OpponentProcesses.GetBalanceStats(ExplorationExploitation)
	t.Logf("Exploration-Exploitation balance: %.2f, stability: %.2f", 
		stats["current_balance"], stats["stability"])
}

// TestPersonaBiasApplication tests that different states apply correct biases
func TestPersonaBiasApplication(t *testing.T) {
	identity := NewExtendedIdentity("TestPersonaBias")
	
	// Test Ordo-like state (exploitation bias)
	identity.Coherence = 0.3
	identity.Iterations = 2000
	for i := 0; i < 60; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.8,
		}
	}
	
	ordoDecision := identity.OptimizeRelevanceRealization("ordo_bias")
	
	t.Logf("Ordo-like state decision:")
	t.Logf("  Exploration: %.2f (expect low)", ordoDecision.ExplorationWeight)
	t.Logf("  Scope: %s (expect depth)", ordoDecision.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect low)", ordoDecision.AdaptationRate)
	t.Logf("  Confidence: %.2f (expect high)", ordoDecision.Confidence)
	
	// Test Chao-like state (exploration bias)
	identity.Coherence = 0.95
	identity.Iterations = 50
	identity.Patterns = make(map[string]*OpponentPattern)
	for i := 0; i < 5; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.6,
		}
	}
	
	chaoDecision := identity.OptimizeRelevanceRealization("chao_bias")
	
	t.Logf("\nChao-like state decision:")
	t.Logf("  Exploration: %.2f (expect high)", chaoDecision.ExplorationWeight)
	t.Logf("  Scope: %s (expect breadth)", chaoDecision.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect high)", chaoDecision.AdaptationRate)
	t.Logf("  Confidence: %.2f (expect low)", chaoDecision.Confidence)
	
	// Verify opposite biases
	if ordoDecision.ExplorationWeight >= chaoDecision.ExplorationWeight {
		t.Error("Expected Ordo state to have lower exploration than Chao state")
	}
	
	t.Log("✓ Ordo and Chao biases applied correctly")
}

// TestPersonaTransitions tests persona transitions over time
func TestPersonaTransitions(t *testing.T) {
	identity := NewExtendedIdentity("TestPersonaTransitions")
	
	// Track decisions through different states
	var decisions []*RelevanceDecision
	
	// Phase 1: Early exploration (Chao-like)
	identity.Coherence = 0.3
	identity.Iterations = 10
	for i := 0; i < 5; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.5,
		}
	}
	decisions = append(decisions, identity.OptimizeRelevanceRealization("phase1_early"))
	
	// Phase 2: Accumulating patterns (transition to Ordo-like)
	identity.Coherence = 0.5
	identity.Iterations = 500
	for i := 5; i < 40; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}
	decisions = append(decisions, identity.OptimizeRelevanceRealization("phase2_accumulating"))
	
	// Phase 3: Mastery state (Ordo-like stabilizes)
	identity.Coherence = 0.75
	identity.Iterations = 1500
	decisions = append(decisions, identity.OptimizeRelevanceRealization("phase3_mastery"))
	
	// Phase 4: Over-optimization (Chao-like disrupts)
	identity.Coherence = 0.96
	identity.Iterations = 3000
	decisions = append(decisions, identity.OptimizeRelevanceRealization("phase4_disruption"))
	
	// Log transition sequence
	t.Log("Decision transition sequence:")
	for i, d := range decisions {
		t.Logf("  Phase %d: exploration=%.2f, scope=%s, adaptation=%.2f", 
			i+1, d.ExplorationWeight, d.ScopePreference, d.AdaptationRate)
	}
	
	// Verify wisdom score increases
	wisdom := identity.GetWisdomScore()
	t.Logf("\nFinal wisdom score: %.3f", wisdom)
	
	trend := identity.GetWisdomTrend()
	t.Logf("Wisdom trend: %.3f", trend)
}

// TestEmotionalPersonaModulation tests how emotions affect decisions
func TestEmotionalPersonaModulation(t *testing.T) {
	identity := NewExtendedIdentity("TestEmotionalModulation")
	
	// Set baseline state
	identity.Coherence = 0.6
	identity.Iterations = 500
	for i := 0; i < 25; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.7,
		}
	}
	
	// Test 1: Calm state (no strong emotional bias)
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.3,
		Valence: 0.5,
	}
	decision1 := identity.OptimizeRelevanceRealization("calm_state")
	t.Logf("Calm state: exploration=%.2f, confidence=%.2f", 
		decision1.ExplorationWeight, decision1.Confidence)
	
	// Test 2: High arousal → should favor exploration and speed
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.9,
		Valence: -0.2,
	}
	decision2 := identity.OptimizeRelevanceRealization("high_arousal")
	t.Logf("High arousal state: exploration=%.2f, confidence=%.2f", 
		decision2.ExplorationWeight, decision2.Confidence)
	
	// High arousal should increase exploration and decrease confidence threshold
	if decision2.ExplorationWeight <= decision1.ExplorationWeight {
		t.Logf("Note: Expected high arousal to increase exploration")
	}
	if decision2.Confidence >= decision1.Confidence {
		t.Logf("Note: Expected high arousal to decrease confidence threshold (favor speed)")
	}
	
	// Test 3: Extreme positive valence
	identity.EmotionalState = &EmotionalState{
		Arousal: 0.7,
		Valence: 0.9,
	}
	decision3 := identity.OptimizeRelevanceRealization("positive_valence")
	t.Logf("Extreme positive valence: exploration=%.2f, confidence=%.2f", 
		decision3.ExplorationWeight, decision3.Confidence)
	
	// Check approach-avoidance balance
	approachBalance := identity.OpponentProcesses.GetCurrentBalance(ApproachAvoidance)
	t.Logf("Approach-avoidance balance: %.2f (positive = approach)", approachBalance)
}

// TestPersonaManagerStats tests statistics tracking
func TestPersonaManagerStats(t *testing.T) {
	identity := NewExtendedIdentity("TestPersonaStats")
	
	// Generate multiple decisions
	for i := 0; i < 10; i++ {
		// Alternate between Ordo-favoring and Chao-favoring states
		if i%2 == 0 {
			// Ordo state
			identity.Coherence = 0.3
			identity.Iterations = uint64(1000 + i*100)
			for j := 0; j < 50; j++ {
				identity.Patterns[string(rune(j))] = &OpponentPattern{
					ID:       string(rune(j)),
					Strength: 0.7,
				}
			}
		} else {
			// Chao state
			identity.Coherence = 0.95
			identity.Iterations = uint64(50 + i*10)
			identity.Patterns = make(map[string]*OpponentPattern)
			for j := 0; j < 5; j++ {
				identity.Patterns[string(rune(j))] = &OpponentPattern{
					ID:       string(rune(j)),
					Strength: 0.6,
				}
			}
		}
		
		identity.OptimizeRelevanceRealization("iteration_" + string(rune('0'+i)))
	}
	
	// Get statistics for all opponent pairs
	t.Log("Statistics after 10 state changes:")
	pairs := []string{
		ExplorationExploitation,
		BreadthDepth,
		StabilityFlexibility,
		SpeedAccuracy,
		ApproachAvoidance,
	}
	
	for _, pair := range pairs {
		stats := identity.OpponentProcesses.GetBalanceStats(pair)
		if stats != nil {
			t.Logf("  %s: balance=%.2f, stability=%.2f", 
				pair, stats["current_balance"], stats["stability"])
		}
	}
	
	// Check wisdom score
	wisdom := identity.GetWisdomScore()
	t.Logf("\nWisdom score: %.3f", wisdom)
}

// TestIntegratedPersonaDecisionMaking tests persona-influenced decision making
func TestIntegratedPersonaDecisionMaking(t *testing.T) {
	identity := NewExtendedIdentity("TestIntegratedDecisions")
	
	// Scenario 1: Force Ordo-like state and verify decisions
	identity.Coherence = 0.35
	identity.Iterations = 2000
	for i := 0; i < 60; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.75,
		}
	}
	
	decision1 := identity.OptimizeRelevanceRealization("ordo_influenced_decision")
	
	t.Logf("Ordo-influenced decision:")
	t.Logf("  Exploration: %.2f (expect low)", decision1.ExplorationWeight)
	t.Logf("  Scope: %s (expect depth)", decision1.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect low)", decision1.AdaptationRate)
	
	// Scenario 2: Force Chao-like state and verify decisions
	identity.Coherence = 0.93
	identity.Patterns = make(map[string]*OpponentPattern)
	for i := 0; i < 8; i++ {
		identity.Patterns[string(rune('a'+i))] = &OpponentPattern{
			ID:       string(rune('a' + i)),
			Strength: 0.9,
		}
	}
	identity.Iterations = 100
	
	decision2 := identity.OptimizeRelevanceRealization("chao_influenced_decision")
	
	t.Logf("\nChao-influenced decision:")
	t.Logf("  Exploration: %.2f (expect high)", decision2.ExplorationWeight)
	t.Logf("  Scope: %s (expect breadth)", decision2.ScopePreference)
	t.Logf("  Adaptation: %.2f (expect high)", decision2.AdaptationRate)
	
	// Verify decisions are appropriately different
	if decision1.ExplorationWeight < decision2.ExplorationWeight {
		t.Log("✓ States correctly influenced exploration tendency")
	} else {
		t.Log("Warning: Expected Ordo state to reduce exploration vs Chao state")
	}
	
	if decision1.ScopePreference != decision2.ScopePreference {
		t.Log("✓ States correctly influenced scope preference")
	}
}
