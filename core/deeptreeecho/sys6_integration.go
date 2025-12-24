package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// =============================================================================
// SYS6 INTEGRATION LAYER
// =============================================================================
//
// This module integrates the sys6 operad architecture with the existing
// echobeats cognitive loop, providing:
//
// 1. Mapping between 12-step echobeats and 30-step sys6 cycles
// 2. LLM-powered cognitive processing for each sys6 component
// 3. Unified state management across both systems
// 4. Introspection commands for sys6 status
//
// =============================================================================

// Sys6Integration bridges the sys6 operad with echobeats
type Sys6Integration struct {
	mu sync.RWMutex

	// Core components
	sys6       *Sys6Operad
	echobeats  *EchobeatsUnified
	llmProvider llm.LLMProvider

	// State synchronization
	stateSync *Sys6StateSync

	// LLM-powered cognitive processors
	c8Processor  *C8CognitiveProcessor
	k9Processor  *K9CognitiveProcessor
	phiProcessor *PhiCognitiveProcessor

	// Running state
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// Sys6StateSync manages state synchronization between sys6 and echobeats
type Sys6StateSync struct {
	mu sync.RWMutex

	// Mapping: 30 sys6 steps -> 12 echobeats steps
	// Every 2.5 sys6 steps = 1 echobeats step (30/12 = 2.5)
	stepMapping map[int]int

	// Shared cognitive state
	sharedState map[string]interface{}

	// Sync events
	syncHistory []Sys6SyncEvent
}

// Sys6SyncEvent represents a synchronization event
type Sys6SyncEvent struct {
	Timestamp     time.Time
	Sys6Step      int
	EchobeatsStep int
	EventType     string
	Description   string
}

// NewSys6Integration creates a new sys6 integration layer
func NewSys6Integration(llmProvider llm.LLMProvider) *Sys6Integration {
	ctx, cancel := context.WithCancel(context.Background())

	si := &Sys6Integration{
		sys6:        NewSys6Operad(llmProvider),
		llmProvider: llmProvider,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize state sync
	si.stateSync = &Sys6StateSync{
		stepMapping: make(map[int]int),
		sharedState: make(map[string]interface{}),
		syncHistory: make([]Sys6SyncEvent, 0),
	}

	// Build step mapping (30 -> 12)
	for i := 1; i <= 30; i++ {
		// Map sys6 step to echobeats step
		// Steps 1-3 -> 1, 4-5 -> 2, 6-8 -> 3, etc.
		echobeatsStep := ((i - 1) * 12 / 30) + 1
		si.stateSync.stepMapping[i] = echobeatsStep
	}

	// Initialize cognitive processors
	si.c8Processor = NewC8CognitiveProcessor(llmProvider)
	si.k9Processor = NewK9CognitiveProcessor(llmProvider)
	si.phiProcessor = NewPhiCognitiveProcessor(llmProvider)

	return si
}

// IntegrateWithEchobeats connects to an existing echobeats system
func (si *Sys6Integration) IntegrateWithEchobeats(echobeats *EchobeatsUnified) {
	si.mu.Lock()
	defer si.mu.Unlock()

	si.echobeats = echobeats
	si.sys6.IntegrateWithEchobeats(echobeats)
}

// Start begins the integrated sys6 system
func (si *Sys6Integration) Start() error {
	si.mu.Lock()
	if si.running {
		si.mu.Unlock()
		return fmt.Errorf("sys6 integration already running")
	}
	si.running = true
	si.mu.Unlock()

	// Start the sys6 operad
	if err := si.sys6.Start(); err != nil {
		return fmt.Errorf("failed to start sys6: %w", err)
	}

	// Start the synchronization loop
	go si.runSyncLoop()

	fmt.Println("🔗 Sys6 Integration: Connected to echobeats cognitive loop")
	return nil
}

// Stop halts the integrated sys6 system
func (si *Sys6Integration) Stop() error {
	si.mu.Lock()
	defer si.mu.Unlock()

	if !si.running {
		return fmt.Errorf("sys6 integration not running")
	}

	si.cancel()
	si.running = false

	if err := si.sys6.Stop(); err != nil {
		return fmt.Errorf("failed to stop sys6: %w", err)
	}

	fmt.Println("🔗 Sys6 Integration: Disconnected")
	return nil
}

// runSyncLoop synchronizes state between sys6 and echobeats
func (si *Sys6Integration) runSyncLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-si.ctx.Done():
			return
		case <-ticker.C:
			si.synchronizeState()
		}
	}
}

