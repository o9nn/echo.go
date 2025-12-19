package echobeats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSharedCognitiveState tests the shared state management
func TestSharedCognitiveState(t *testing.T) {
	t.Run("NewSharedCognitiveState", func(t *testing.T) {
		state := NewSharedCognitiveState()
		require.NotNil(t, state)
		assert.Equal(t, 1.0, state.GetCoherence())
	})

	t.Run("UpdateCoherence", func(t *testing.T) {
		state := NewSharedCognitiveState()
		state.UpdateCoherence(0.85)
		assert.Equal(t, 0.85, state.GetCoherence())
	})

	t.Run("SetAndGetCurrentStep", func(t *testing.T) {
		state := NewSharedCognitiveState()
		state.SetCurrentStep(5)
		assert.Equal(t, 5, state.GetCurrentStep())
	})

	t.Run("AddPastContext", func(t *testing.T) {
		state := NewSharedCognitiveState()
		
		// Add multiple contexts
		for i := 0; i < 15; i++ {
			state.AddPastContext(i)
		}
		
		// Should keep only last 10
		state.mu.RLock()
		assert.LessOrEqual(t, len(state.pastContext), 10)
		state.mu.RUnlock()
	})

	t.Run("SetPresentFocus", func(t *testing.T) {
		state := NewSharedCognitiveState()
		focus := map[string]interface{}{"attention": "test"}
		state.SetPresentFocus(focus)
		
		state.mu.RLock()
		assert.Equal(t, focus, state.presentFocus)
		state.mu.RUnlock()
	})

	t.Run("AddFutureOption", func(t *testing.T) {
		state := NewSharedCognitiveState()
		
		// Add multiple options
		for i := 0; i < 10; i++ {
			state.AddFutureOption(i)
		}
		
		// Should keep only last 5
		state.mu.RLock()
		assert.LessOrEqual(t, len(state.futureOptions), 5)
		state.mu.RUnlock()
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		state := NewSharedCognitiveState()
		done := make(chan bool)

		// Concurrent writers
		go func() {
			for i := 0; i < 100; i++ {
				state.UpdateCoherence(float64(i) / 100.0)
				state.SetCurrentStep(i % 12)
			}
			done <- true
		}()

		// Concurrent readers
		go func() {
			for i := 0; i < 100; i++ {
				_ = state.GetCoherence()
				_ = state.GetCurrentStep()
			}
			done <- true
		}()

		<-done
		<-done
	})
}

// TestGoAktConfig tests the configuration
func TestGoAktConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultGoAktConfig()
		require.NotNil(t, config)
		assert.Equal(t, "deep-tree-echo", config.SystemName)
		assert.Equal(t, time.Millisecond*100, config.StepDuration)
		assert.Equal(t, time.Second*5, config.PivotalTimeout)
		assert.Equal(t, 3, config.MaxRetries)
		assert.True(t, config.EnableTelemetry)
	})

	t.Run("CustomConfig", func(t *testing.T) {
		config := &GoAktConfig{
			SystemName:      "custom-system",
			StepDuration:    time.Millisecond * 50,
			PivotalTimeout:  time.Second * 10,
			MaxRetries:      5,
			EnableTelemetry: false,
		}
		assert.Equal(t, "custom-system", config.SystemName)
		assert.Equal(t, time.Millisecond*50, config.StepDuration)
	})
}

// TestCognitiveSystemMetrics tests the metrics tracking
func TestCognitiveSystemMetrics(t *testing.T) {
	t.Run("CreateMetrics", func(t *testing.T) {
		metrics := &CognitiveSystemMetrics{
			Running:        true,
			CycleCount:     10,
			LastCycleTime:  time.Now(),
			CoherenceScore: 0.95,
		}

		assert.True(t, metrics.Running)
		assert.Equal(t, uint64(10), metrics.CycleCount)
		assert.Equal(t, 0.95, metrics.CoherenceScore)
	})
}

// TestMessageTypes tests the actor message types
func TestMessageTypes(t *testing.T) {
	t.Run("StartCycleMsg", func(t *testing.T) {
		msg := &StartCycleMsg{
			Timestamp: time.Now(),
		}
		assert.False(t, msg.Timestamp.IsZero())
	})

	t.Run("StepMsg", func(t *testing.T) {
		msg := &StepMsg{
			StepNumber: 5,
			Timestamp:  time.Now(),
			Payload:    "test payload",
		}
		assert.Equal(t, 5, msg.StepNumber)
		assert.Equal(t, "test payload", msg.Payload)
	})

	t.Run("StepResultMsg", func(t *testing.T) {
		msg := &StepResultMsg{
			StepNumber:     3,
			EngineID:       "affordance",
			Success:        true,
			Output:         "step completed",
			Confidence:     0.9,
			ProcessingTime: time.Millisecond * 50,
			Error:          nil,
		}
		assert.Equal(t, 3, msg.StepNumber)
		assert.Equal(t, "affordance", msg.EngineID)
		assert.True(t, msg.Success)
		assert.Equal(t, 0.9, msg.Confidence)
	})

	t.Run("PivotalSyncMsg", func(t *testing.T) {
		msg := &PivotalSyncMsg{
			PivotalStep:   0,
			EngineID:      "relevance",
			StateSnapshot: map[string]interface{}{"focus": "test"},
			Timestamp:     time.Now(),
		}
		assert.Equal(t, 0, msg.PivotalStep)
		assert.Equal(t, "relevance", msg.EngineID)
	})

	t.Run("SyncAckMsg", func(t *testing.T) {
		msg := &SyncAckMsg{
			PivotalStep: 6,
			EngineID:    "salience",
			Ready:       true,
		}
		assert.Equal(t, 6, msg.PivotalStep)
		assert.True(t, msg.Ready)
	})

	t.Run("StateUpdateMsg", func(t *testing.T) {
		msg := &StateUpdateMsg{
			SourceEngine: "affordance",
			UpdateType:   "attention",
			StateData:    []byte("state data"),
			Timestamp:    time.Now(),
		}
		assert.Equal(t, "affordance", msg.SourceEngine)
		assert.Equal(t, "attention", msg.UpdateType)
	})
}

