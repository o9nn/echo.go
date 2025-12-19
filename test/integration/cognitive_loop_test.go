package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/EchoCog/echollama/core/echobeats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCognitiveLoopIntegration tests the full 12-step cognitive loop
func TestCognitiveLoopIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("SharedStateIntegration", func(t *testing.T) {
		state := echobeats.NewSharedCognitiveState()
		require.NotNil(t, state)

		// Simulate a full cognitive cycle through shared state
		for step := 0; step < 12; step++ {
			state.SetCurrentStep(step)
			assert.Equal(t, step, state.GetCurrentStep())

			// Update coherence based on step
			coherence := 1.0 - (float64(step) * 0.05)
			state.UpdateCoherence(coherence)

			// Add context at appropriate steps
			if step >= 1 && step <= 5 {
				state.AddPastContext(map[string]interface{}{
					"step":   step,
					"action": "affordance",
				})
			}

			if step >= 7 && step <= 11 {
				state.AddFutureOption(map[string]interface{}{
					"step":     step,
					"scenario": "salience",
				})
			}

			// Set focus at pivotal steps
			if step == 0 || step == 6 {
				state.SetPresentFocus(map[string]interface{}{
					"pivotal_step": step,
					"relevance":    "realized",
				})
			}
		}

		// Verify final state
		assert.Equal(t, 11, state.GetCurrentStep())
		assert.InDelta(t, 0.45, state.GetCoherence(), 0.01)
	})

	t.Run("ConcurrentEngineSimulation", func(t *testing.T) {
		state := echobeats.NewSharedCognitiveState()
		var wg sync.WaitGroup
		results := make(chan string, 36) // 3 engines × 12 steps

		// Simulate 3 concurrent engines
		engines := []string{"affordance", "relevance", "salience"}

		for _, engine := range engines {
			wg.Add(1)
			go func(engineName string) {
				defer wg.Done()

				for step := 0; step < 12; step++ {
					// Determine if this engine should process this step
					shouldProcess := false
					switch engineName {
					case "affordance":
						shouldProcess = step >= 0 && step <= 5
					case "relevance":
						shouldProcess = step == 0 || step == 6
					case "salience":
						shouldProcess = step >= 6 && step <= 11
					}

					if shouldProcess {
						// Simulate processing
						time.Sleep(time.Millisecond * 10)
						results <- engineName + "-" + string(rune('0'+step))
					}
				}
			}(engine)
		}

		wg.Wait()
		close(results)

		// Count results
		resultCount := 0
		for range results {
			resultCount++
		}

		// Affordance: 6 steps (0-5)
		// Relevance: 2 steps (0, 6)
		// Salience: 6 steps (6-11)
		// Total: 14 (some overlap at pivotal steps)
		assert.GreaterOrEqual(t, resultCount, 12)
	})

	t.Run("TriadInterleaving", func(t *testing.T) {
		// Test that triads are correctly interleaved
		triads := map[int]string{
			1:  "pivotal_relevance_1",
			5:  "pivotal_relevance_1",
			9:  "pivotal_relevance_1",
			2:  "affordance_action",
			6:  "affordance_action",
			10: "affordance_action",
			3:  "salience_simulation",
			7:  "salience_simulation",
			11: "salience_simulation",
		}

		for step, expectedTriad := range triads {
			actualTriad := echobeats.GetTriadForStep(step)
			assert.Equal(t, expectedTriad, actualTriad, "Step %d triad mismatch", step)
		}
	})

	t.Run("PhaseTransitions", func(t *testing.T) {
		// Test phase transitions through the 12-step cycle
		phases := make([]string, 12)
		for step := 0; step < 12; step++ {
			phases[step] = echobeats.PhaseForStep(step)
		}

		// Verify phase sequence
		assert.Equal(t, "relevance", phases[0])  // Pivotal
		assert.Equal(t, "affordance", phases[1]) // Start affordance
		assert.Equal(t, "affordance", phases[5]) // End affordance
		assert.Equal(t, "relevance", phases[6])  // Pivotal
		assert.Equal(t, "salience", phases[7])   // Start salience
		assert.Equal(t, "salience", phases[11])  // End salience
	})
}

// TestMemoryIntegration tests memory system integration
func TestMemoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("MemoryContextFlow", func(t *testing.T) {
		state := echobeats.NewSharedCognitiveState()

		// Simulate memory flow through cognitive cycle
		// Step 0: Initialize with present focus
		state.SetPresentFocus(map[string]interface{}{
			"attention": "current_task",
			"timestamp": time.Now(),
		})

		// Steps 1-5: Build past context from affordances
		for i := 1; i <= 5; i++ {
			state.AddPastContext(map[string]interface{}{
				"step":       i,
				"experience": "past_action_" + string(rune('0'+i)),
				"success":    0.8,
			})
		}

		// Step 6: Reorient
		state.SetPresentFocus(map[string]interface{}{
			"attention": "future_planning",
			"timestamp": time.Now(),
		})

		// Steps 7-11: Build future options from salience
		for i := 7; i <= 11; i++ {
			state.AddFutureOption(map[string]interface{}{
				"step":        i,
				"scenario":    "future_scenario_" + string(rune('0'+i)),
				"probability": 0.7,
			})
		}

		// Verify state consistency
		state.mu.RLock()
		assert.NotEmpty(t, state.pastContext)
		assert.NotEmpty(t, state.futureOptions)
		assert.NotNil(t, state.presentFocus)
		state.mu.RUnlock()
	})
}

// TestCoherenceTracking tests coherence score tracking across cycles
func TestCoherenceTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("CoherenceOverCycles", func(t *testing.T) {
		state := echobeats.NewSharedCognitiveState()
		coherenceHistory := make([]float64, 0)

		// Simulate multiple cognitive cycles
		for cycle := 0; cycle < 5; cycle++ {
			for step := 0; step < 12; step++ {
				state.SetCurrentStep(step)

				// Calculate coherence based on cycle progress
				// Coherence should increase as the system stabilizes
				baseCoherence := 0.5 + (float64(cycle) * 0.1)
				stepAdjustment := float64(step) * 0.01
				coherence := baseCoherence + stepAdjustment

				if coherence > 1.0 {
					coherence = 1.0
				}

				state.UpdateCoherence(coherence)
			}
			coherenceHistory = append(coherenceHistory, state.GetCoherence())
		}

		// Verify coherence trend (should generally increase)
		for i := 1; i < len(coherenceHistory); i++ {
			assert.GreaterOrEqual(t, coherenceHistory[i], coherenceHistory[i-1]*0.9,
				"Coherence should not decrease significantly")
		}
	})
}

// BenchmarkCognitiveLoopIntegration benchmarks the full cognitive loop
func BenchmarkCognitiveLoopIntegration(b *testing.B) {
	b.Run("FullCycleSimulation", func(b *testing.B) {
		state := echobeats.NewSharedCognitiveState()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for step := 0; step < 12; step++ {
				state.SetCurrentStep(step)
				state.UpdateCoherence(0.9)

				if step >= 1 && step <= 5 {
					state.AddPastContext(step)
				}
				if step >= 7 && step <= 11 {
					state.AddFutureOption(step)
				}
				if step == 0 || step == 6 {
					state.SetPresentFocus(step)
				}
			}
		}
	})

	b.Run("ConcurrentCycles", func(b *testing.B) {
		state := echobeats.NewSharedCognitiveState()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for step := 0; step < 12; step++ {
					state.SetCurrentStep(step)
					state.UpdateCoherence(0.9)
				}
			}
		})
	})
}