// synchronizeState syncs state between sys6 and echobeats
func (si *Sys6Integration) synchronizeState() {
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.echobeats == nil {
		return
	}

	// Get current sys6 status
	sys6Status := si.sys6.GetStatus()

	// Map to echobeats step
	echobeatsStep := si.stateSync.stepMapping[int(sys6Status.ClockState.Step)]

	// Record sync event
	si.stateSync.syncHistory = append(si.stateSync.syncHistory, Sys6SyncEvent{
		Timestamp:     time.Now(),
		Sys6Step:      sys6Status.ClockState.Step,
		EchobeatsStep: echobeatsStep,
		EventType:     "sync",
		Description:   fmt.Sprintf("Sys6 step %d -> Echobeats step %d", sys6Status.ClockState.Step, echobeatsStep),
	})

	// Keep only last 100 events
	if len(si.stateSync.syncHistory) > 100 {
		si.stateSync.syncHistory = si.stateSync.syncHistory[1:]
	}

	// Update shared state
	si.stateSync.sharedState["sys6_step"] = sys6Status.ClockState.Step
	si.stateSync.sharedState["echobeats_step"] = echobeatsStep
	si.stateSync.sharedState["dyadic_phase"] = sys6Status.ClockState.DyadicPhase
	si.stateSync.sharedState["triadic_phase"] = sys6Status.ClockState.TriadicPhase
	si.stateSync.sharedState["pentadic_stage"] = sys6Status.ClockState.PentadicStage
	si.stateSync.sharedState["current_stage"] = sys6Status.CurrentStage
}

// GetStatus returns the current integration status
func (si *Sys6Integration) GetStatus() Sys6IntegrationStatus {
	si.mu.RLock()
	defer si.mu.RUnlock()

	sys6Status := si.sys6.GetStatus()

	return Sys6IntegrationStatus{
		Running:       si.running,
		Sys6Status:    sys6Status,
		SyncEventCount: len(si.stateSync.syncHistory),
		SharedState:   si.stateSync.sharedState,
	}
}

// Sys6IntegrationStatus represents the integration status
type Sys6IntegrationStatus struct {
	Running        bool
	Sys6Status     Sys6Status
	SyncEventCount int
	SharedState    map[string]interface{}
}

// =============================================================================
// C₈ COGNITIVE PROCESSOR (LLM-POWERED)
// =============================================================================

// C8CognitiveProcessor provides LLM-powered processing for cubic concurrency
type C8CognitiveProcessor struct {
	llmProvider llm.LLMProvider
	mu          sync.RWMutex
}

// NewC8CognitiveProcessor creates a new C8 processor
func NewC8CognitiveProcessor(llmProvider llm.LLMProvider) *C8CognitiveProcessor {
	return &C8CognitiveProcessor{
		llmProvider: llmProvider,
	}
}

