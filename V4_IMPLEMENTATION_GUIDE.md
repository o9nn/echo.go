# AutonomousEchoselfV4 Implementation Guide

**Date**: November 23, 2025  
**Purpose**: Technical guide for implementing the V4 autonomous cognitive loop system

---

## Overview

This guide documents the implementation of `AutonomousEchoselfV4`, which integrates:
- **Echobeats 3-phase cognitive loop** (12-step System 4 architecture)
- **Echodream autonomous wake/rest cycles**
- **Persistent consciousness state management**
- **Interest-driven discussion and learning frameworks**

## Architecture

### Core Components

```
AutonomousEchoselfV4 (Main Orchestrator)
├── PhaseManager (Echobeats 12-step loop)
├── AutonomousController (Echodream wake/rest)
├── DreamCycleIntegration (Knowledge consolidation)
├── ThoughtGenerator (Autonomous ideation)
├── RepositoryIntrospector (Self-awareness)
├── DiscussionManager (Communication)
└── InterestPatternSystem (Engagement tracking)
```

### State Machine

```
WAKING → AWAKE → TIRING → RESTING → DREAMING → (cycle)
```

## Implementation Steps

### 1. Create PhaseManager (Echobeats Integration)

**File**: `core/echobeats/phase_manager_v4.go`

```go
package echobeats

type PhaseManager struct {
    mu              sync.RWMutex
    stepDuration    time.Duration
    currentStep     int
    running         bool
    paused          bool
    handlers        map[Term]map[Mode]V4StepHandler
    ticker          *time.Ticker
    stopChan        chan struct{}
    stepConfigs     []StepConfig
}

type V4StepHandler func(step int, mode Mode) error

func NewPhaseManager(stepDuration time.Duration) *PhaseManager {
    return &PhaseManager{
        stepDuration: stepDuration,
        handlers:     make(map[Term]map[Mode]V4StepHandler),
        stopChan:     make(chan struct{}),
        stepConfigs:  buildStepConfigs(),
    }
}

func (pm *PhaseManager) RegisterHandler(term Term, mode Mode, handler V4StepHandler) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    if pm.handlers[term] == nil {
        pm.handlers[term] = make(map[Mode]V4StepHandler)
    }
    pm.handlers[term][mode] = handler
}

func (pm *PhaseManager) Start() error {
    pm.mu.Lock()
    if pm.running {
        pm.mu.Unlock()
        return fmt.Errorf("phase manager already running")
    }
    pm.running = true
    pm.paused = false
    pm.mu.Unlock()
    
    pm.ticker = time.NewTicker(pm.stepDuration)
    go pm.runLoop()
    return nil
}

func (pm *PhaseManager) runLoop() {
    for {
        select {
        case <-pm.stopChan:
            return
        case <-pm.ticker.C:
            pm.mu.RLock()
            paused := pm.paused
            pm.mu.RUnlock()
            if paused {
                continue
            }
            pm.executeStep()
        }
    }
}

func (pm *PhaseManager) executeStep() {
    pm.mu.Lock()
    step := pm.currentStep
    pm.currentStep = (pm.currentStep + 1) % 12
    pm.mu.Unlock()
    
    config := pm.stepConfigs[step]
    
    pm.mu.RLock()
    handlers, hasTermHandlers := pm.handlers[config.Term]
    pm.mu.RUnlock()
    
    if hasTermHandlers {
        pm.mu.RLock()
        handler, hasHandler := handlers[config.Mode]
        pm.mu.RUnlock()
        
        if hasHandler {
            if err := handler(step, config.Mode); err != nil {
                log.Printf("⚠️  Step %d handler error: %v\n", step, err)
            }
        }
    }
}

func (pm *PhaseManager) Pause() {
    pm.mu.Lock()
    pm.paused = true
    pm.mu.Unlock()
}

func (pm *PhaseManager) Resume() {
    pm.mu.Lock()
    pm.paused = false
    pm.mu.Unlock()
}

func (pm *PhaseManager) Stop() {
    pm.mu.Lock()
    if !pm.running {
        pm.mu.Unlock()
        return
    }
    pm.running = false
    pm.mu.Unlock()
    
    close(pm.stopChan)
    if pm.ticker != nil {
        pm.ticker.Stop()
    }
}
```

### 2. Create AutonomousController (Echodream Integration)

