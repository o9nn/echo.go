package deeptreeecho

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/o9nn/echo.go/core/backendcap"
	"github.com/o9nn/echo.go/core/llm"
)

// EvolutionSystem is the main entry point for the Deep Tree Echo evolution system
// It manages real LLM provider connections and coordinates all subsystems
type EvolutionSystem struct {
	mu sync.RWMutex

	// Core components
	optimizer       *EvolutionOptimizer
	providerManager *llm.ProviderManager
	eventBus        *CognitiveEventBus
	lastBackend     backendcap.Decision
	hasLastBackend  bool

	// Configuration
	config EvolutionSystemConfig

	// State
	initialized bool
	running     bool
}

// EvolutionSystemConfig configures the evolution system
type EvolutionSystemConfig struct {
	// LLM provider preferences (in order of preference)
	PreferredProviders []string

	// EventBus receives backend capability and routing events. If nil, a local bus is created.
	EventBus *CognitiveEventBus

	// Evolution configuration
	EvolutionConfig EvolutionConfig

	// Enable debug output
	Debug bool
}

// DefaultEvolutionSystemConfig returns default configuration
func DefaultEvolutionSystemConfig() EvolutionSystemConfig {
	return EvolutionSystemConfig{
		PreferredProviders: []string{"anthropic", "openrouter", "openai"},
		EvolutionConfig:    DefaultEvolutionConfig(),
		Debug:              false,
	}
}

// NewEvolutionSystem creates a new evolution system with real LLM integration
func NewEvolutionSystem(config EvolutionSystemConfig) (*EvolutionSystem, error) {
	es := &EvolutionSystem{
		config:   config,
		eventBus: config.EventBus,
	}
	if es.eventBus == nil {
		es.eventBus = NewCognitiveEventBus(context.Background())
	}

	// Initialize provider manager
	pm := llm.NewProviderManager()

	// Register available providers based on environment
	if err := es.registerProviders(pm); err != nil {
		return nil, fmt.Errorf("failed to register providers: %w", err)
	}

	es.providerManager = pm
	es.refreshBackendRouting("initialization")

	// Check if any provider is available
	if !pm.Available() {
		return nil, fmt.Errorf("no LLM providers available - set ANTHROPIC_API_KEY, OPENROUTER_API_KEY, or OPENAI_API_KEY")
	}

	// Create evolution optimizer with real provider
	es.optimizer = NewEvolutionOptimizer(pm, config.EvolutionConfig)

	es.initialized = true

	if config.Debug {
		fmt.Printf("🧬 Evolution System initialized with providers: %v\n", pm.ListProviders())
	}

	return es, nil
}

// registerProviders registers available LLM providers
func (es *EvolutionSystem) registerProviders(pm *llm.ProviderManager) error {
	registered := 0

	// Try Anthropic
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		provider := llm.NewAnthropicProvider("claude-3-sonnet-20240229")
		if err := pm.RegisterProvider(provider); err == nil {
			registered++
			if es.config.Debug {
				fmt.Println("   ✓ Registered Anthropic provider")
			}
		}
	}

	// Try OpenRouter
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		provider := llm.NewOpenRouterProvider("")
		if err := pm.RegisterProvider(provider); err == nil {
			registered++
			if es.config.Debug {
				fmt.Println("   ✓ Registered OpenRouter provider")
			}
		}
	}

	// Try OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		provider := llm.NewOpenAIProvider(apiKey)
		if err := pm.RegisterProvider(provider); err == nil {
			registered++
			if es.config.Debug {
				fmt.Println("   ✓ Registered OpenAI provider")
			}
		}
	}

	// Always keep the simple continuity fallback registered so autonomous loops can degrade gracefully.
	if err := pm.RegisterProvider(&llm.SimpleFallbackProvider{}); err == nil {
		registered++
		if es.config.Debug {
			fmt.Println("   ✓ Registered SimpleFallback provider")
		}
	}

	// Set fallback chain from explicit preferences, then append any remaining providers.
	available := make(map[string]bool)
	for _, name := range pm.ListProviders() {
		available[name] = true
	}
	chain := make([]string, 0, len(available))
	for _, preferred := range es.config.PreferredProviders {
		if available[preferred] {
			chain = append(chain, preferred)
			delete(available, preferred)
		}
	}
	if available["SimpleFallback"] {
		delete(available, "SimpleFallback")
	}
	for name := range available {
		chain = append(chain, name)
	}
	chain = append(chain, "SimpleFallback")
	if len(chain) > 0 {
		pm.SetFallbackChain(chain)
	}

	if registered == 0 {
		return fmt.Errorf("no providers could be registered")
	}

	return nil
}

// Start begins the evolution system
func (es *EvolutionSystem) Start(ctx context.Context) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.initialized {
		return fmt.Errorf("evolution system not initialized")
	}

	if es.running {
		return fmt.Errorf("evolution system already running")
	}

	if err := es.optimizer.Start(); err != nil {
		return fmt.Errorf("failed to start optimizer: %w", err)
	}

	es.running = true

	if es.config.Debug {
		fmt.Println("🧬 Evolution System started")
	}

	return nil
}