// ProcessState processes a single concurrency state using LLM
func (cp *C8CognitiveProcessor) ProcessState(ctx context.Context, stateID int, binaryCode string, input interface{}) (string, error) {
	systemPrompt := fmt.Sprintf(`You are processing state %d (binary: %s) in the C₈ cubic concurrency system.
This is one of 8 parallel cognitive states derived from 2³ prime-power delegation.

Your state represents the following cognitive aspect based on binary encoding:
- Bit 0 (LSB): Perception/Expression polarity
- Bit 1: Action/Reflection polarity  
- Bit 2 (MSB): Learning/Integration polarity

Binary %s means:
- Perception/Expression: %s
- Action/Reflection: %s
- Learning/Integration: %s

Process the input from your unique cognitive perspective and provide insights.`,
		stateID, binaryCode,
		binaryCode,
		map[bool]string{true: "Expression", false: "Perception"}[binaryCode[2] == '1'],
		map[bool]string{true: "Reflection", false: "Action"}[binaryCode[1] == '1'],
		map[bool]string{true: "Integration", false: "Learning"}[binaryCode[0] == '1'],
	)

	prompt := fmt.Sprintf("Process this input from your cognitive perspective:\n%v", input)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    200,
		Temperature:  0.7,
	}

	return cp.llmProvider.Generate(ctx, prompt, opts)
}

// =============================================================================
// K₉ COGNITIVE PROCESSOR (LLM-POWERED)
// =============================================================================

// K9CognitiveProcessor provides LLM-powered processing for triadic convolution
type K9CognitiveProcessor struct {
	llmProvider llm.LLMProvider
	mu          sync.RWMutex

	// Phase descriptions for the 3x3 grid
	phaseDescriptions [9]string
}

// NewK9CognitiveProcessor creates a new K9 processor
func NewK9CognitiveProcessor(llmProvider llm.LLMProvider) *K9CognitiveProcessor {
	kp := &K9CognitiveProcessor{
		llmProvider: llmProvider,
	}

	// Initialize phase descriptions (3x3 grid)
	kp.phaseDescriptions = [9]string{
		"Past-Universal: Historical patterns and universal principles",
		"Past-Particular: Specific memories and experiences",
		"Past-Relational: Connections between past events",
		"Present-Universal: Current universal truths",
		"Present-Particular: Current specific situation",
		"Present-Relational: Current relationships and dynamics",
		"Future-Universal: Potential universal outcomes",
		"Future-Particular: Specific anticipated results",
		"Future-Relational: Anticipated relationship changes",
	}

	return kp
}

// ProcessPhase processes a convolution phase using LLM
func (kp *K9CognitiveProcessor) ProcessPhase(ctx context.Context, phaseID int, input interface{}) (string, error) {
	row := phaseID / 3
	col := phaseID % 3

	timeAspect := []string{"Past", "Present", "Future"}[row]
	scopeAspect := []string{"Universal", "Particular", "Relational"}[col]

	systemPrompt := fmt.Sprintf(`You are processing phase %d in the K₉ triadic convolution system.
This is one of 9 orthogonal phases derived from 3² prime-power delegation.

Your phase represents: %s
Grid position: [%d, %d] (row: %s, column: %s)

The 9 phases form a 3x3 grid:
- Rows represent temporal aspects: Past, Present, Future
- Columns represent scope aspects: Universal, Particular, Relational

Your role is to analyze input through the lens of %s-%s cognition.`,
		phaseID, kp.phaseDescriptions[phaseID],
		row, col, timeAspect, scopeAspect,
		timeAspect, scopeAspect,
	)

	prompt := fmt.Sprintf("Analyze this input from your %s-%s perspective:\n%v", timeAspect, scopeAspect, input)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    200,
		Temperature:  0.7,
	}

	return kp.llmProvider.Generate(ctx, prompt, opts)
}

// =============================================================================
// φ (PHI) COGNITIVE PROCESSOR (LLM-POWERED)
// =============================================================================

// PhiCognitiveProcessor provides LLM-powered processing for delay fold
type PhiCognitiveProcessor struct {
	llmProvider llm.LLMProvider
	mu          sync.RWMutex
}

// NewPhiCognitiveProcessor creates a new Phi processor
func NewPhiCognitiveProcessor(llmProvider llm.LLMProvider) *PhiCognitiveProcessor {
	return &PhiCognitiveProcessor{
		llmProvider: llmProvider,
	}
}