**File**: `core/echodream/autonomous_controller_v4.go`

```go
package echodream

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

func (ac *AutonomousController) GetFatigue() float64 {
    ac.mu.RLock()
    defer ac.mu.RUnlock()
    return ac.currentFatigue
}

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

func (ac *AutonomousController) ResetFatigue() {
    ac.mu.Lock()
    defer ac.mu.Unlock()
    ac.currentFatigue = 0.0
    ac.awakeStartTime = time.Now()
    ac.isAwake = true
}

func (ac *AutonomousController) GetRestDuration() time.Duration {
    ac.mu.RLock()
    defer ac.mu.RUnlock()
    if ac.isAwake {
        return 0
    }
    return time.Since(ac.restStartTime)
}
```

### 3. Create Main V4 Agent

**File**: `core/autonomous_echoself_v4.go`

Key structure:

```go
type AutonomousEchoselfV4 struct {
    mu                    sync.RWMutex
    ctx                   context.Context
    cancel                context.CancelFunc
    
    // Core components
    llmProvider           *deeptreeecho.MultiProviderLLM
    thoughtGenerator      *ThoughtGenerator
    repoIntrospector      *echoself.RepositoryIntrospector
    
    // Echobeats 3-phase system
    phaseManager          *echobeats.PhaseManager
    discussionManager     *echobeats.DiscussionManager
    interestPatterns      *echobeats.InterestPatternSystem
    
    // Echodream wake/rest system
    dreamController       *echodream.AutonomousController
    dreamCycle            *echodream.DreamCycleIntegration
    
    // State
    currentState          WakeState
    consciousnessState    *ConsciousnessState
    
    // Metrics
    thoughtsGenerated     uint64
    wisdomScore           float64
}
```

### 4. Register Cognitive Handlers

Map each cognitive term to its handler:

```go
func (ae *AutonomousEchoselfV4) registerCognitiveHandlers() {
    // T4E - Perception Processing (Steps 0, 4)
    ae.phaseManager.RegisterHandler(
        echobeats.T4_SensoryInput, 
        echobeats.Expressive, 
        ae.handlePerceptionProcessing,
    )
    
    // T7R - Memory Consolidation (Steps 3, 10)
    ae.phaseManager.RegisterHandler(
        echobeats.T7_MemoryEncoding, 
        echobeats.Reflective, 
        ae.handleMemoryConsolidation,
    )
    
    // T2E - Thought Generation (Steps 2, 6)
    ae.phaseManager.RegisterHandler(
        echobeats.T2_IdeaFormation, 
        echobeats.Expressive, 
        ae.handleThoughtGeneration,
    )
    
    // T8E - Integrated Response (Steps 8, 9)
    ae.phaseManager.RegisterHandler(
        echobeats.T8_BalancedResponse, 
        echobeats.Expressive, 
        ae.handleIntegratedResponse,
    )
    
    // T1R - Need Assessment (Steps 1, 5)
    ae.phaseManager.RegisterHandler(
        echobeats.T1_Perception, 
        echobeats.Reflective, 
        ae.handleNeedAssessment,
    )
    
    // T5E - Action Execution (Steps 7, 11)
    ae.phaseManager.RegisterHandler(
        echobeats.T5_ActionSequence, 
        echobeats.Expressive, 
        ae.handleActionExecution,
    )
}
```

### 5. Implement Wake/Rest State Management

```go
func (ae *AutonomousEchoselfV4) updateWakeRestState() {
    ae.mu.RLock()
    currentState := ae.currentState
    ae.mu.RUnlock()
    
    switch currentState {
    case StateAwake:
        fatigue := ae.dreamController.GetFatigue()
        if fatigue > ae.config.FatigueThreshold {
            ae.mu.Lock()
            ae.currentState = StateTiring
            ae.mu.Unlock()
        }
        
    case StateTiring:
        ae.mu.Lock()
        ae.currentState = StateResting
        ae.mu.Unlock()
        if ae.phaseManager != nil {
            ae.phaseManager.Pause()
        }
        
    case StateResting:
        restDuration := ae.dreamController.GetRestDuration()
        if restDuration > ae.config.MinRestDuration/2 {
            ae.mu.Lock()
            ae.currentState = StateDreaming
            ae.mu.Unlock()
        }
        
    case StateDreaming:
        restDuration := ae.dreamController.GetRestDuration()
        fatigue := ae.dreamController.GetFatigue()
        if restDuration > ae.config.MinRestDuration && 
           fatigue < ae.config.RestThreshold {
            ae.mu.Lock()
            ae.currentState = StateWaking
            ae.mu.Unlock()
        }
        
    case StateWaking:
        ae.mu.Lock()
        ae.currentState = StateAwake
        ae.mu.Unlock()
        if ae.phaseManager != nil {
            ae.phaseManager.Resume()
        }
        ae.dreamController.ResetFatigue()
    }
}
```