// Stop gracefully stops the evolution system
func (es *EvolutionSystem) Stop() error {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.running {
		return fmt.Errorf("evolution system not running")
	}

	if err := es.optimizer.Stop(); err != nil {
		return fmt.Errorf("failed to stop optimizer: %w", err)
	}

	es.running = false

	if es.config.Debug {
		fmt.Println("🧬 Evolution System stopped")
	}

	return nil
}

// GetOptimizer returns the evolution optimizer
func (es *EvolutionSystem) GetOptimizer() *EvolutionOptimizer {
	return es.optimizer
}

// GetProviderManager returns the LLM provider manager
func (es *EvolutionSystem) GetProviderManager() *llm.ProviderManager {
	return es.providerManager
}

// GetStatus returns the current system status
func (es *EvolutionSystem) GetStatus() map[string]interface{} {
	es.mu.RLock()
	defer es.mu.RUnlock()

	backendDecision := es.BackendDecision()
	status := map[string]interface{}{
		"initialized":          es.initialized,
		"running":              es.running,
		"providers":            es.providerManager.ListProviders(),
		"provider_metrics":     es.providerManager.GetMetrics(),
		"backend_capabilities": backendcap.Snapshot(),
		"backend_decision":     backendDecision,
		"backend_degraded":     backendDecision.Degraded,
	}

	if es.optimizer != nil {
		status["evolution"] = es.optimizer.GetEvolutionStatus()
		status["genetic_traits"] = es.optimizer.GetGeneticTraits()
	}

	return status
}

// BackendDecision returns the scheduler-facing backend selection for evolution inference work.
func (es *EvolutionSystem) BackendDecision() backendcap.Decision {
	workload := backendcap.Workload{
		NeedOffline:   false,
		PreferNative:  true,
		MinMemoryTier: backendcap.MemoryConstrained,
	}
	return backendcap.Select(workload)
}

// InjectStimulus injects an external stimulus into the evolution system
func (es *EvolutionSystem) InjectStimulus(stimulus string, importance float64) {
	if es.optimizer != nil {
		es.optimizer.InjectExternalStimulus(stimulus, importance)
	}
}

// Generate produces a completion using the LLM provider
func (es *EvolutionSystem) Generate(ctx context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	es.refreshBackendRouting("generation")
	return es.providerManager.Generate(ctx, prompt, opts)
}

// GetEventBus returns the event stream used by the evolution system.
func (es *EvolutionSystem) GetEventBus() *CognitiveEventBus {
	return es.eventBus
}

func (es *EvolutionSystem) refreshBackendRouting(reason string) []string {
	if es.providerManager == nil {
		return nil
	}
	decision := es.BackendDecision()
	providerRoute := es.providerManager.ApplyBackendDecision(decision)
	if es.backendDecisionChanged(decision) {
		es.emitBackendCapabilityEvent(decision, providerRoute, reason)
		es.mu.Lock()
		es.lastBackend = decision
		es.hasLastBackend = true
		es.mu.Unlock()
	}
	return providerRoute
}

func (es *EvolutionSystem) backendDecisionChanged(decision backendcap.Decision) bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	if !es.hasLastBackend {
		return true
	}
	return es.lastBackend.Selected.Name != decision.Selected.Name ||
		es.lastBackend.Selected.Kind != decision.Selected.Kind ||
		es.lastBackend.Degraded != decision.Degraded ||
		es.lastBackend.Selected.MemoryTier != decision.Selected.MemoryTier
}

func (es *EvolutionSystem) emitBackendCapabilityEvent(decision backendcap.Decision, providerRoute []string, reason string) {
	if es.eventBus == nil {
		return
	}
	es.eventBus.PublishSync(NewPriorityCognitiveEvent(EventBackendCapabilityChanged, "evolution_system", map[string]interface{}{
		"reason":                reason,
		"decision":              decision,
		"snapshot":              backendcap.Snapshot(),
		"host_memory":           backendcap.ProbeHostMemory(),
		"provider_route":        providerRoute,
		"selected_backend":      decision.Selected.Name,
		"selected_backend_kind": decision.Selected.Kind,
		"degraded":              decision.Degraded,
	}, 0.8))
}

// GetConsciousnessMetrics returns stream of consciousness metrics
func (es *EvolutionSystem) GetConsciousnessMetrics() map[string]interface{} {
	if es.optimizer != nil && es.optimizer.consciousness != nil {
		return es.optimizer.consciousness.GetMetrics()
	}
	return nil
}

// GetSchedulerMetrics returns echobeats scheduler metrics
func (es *EvolutionSystem) GetSchedulerMetrics() map[string]interface{} {
	if es.optimizer != nil && es.optimizer.scheduler != nil {
		return es.optimizer.scheduler.GetMetrics()
	}
	return nil
}

// GetDreamMetrics returns echodream integration metrics
func (es *EvolutionSystem) GetDreamMetrics() map[string]interface{} {
	if es.optimizer != nil && es.optimizer.dreamIntegration != nil {
		return es.optimizer.dreamIntegration.GetMetrics()
	}
	return nil
}