// ProcessFold processes a delay fold state using LLM
func (pp *PhiCognitiveProcessor) ProcessFold(ctx context.Context, state DelayFoldState, dyadInput, triadInput interface{}) (string, error) {
	systemPrompt := fmt.Sprintf(`You are processing the φ (phi) delay fold operation in the sys6 architecture.
This compresses the 2×3=6 dyad-triad multiplex into 4 real steps.

Current state:
- Step: %d of 4
- State number: %d
- Dyad: %s (polarity)
- Triad: %d (phase)
- Dyad held: %v (using previous dyad value)
- Triad held: %v (using previous triad value)

The delay fold pattern alternates:
Step 1: State 1, Dyad A, Triad 1
Step 2: State 4, Dyad A, Triad 2 (dyad held)
Step 3: State 6, Dyad B, Triad 2 (triad held)
Step 4: State 1, Dyad B, Triad 3

Your role is to integrate the dyadic and triadic inputs according to the current fold state.`,
		state.Step, state.StateNum, state.Dyad, state.Triad, state.DyadHeld, state.TriadHeld,
	)

	var prompt string
	if state.DyadHeld {
		prompt = fmt.Sprintf("Dyad is HELD (using previous). Triad input: %v\nIntegrate with emphasis on triadic progression.", triadInput)
	} else if state.TriadHeld {
		prompt = fmt.Sprintf("Triad is HELD (using previous). Dyad input: %v\nIntegrate with emphasis on dyadic transition.", dyadInput)
	} else {
		prompt = fmt.Sprintf("Both inputs active.\nDyad input: %v\nTriad input: %v\nIntegrate both streams.", dyadInput, triadInput)
	}

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    200,
		Temperature:  0.7,
	}

	return pp.llmProvider.Generate(ctx, prompt, opts)
}

// =============================================================================
// INTROSPECTION COMMANDS FOR SYS6
// =============================================================================

// GetSys6IntrospectionCommands returns introspection command handlers for sys6
func (si *Sys6Integration) GetSys6IntrospectionCommands() map[string]func(args []string) string {
	return map[string]func(args []string) string{
		"/sys6":     si.cmdSys6Status,
		"/clock":    si.cmdClock,
		"/c8":       si.cmdC8Status,
		"/k9":       si.cmdK9Status,
		"/phi":      si.cmdPhiStatus,
		"/stages":   si.cmdStages,
		"/sync":     si.cmdSyncStatus,
	}
}