// TestTriadSteps tests the triad step groupings
func TestTriadSteps(t *testing.T) {
	t.Run("TriadDefinitions", func(t *testing.T) {
		assert.Contains(t, TriadSteps, "pivotal_relevance_1")
		assert.Contains(t, TriadSteps, "affordance_action")
		assert.Contains(t, TriadSteps, "salience_simulation")
		assert.Contains(t, TriadSteps, "meta_reflection")
	})

	t.Run("GetTriadForStep", func(t *testing.T) {
		testCases := []struct {
			step     int
			expected string
		}{
			{1, "pivotal_relevance_1"},
			{5, "pivotal_relevance_1"},
			{9, "pivotal_relevance_1"},
			{2, "affordance_action"},
			{6, "affordance_action"},
			{10, "affordance_action"},
			{3, "salience_simulation"},
			{7, "salience_simulation"},
			{11, "salience_simulation"},
		}

		for _, tc := range testCases {
			result := GetTriadForStep(tc.step)
			assert.Equal(t, tc.expected, result, "Step %d should be in triad %s", tc.step, tc.expected)
		}
	})

	t.Run("PhaseForStep", func(t *testing.T) {
		testCases := []struct {
			step     int
			expected string
		}{
			{0, "relevance"},
			{6, "relevance"},
			{1, "affordance"},
			{2, "affordance"},
			{3, "affordance"},
			{4, "affordance"},
			{5, "affordance"},
			{7, "salience"},
			{8, "salience"},
			{9, "salience"},
			{10, "salience"},
			{11, "salience"},
		}

		for _, tc := range testCases {
			result := PhaseForStep(tc.step)
			assert.Equal(t, tc.expected, result, "Step %d should be in phase %s", tc.step, tc.expected)
		}
	})
}

// TestAffordance tests the Affordance struct
func TestAffordance(t *testing.T) {
	t.Run("CreateAffordance", func(t *testing.T) {
		aff := Affordance{
			Action:      "test-action",
			Context:     map[string]interface{}{"key": "value"},
			PastSuccess: 0.8,
			Confidence:  0.9,
			Timestamp:   time.Now(),
		}

		assert.Equal(t, "test-action", aff.Action)
		assert.Equal(t, 0.8, aff.PastSuccess)
		assert.Equal(t, 0.9, aff.Confidence)
	})
}

// TestScenario tests the Scenario struct
func TestScenario(t *testing.T) {
	t.Run("CreateScenario", func(t *testing.T) {
		scenario := Scenario{
			ID:           "scenario-1",
			Description:  "Test scenario",
			Probability:  0.7,
			Desirability: 0.85,
			Consequences: []interface{}{"consequence-1"},
			Timestamp:    time.Now(),
		}

		assert.Equal(t, "scenario-1", scenario.ID)
		assert.Equal(t, "Test scenario", scenario.Description)
		assert.Equal(t, 0.7, scenario.Probability)
		assert.Equal(t, 0.85, scenario.Desirability)
	})

	t.Run("SalienceCalculation", func(t *testing.T) {
		scenario := Scenario{
			Probability:  0.6,
			Desirability: 0.8,
		}

		// Salience = Probability × Desirability
		expectedSalience := scenario.Probability * scenario.Desirability
		assert.Equal(t, 0.48, expectedSalience)
	})
}

// BenchmarkSharedState benchmarks the shared state operations
func BenchmarkSharedState(b *testing.B) {
	b.Run("UpdateCoherence", func(b *testing.B) {
		state := NewSharedCognitiveState()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			state.UpdateCoherence(float64(i%100) / 100.0)
		}
	})

	b.Run("GetCoherence", func(b *testing.B) {
		state := NewSharedCognitiveState()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = state.GetCoherence()
		}
	})

	b.Run("ConcurrentReadWrite", func(b *testing.B) {
		state := NewSharedCognitiveState()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					state.UpdateCoherence(0.5)
				} else {
					_ = state.GetCoherence()
				}
				i++
			}
		})
	})
}