// AddGoal adds a goal to the scheduler
func (es *EvolutionSystem) AddGoal(description string, priority float64) string {
	if es.optimizer != nil && es.optimizer.scheduler != nil {
		return es.optimizer.scheduler.AddGoal(description, priority)
	}
	return ""
}

// TriggerDreamCycle triggers a dream consolidation cycle
func (es *EvolutionSystem) TriggerDreamCycle() error {
	if es.optimizer != nil && es.optimizer.dreamIntegration != nil {
		return es.optimizer.dreamIntegration.StartDreamCycle()
	}
	return fmt.Errorf("dream integration not available")
}

// GetEvolutionHistory returns the evolution history
func (es *EvolutionSystem) GetEvolutionHistory() []EvolutionSnapshot {
	if es.optimizer != nil {
		return es.optimizer.GetEvolutionHistory()
	}
	return nil
}

// RunDiagnostics runs comprehensive system diagnostics
func (es *EvolutionSystem) RunDiagnostics(ctx context.Context) (*EvolutionDiagnostics, error) {
	diag := &EvolutionDiagnostics{
		Tests: make([]DiagnosticTestResult, 0),
	}

	// Test LLM provider
	providerTest := DiagnosticTestResult{Name: "LLM Provider"}
	if es.providerManager.Available() {
		// Try a simple generation
		_, err := es.providerManager.Generate(ctx, "Test", llm.GenerateOptions{MaxTokens: 10})
		if err != nil {
			providerTest.Status = "warn"
			providerTest.Message = fmt.Sprintf("Provider available but generation failed: %v", err)
		} else {
			providerTest.Status = "pass"
			providerTest.Message = "LLM provider responding correctly"
		}
	} else {
		providerTest.Status = "fail"
		providerTest.Message = "No LLM provider available"
	}
	diag.Tests = append(diag.Tests, providerTest)

	// Test consciousness
	consTest := DiagnosticTestResult{Name: "Stream of Consciousness"}
	if es.optimizer != nil && es.optimizer.consciousness != nil {
		consTest.Status = "pass"
		consTest.Message = "Consciousness subsystem initialized"
	} else {
		consTest.Status = "fail"
		consTest.Message = "Consciousness subsystem not available"
	}
	diag.Tests = append(diag.Tests, consTest)

	// Test scheduler
	schedTest := DiagnosticTestResult{Name: "Echobeats Scheduler"}
	if es.optimizer != nil && es.optimizer.scheduler != nil {
		triads := es.optimizer.scheduler.GetTriadStates()
		if len(triads) == 4 {
			schedTest.Status = "pass"
			schedTest.Message = "Scheduler with 4 triads operational"
		} else {
			schedTest.Status = "warn"
			schedTest.Message = fmt.Sprintf("Scheduler has %d triads (expected 4)", len(triads))
		}
	} else {
		schedTest.Status = "fail"
		schedTest.Message = "Scheduler subsystem not available"
	}
	diag.Tests = append(diag.Tests, schedTest)

	// Test dream integration
	dreamTest := DiagnosticTestResult{Name: "Echodream Integration"}
	if es.optimizer != nil && es.optimizer.dreamIntegration != nil {
		dreamTest.Status = "pass"
		dreamTest.Message = fmt.Sprintf("Dream phase: %s", es.optimizer.dreamIntegration.GetDreamPhase().String())
	} else {
		dreamTest.Status = "fail"
		dreamTest.Message = "Dream integration not available"
	}
	diag.Tests = append(diag.Tests, dreamTest)

	// Test genetic traits
	traitTest := DiagnosticTestResult{Name: "Genetic Traits"}
	if es.optimizer != nil {
		traits := es.optimizer.GetGeneticTraits()
		if len(traits) == 5 {
			traitTest.Status = "pass"
			traitTest.Message = "All 5 genetic traits initialized"
		} else {
			traitTest.Status = "warn"
			traitTest.Message = fmt.Sprintf("%d traits initialized (expected 5)", len(traits))
		}
	} else {
		traitTest.Status = "fail"
		traitTest.Message = "Optimizer not available"
	}
	diag.Tests = append(diag.Tests, traitTest)

	// Calculate overall health
	passCount := 0
	failCount := 0
	for _, test := range diag.Tests {
		switch test.Status {
		case "pass":
			passCount++
		case "fail":
			failCount++
		}
	}

	if failCount > 0 {
		diag.OverallHealth = "degraded"
	} else if passCount == len(diag.Tests) {
		diag.OverallHealth = "optimal"
	} else {
		diag.OverallHealth = "stable"
	}

	return diag, nil
}

// EvolutionDiagnostics contains diagnostic results
type EvolutionDiagnostics struct {
	Tests         []DiagnosticTestResult `json:"tests"`
	OverallHealth string                 `json:"overall_health"`
}

// DiagnosticTestResult represents a single diagnostic test
type DiagnosticTestResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
