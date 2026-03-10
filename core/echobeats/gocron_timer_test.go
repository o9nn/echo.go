package echobeats

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoCronCycleTimer_BasicCycle(t *testing.T) {
	config := CycleTimerConfig{
		StepInterval:    10 * time.Millisecond,
		DreamInterval:   100 * time.Millisecond,
		MetricsInterval: 200 * time.Millisecond,
		GoalInterval:    50 * time.Millisecond,
	}

	timer, err := NewGoCronCycleTimer(config)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	var stepCount atomic.Int64
	var cycleCount atomic.Int64

	timer.SetBeatStepCallback(func(step int) {
		stepCount.Add(1)
	})

	timer.SetCycleCompleteCallback(func(cycle uint64) {
		cycleCount.Add(1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := timer.Start(ctx); err != nil {
		t.Fatalf("Failed to start timer: %v", err)
	}

	<-ctx.Done()
	time.Sleep(20 * time.Millisecond) // Allow final callbacks to complete

	steps := stepCount.Load()
	cycles := cycleCount.Load()

	if steps < 12 {
		t.Errorf("Expected at least 12 steps, got %d", steps)
	}
	if cycles < 1 {
		t.Errorf("Expected at least 1 complete cycle, got %d", cycles)
	}

	t.Logf("Completed %d steps and %d cycles in 200ms", steps, cycles)
}

func TestStreamForStep(t *testing.T) {
	tests := []struct {
		step     int
		expected string
	}{
		{1, "perception"},
		{2, "action"},
		{3, "simulation"},
		{4, "integration"},
		{5, "perception"},
		{6, "action"},
		{7, "simulation"},
		{8, "integration"},
		{9, "perception"},
		{10, "action"},
		{11, "simulation"},
		{12, "integration"},
	}

	for _, tt := range tests {
		got := StreamForStep(tt.step)
		if got != tt.expected {
			t.Errorf("StreamForStep(%d) = %q, want %q", tt.step, got, tt.expected)
		}
	}
}

func TestPhaseForStep(t *testing.T) {
	tests := []struct {
		step     int
		expected string
	}{
		{1, "sense"},
		{3, "sense"},
		{4, "process"},
		{6, "process"},
		{7, "emit"},
		{9, "emit"},
		{10, "integrate"},
		{12, "integrate"},
	}

	for _, tt := range tests {
		got := PhaseForStep(tt.step)
		if got != tt.expected {
			t.Errorf("PhaseForStep(%d) = %q, want %q", tt.step, got, tt.expected)
		}
	}
}

func TestGoCronCycleTimer_AdjustInterval(t *testing.T) {
	config := DefaultCycleTimerConfig()
	timer, err := NewGoCronCycleTimer(config)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	// Adjust before start (should just update the field)
	err = timer.AdjustStepInterval(200 * time.Millisecond)
	if err != nil {
		t.Errorf("AdjustStepInterval before start failed: %v", err)
	}

	state := timer.GetState()
	if state.StepInterval != 200*time.Millisecond {
		t.Errorf("Expected interval 200ms, got %v", state.StepInterval)
	}
}

func TestGoCronCycleTimer_GetState(t *testing.T) {
	config := DefaultCycleTimerConfig()
	timer, err := NewGoCronCycleTimer(config)
	if err != nil {
		t.Fatalf("Failed to create timer: %v", err)
	}

	state := timer.GetState()
	if state.Running {
		t.Error("Timer should not be running before Start()")
	}
	if state.CurrentStep != 0 {
		t.Errorf("Expected step 0, got %d", state.CurrentStep)
	}
	if state.CycleCount != 0 {
		t.Errorf("Expected cycle 0, got %d", state.CycleCount)
	}
}