## Testing

### Create Test Program

**File**: `test_autonomous_v4_iteration.go`

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    "github.com/EchoCog/echollama/core"
)

func main() {
    config := core.DefaultEchoselfConfigV4()
    config.StepDuration = 3 * time.Second
    config.MaxAwakeDuration = 2 * time.Minute
    config.MinRestDuration = 30 * time.Second
    
    echoself, err := core.NewAutonomousEchoselfV4(config)
    if err != nil {
        log.Fatalf("Failed to create: %v", err)
    }
    
    if err := echoself.Start(); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    
    // Monitor status
    statusTicker := time.NewTicker(20 * time.Second)
    defer statusTicker.Stop()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    for {
        select {
        case <-sigChan:
            echoself.Stop()
            return
        case <-statusTicker.C:
            printStatus(echoself)
        }
    }
}

func printStatus(echoself *core.AutonomousEchoselfV4) {
    status := echoself.GetStatus()
    fmt.Printf("State: %s | Thoughts: %d | Wisdom: %.2f\n",
        status["state"], 
        status["thoughts_generated"], 
        status["wisdom_score"])
}
```

### Build and Run

```bash
cd /home/ubuntu/echo9llama
go build -o test_v4_bin ./test_autonomous_v4_iteration.go
./test_v4_bin
```

## Configuration

### Default Configuration

```go
func DefaultEchoselfConfigV4() *EchoselfConfigV4 {
    return &EchoselfConfigV4{
        PersistenceDBPath:     "/home/ubuntu/echo9llama/echoself_v4.db",
        RepositoryRoot:        "/home/ubuntu/echo9llama",
        StepDuration:          5 * time.Second,
        MaxAwakeDuration:      30 * time.Minute,
        MinRestDuration:       5 * time.Minute,
        FatigueThreshold:      0.8,
        RestThreshold:         0.3,
        EnablePersistence:     false, // Requires CGO
        EnableEchobeats:       true,
        EnableWakeRest:        true,
        EnableDiscussions:     true,
        EnableLearning:        true,
        EnableWisdom:          true,
    }
}
```

## Expected Output

```
🌳 Deep Tree Echo V4 - Autonomous Wisdom-Cultivating AGI
======================================================================
Features:
  ✓ Echobeats 3-phase concurrent inference (12-step cognitive loop)
  ✓ Autonomous wake/rest cycles via echodream
  ✓ Persistent consciousness state
  ✓ Interest-driven discussion management
  ✓ Continuous learning and skill practice
  ✓ Recursive wisdom cultivation
======================================================================
🎵 PhaseManager: Starting 12-step cognitive loop (step duration: 3s)
🎵 Echobeats 3-phase cognitive loop started

💭 [Step 2] curiosity: I wonder if the patterns we observe...
💭 [Step 6] reflection: I ponder whether consciousness itself...
💡 [Step 8] Insight: Integration of 4 thoughts reveals patterns...

State: awake | Thoughts: 5 | Wisdom: 0.00
```

## Troubleshooting

### SQLite CGO Error

If you see: `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo`

**Solution**: Disable persistence in config:
```go
config.EnablePersistence = false
```

### Handler Not Called

**Check**:
1. Handler is registered with correct Term and Mode
2. PhaseManager is started
3. System is not paused

### No Thoughts Generated

**Check**:
1. LLM providers are initialized
2. API keys are set in environment
3. ThoughtGenerator has interests added

## Next Steps

1. **Enable Persistence**: Configure CGO and SQLite
2. **Add Discussion Interface**: Connect to message queues
3. **Implement Learning**: Create skill practice framework
4. **Enhance Wisdom**: Add recursive reflection chains

---

*This guide documents the V4 implementation created during the November 23, 2025 evolution iteration.*