func (si *Sys6Integration) cmdSys6Status(args []string) string {
	status := si.GetStatus()
	
	var sb strings.Builder
	sb.WriteString("\n⚙️  SYS6 OPERAD STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString(fmt.Sprintf("Running:          %v\n", status.Running))
	sb.WriteString(fmt.Sprintf("Total Steps:      %d\n", status.Sys6Status.StepCount))
	sb.WriteString(fmt.Sprintf("Total Cycles:     %d\n", status.Sys6Status.TotalCycles))
	sb.WriteString(fmt.Sprintf("Total Syncs:      %d\n", status.Sys6Status.TotalSyncs))
	sb.WriteString(fmt.Sprintf("Avg Step Time:    %v\n", status.Sys6Status.AverageStepTime))
	sb.WriteString(fmt.Sprintf("Current Stage:    %s\n", status.Sys6Status.CurrentStage))
	sb.WriteString(fmt.Sprintf("Active C₈ States: %d/8\n", status.Sys6Status.ActiveC8States))
	sb.WriteString(fmt.Sprintf("Active K₉ Phases: %v\n", status.Sys6Status.ActiveK9Phases))
	
	return sb.String()
}

func (si *Sys6Integration) cmdClock(args []string) string {
	status := si.sys6.GetStatus()
	cs := status.ClockState
	
	var sb strings.Builder
	sb.WriteString("\n⏰ CLOCK30 (μ) STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString(fmt.Sprintf("Global Step:      %d / 30\n", cs.Step))
	sb.WriteString(fmt.Sprintf("Dyadic Phase:     %d (mod 2)\n", cs.DyadicPhase))
	sb.WriteString(fmt.Sprintf("Triadic Phase:    %d (mod 3)\n", cs.TriadicPhase))
	sb.WriteString(fmt.Sprintf("Pentadic Stage:   %d / 5\n", cs.PentadicStage))
	sb.WriteString(fmt.Sprintf("Four-Step Phase:  %d / 4\n", cs.FourStepPhase))
	
	// Visual representation
	sb.WriteString("\n30-Step Cycle:\n")
	for i := 1; i <= 30; i++ {
		if i == cs.Step {
			sb.WriteString("●")
		} else {
			sb.WriteString("○")
		}
		if i % 6 == 0 {
			sb.WriteString(fmt.Sprintf(" Stage %d\n", i/6))
		}
	}
	
	return sb.String()
}

func (si *Sys6Integration) cmdC8Status(args []string) string {
	status := si.sys6.GetStatus()
	
	var sb strings.Builder
	sb.WriteString("\n🧊 C₈ CUBIC CONCURRENCY STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString("8-way parallel states (2³ = 8):\n\n")
	
	// Show all 8 states with their binary codes
	for i := 0; i < 8; i++ {
		binary := fmt.Sprintf("%03b", i)
		active := "●"
		if i >= status.ActiveC8States {
			active = "○"
		}
		
		// Decode binary meaning
		pe := map[bool]string{true: "Expression", false: "Perception"}[binary[2] == '1']
		ar := map[bool]string{true: "Reflection", false: "Action"}[binary[1] == '1']
		li := map[bool]string{true: "Integration", false: "Learning"}[binary[0] == '1']
		
		sb.WriteString(fmt.Sprintf("%s State %d [%s]: %s | %s | %s\n", active, i, binary, pe, ar, li))
	}
	
	sb.WriteString("\nComplementary thread pairs:\n")
	sb.WriteString("  Thread 0: States 0 ↔ 7 (000 ↔ 111)\n")
	sb.WriteString("  Thread 1: States 1 ↔ 6 (001 ↔ 110)\n")
	sb.WriteString("  Thread 2: States 2 ↔ 5 (010 ↔ 101)\n")
	sb.WriteString("  Thread 3: States 3 ↔ 4 (011 ↔ 100)\n")
	
	return sb.String()
}

func (si *Sys6Integration) cmdK9Status(args []string) string {
	status := si.sys6.GetStatus()
	
	var sb strings.Builder
	sb.WriteString("\n🔮 K₉ TRIADIC CONVOLUTION STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString("9-phase kernel bank (3² = 9):\n\n")
	
	// Show 3x3 grid
	sb.WriteString("         Universal  Particular  Relational\n")
	sb.WriteString("        ┌──────────┬──────────┬──────────┐\n")
	
	rows := []string{"Past   ", "Present", "Future "}
	for row := 0; row < 3; row++ {
		sb.WriteString(fmt.Sprintf("%s │", rows[row]))
		for col := 0; col < 3; col++ {
			phaseID := row*3 + col
			active := "   ○    "
			for _, ap := range status.ActiveK9Phases {
				if ap == phaseID {
					active = "   ●    "
					break
				}
			}
			sb.WriteString(active)
			if col < 2 {
				sb.WriteString("│")
			}
		}
		sb.WriteString("│\n")
		if row < 2 {
			sb.WriteString("        ├──────────┼──────────┼──────────┤\n")
		}
	}
	sb.WriteString("        └──────────┴──────────┴──────────┘\n")
	
	sb.WriteString(fmt.Sprintf("\nActive phases: %v\n", status.ActiveK9Phases))
	sb.WriteString("3 rotation cores cycle through phases each step.\n")
	
	return sb.String()
}

func (si *Sys6Integration) cmdPhiStatus(args []string) string {
	status := si.sys6.GetStatus()
	ds := status.DelayFoldState
	
	var sb strings.Builder
	sb.WriteString("\n🔄 φ (PHI) DELAY FOLD STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString("Double-step delay pattern (2×3 → 4):\n\n")
	
	// Show the 4-step pattern
	sb.WriteString("Step │ State │ Dyad │ Triad │ Held\n")
	sb.WriteString("─────┼───────┼──────┼───────┼─────────────\n")
	
	patterns := []struct {
		step, state int
		dyad        string
		triad       int
		held        string
	}{
		{1, 1, "A", 1, "-"},
		{2, 4, "A", 2, "Dyad held"},
		{3, 6, "B", 2, "Triad held"},
		{4, 1, "B", 3, "-"},
	}
	
	for _, p := range patterns {
		marker := "  "
		if p.step == ds.Step {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%s%d  │   %d   │  %s   │   %d   │ %s\n",
			marker, p.step, p.state, p.dyad, p.triad, p.held))
	}
	
	sb.WriteString(fmt.Sprintf("\nCurrent: Step %d, Dyad %s, Triad %d\n", ds.Step, ds.Dyad, ds.Triad))
	
	return sb.String()
}

func (si *Sys6Integration) cmdStages(args []string) string {
	status := si.sys6.GetStatus()
	
	var sb strings.Builder
	sb.WriteString("\n📊 σ (SIGMA) STAGE SCHEDULER STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString("5 stages × 6 steps = 30-step cycle:\n\n")
	
	stages := []struct {
		name string
		desc string
	}{
		{"Perception", "Gather and process sensory input"},
		{"Analysis", "Analyze perceptions and identify patterns"},
		{"Planning", "Generate and evaluate action plans"},
		{"Execution", "Execute selected actions"},
		{"Integration", "Integrate results and update knowledge"},
	}
	
	currentStage := status.ClockState.PentadicStage
	
	for i, s := range stages {
		marker := "  "
		if i+1 == currentStage {
			marker = "→ "
		}
		sb.WriteString(fmt.Sprintf("%sStage %d: %s\n", marker, i+1, s.name))
		sb.WriteString(fmt.Sprintf("          %s\n", s.desc))
		
		// Show steps within stage
		sb.WriteString("          Steps: ")
		for j := 1; j <= 6; j++ {
			globalStep := i*6 + j
			if globalStep == status.ClockState.Step {
				sb.WriteString("●")
			} else {
				sb.WriteString("○")
			}
		}
		if i+1 == currentStage {
			sb.WriteString(" ← current")
		}
		sb.WriteString("\n\n")
	}
	
	return sb.String()
}

func (si *Sys6Integration) cmdSyncStatus(args []string) string {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	var sb strings.Builder
	sb.WriteString("\n🔗 SYS6 ↔ ECHOBEATS SYNC STATUS\n")
	sb.WriteString("══════════════════════════════════════════\n\n")
	
	sb.WriteString("Step mapping (30 sys6 → 12 echobeats):\n\n")
	
	// Show mapping in groups
	for eb := 1; eb <= 12; eb++ {
		sys6Steps := []int{}
		for s6 := 1; s6 <= 30; s6++ {
			if si.stateSync.stepMapping[s6] == eb {
				sys6Steps = append(sys6Steps, s6)
			}
		}
		sb.WriteString(fmt.Sprintf("  Echobeats %2d ← Sys6 %v\n", eb, sys6Steps))
	}
	
	sb.WriteString(fmt.Sprintf("\nSync events recorded: %d\n", len(si.stateSync.syncHistory)))
	
	// Show last 5 sync events
	if len(si.stateSync.syncHistory) > 0 {
		sb.WriteString("\nRecent sync events:\n")
		start := len(si.stateSync.syncHistory) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(si.stateSync.syncHistory); i++ {
			e := si.stateSync.syncHistory[i]
			sb.WriteString(fmt.Sprintf("  %s: %s\n", e.Timestamp.Format("15:04:05"), e.Description))
		}
	}
	
	return sb.String()
}
