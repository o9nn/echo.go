package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// =============================================================================
// SYS6 PAYLOAD PROCESSOR
// =============================================================================
//
// This module implements the payload processing pipeline that routes cognitive
// tokens and graph messages through the sys6 architecture components.
//
// Processing flow:
//   1. Payload enters at current Clock30 step
//   2. Routed through σ stage scheduler to appropriate stage
//   3. Processed by C₈ cubic concurrency (8-way parallel)
//   4. Processed by K₉ triadic convolution (9-phase)
//   5. Integrated by φ delay fold
//   6. Output collected and routed to next stage or completion
//
// =============================================================================

// Sys6PayloadProcessor processes payloads through the sys6 pipeline
type Sys6PayloadProcessor struct {
	mu sync.RWMutex

	// Core components
	sys6        *Sys6Operad
	llmProvider llm.LLMProvider
	factory     *PayloadFactory

	// Processing queues
	inputQueue    chan *PayloadEnvelope
	outputQueue   chan *PayloadEnvelope
	processingMap sync.Map // ID -> *PayloadEnvelope

	// Thread pool for C₈ parallel processing
	c8Workers [8]chan *PayloadEnvelope

	// K₉ convolution channels
	k9Channels [9]chan *PayloadEnvelope

	// Metrics
	metrics *ProcessorMetrics

	// Running state
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// ProcessorMetrics tracks processing statistics
type ProcessorMetrics struct {
	mu sync.RWMutex

	TokensProcessed  uint64
	GraphsProcessed  uint64
	TotalLatency     time.Duration
	AverageLatency   time.Duration
	ErrorCount       uint64
	C8Utilization    [8]float64
	K9Utilization    [9]float64
}

// NewSys6PayloadProcessor creates a new payload processor
func NewSys6PayloadProcessor(sys6 *Sys6Operad, llmProvider llm.LLMProvider) *Sys6PayloadProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	pp := &Sys6PayloadProcessor{
		sys6:        sys6,
		llmProvider: llmProvider,
		factory:     NewPayloadFactory(),
		inputQueue:  make(chan *PayloadEnvelope, 100),
		outputQueue: make(chan *PayloadEnvelope, 100),
		metrics:     &ProcessorMetrics{},
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize C₈ worker channels
	for i := 0; i < 8; i++ {
		pp.c8Workers[i] = make(chan *PayloadEnvelope, 10)
	}

	// Initialize K₉ channels
	for i := 0; i < 9; i++ {
		pp.k9Channels[i] = make(chan *PayloadEnvelope, 10)
	}

	return pp
}

// Start begins the payload processor
func (pp *Sys6PayloadProcessor) Start() error {
	pp.mu.Lock()
	if pp.running {
		pp.mu.Unlock()
		return fmt.Errorf("processor already running")
	}
	pp.running = true
	pp.mu.Unlock()

	// Start main processing loop
	go pp.runMainLoop()

	// Start C₈ workers
	for i := 0; i < 8; i++ {
		go pp.runC8Worker(i)
	}

	// Start K₉ workers
	for i := 0; i < 9; i++ {
		go pp.runK9Worker(i)
	}

	fmt.Println("📦 Sys6 Payload Processor: Started")
	return nil
}

// Stop halts the payload processor
func (pp *Sys6PayloadProcessor) Stop() error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if !pp.running {
		return fmt.Errorf("processor not running")
	}

	pp.cancel()
	pp.running = false

	fmt.Println("📦 Sys6 Payload Processor: Stopped")
	return nil
}

// Submit adds a payload to the processing queue
func (pp *Sys6PayloadProcessor) Submit(envelope *PayloadEnvelope) error {
	select {
	case pp.inputQueue <- envelope:
		return nil
	case <-pp.ctx.Done():
		return fmt.Errorf("processor stopped")
	default:
		return fmt.Errorf("input queue full")
	}
}

// SubmitToken creates and submits a token payload
func (pp *Sys6PayloadProcessor) SubmitToken(content string, contentType TokenType, source string, priority int) (*PayloadEnvelope, error) {
	token := pp.factory.CreateToken(content, contentType, source)
	envelope := pp.factory.CreateEnvelope(token, priority)
	if err := pp.Submit(envelope); err != nil {
		return nil, err
	}
	return envelope, nil
}

// SubmitGraph creates and submits a graph payload
func (pp *Sys6PayloadProcessor) SubmitGraph(graph *GraphMessage, priority int) (*PayloadEnvelope, error) {
	envelope := pp.factory.CreateEnvelope(graph, priority)
	if err := pp.Submit(envelope); err != nil {
		return nil, err
	}
	return envelope, nil
}

// GetOutput retrieves processed payloads
func (pp *Sys6PayloadProcessor) GetOutput() <-chan *PayloadEnvelope {
	return pp.outputQueue
}

// runMainLoop is the main processing loop
func (pp *Sys6PayloadProcessor) runMainLoop() {
	for {
		select {
		case <-pp.ctx.Done():
			return
		case envelope := <-pp.inputQueue:
			go pp.processEnvelope(envelope)
		}
	}
}

// processEnvelope processes a single payload envelope
func (pp *Sys6PayloadProcessor) processEnvelope(envelope *PayloadEnvelope) {
	startTime := time.Now()

	// Store in processing map
	pp.processingMap.Store(envelope.ID, envelope)
	defer pp.processingMap.Delete(envelope.ID)

	// Get current clock state
	status := pp.sys6.GetStatus()
	envelope.CurrentStep = status.ClockState.Step

	// Determine stage based on current step
	stage := pp.determineStage(envelope.CurrentStep)
	envelope.Route = append(envelope.Route, fmt.Sprintf("stage:%s", stage))

	// Process based on payload type
	var err error
	switch envelope.PayloadType {
	case PayloadTypeToken:
		err = pp.processToken(envelope)
	case PayloadTypeGraph:
		err = pp.processGraph(envelope)
	case PayloadTypeTokenBatch:
		err = pp.processTokenBatch(envelope)
	case PayloadTypeGraphBatch:
		err = pp.processGraphBatch(envelope)
	}

	if err != nil {
		envelope.Errors = append(envelope.Errors, PayloadError{
			Timestamp: time.Now(),
			Component: "processor",
			Step:      envelope.CurrentStep,
			Error:     err.Error(),
			Recovered: false,
		})
		pp.metrics.mu.Lock()
		pp.metrics.ErrorCount++
		pp.metrics.mu.Unlock()
	}

	// Mark completion
	envelope.ExitStep = pp.sys6.GetStatus().ClockState.Step
	envelope.ExitTime = time.Now()

	// Update metrics
	latency := time.Since(startTime)
	pp.metrics.mu.Lock()
	pp.metrics.TotalLatency += latency
	switch envelope.PayloadType {
	case PayloadTypeToken:
		pp.metrics.TokensProcessed++
	case PayloadTypeGraph:
		pp.metrics.GraphsProcessed++
	}
	total := pp.metrics.TokensProcessed + pp.metrics.GraphsProcessed
	if total > 0 {
		pp.metrics.AverageLatency = pp.metrics.TotalLatency / time.Duration(total)
	}
	pp.metrics.mu.Unlock()

	// Send to output
	select {
	case pp.outputQueue <- envelope:
	default:
		// Output queue full, log warning
	}
}

// determineStage returns the stage name for a given step
func (pp *Sys6PayloadProcessor) determineStage(step int) string {
	for _, stage := range SigmaStages {
		for _, s := range stage.Steps {
			if s == step {
				return stage.Name
			}
		}
	}
	return "unknown"
}

// processToken processes a single cognitive token
func (pp *Sys6PayloadProcessor) processToken(envelope *PayloadEnvelope) error {
	token := envelope.Token
	if token == nil {
		return fmt.Errorf("no token in envelope")
	}

	// Phase 1: C₈ Cubic Concurrency Processing
	if err := pp.processTokenC8(token); err != nil {
		return fmt.Errorf("C8 processing failed: %w", err)
	}
	token.PipelineState.C8Processed = true
	envelope.Route = append(envelope.Route, "c8:complete")

	// Phase 2: K₉ Triadic Convolution Processing
	if err := pp.processTokenK9(token); err != nil {
		return fmt.Errorf("K9 processing failed: %w", err)
	}
	token.PipelineState.K9Processed = true
	envelope.Route = append(envelope.Route, "k9:complete")

	// Phase 3: φ Delay Fold Integration
	if err := pp.processTokenPhi(token); err != nil {
		return fmt.Errorf("Phi processing failed: %w", err)
	}
	token.PipelineState.PhiProcessed = true
	envelope.Route = append(envelope.Route, "phi:complete")

	token.PipelineState.Completed = true
	return nil
}

// processTokenC8 processes a token through all 8 C₈ states
func (pp *Sys6PayloadProcessor) processTokenC8(token *CognitiveToken) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(stateID int) {
			defer wg.Done()

			perspective := C8Perspectives[stateID]
			result, err := pp.processC8State(token, stateID, perspective)

			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
				return
			}

			token.C8Results[stateID] = result
		}(i)
	}

	wg.Wait()

	// Record transformation
	token.Transformations = append(token.Transformations, TokenTransformation{
		Timestamp:   time.Now(),
		Component:   "c8",
		Operation:   "parallel_processing",
		Description: "Processed through 8 concurrent cognitive perspectives",
	})

	return firstErr
}

// processC8State processes a token through a single C₈ state
func (pp *Sys6PayloadProcessor) processC8State(token *CognitiveToken, stateID int, perspective C8Perspective) (*C8TokenResult, error) {
	systemPrompt := fmt.Sprintf(`You are processing cognitive content from the %s perspective.
Binary code: %s
Description: %s

Analyze the input and provide insights from this specific cognitive angle.
Focus on aspects relevant to your perspective's combination of:
- Perception vs Expression (input vs output orientation)
- Action vs Reflection (immediate vs contemplative processing)
- Learning vs Integration (acquiring vs consolidating knowledge)

Be concise but insightful.`, perspective.Name, perspective.Binary, perspective.Description)

	prompt := fmt.Sprintf("Content type: %s\nContent: %s\nSalience: %.2f\nRelevance: %.2f",
		token.ContentType, token.Content.Text, token.Salience, token.Relevance)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    150,
		Temperature:  0.7,
	}

	output, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	return &C8TokenResult{
		StateID:     stateID,
		BinaryCode:  perspective.Binary,
		Perspective: perspective.Name,
		Output:      output,
		Confidence:  0.7, // Could be derived from LLM response
		Metadata:    make(map[string]interface{}),
	}, nil
}

// processTokenK9 processes a token through all 9 K₉ phases
func (pp *Sys6PayloadProcessor) processTokenK9(token *CognitiveToken) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := 0; i < 9; i++ {
		wg.Add(1)
		go func(phaseID int) {
			defer wg.Done()

			phase := K9Phases[phaseID]
			result, err := pp.processK9Phase(token, phaseID, phase)

			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
				return
			}

			token.K9Results[phaseID] = result
		}(i)
	}

	wg.Wait()

	// Record transformation
	token.Transformations = append(token.Transformations, TokenTransformation{
		Timestamp:   time.Now(),
		Component:   "k9",
		Operation:   "triadic_convolution",
		Description: "Processed through 9 temporal-scope phases",
	})

	return firstErr
}

// processK9Phase processes a token through a single K₉ phase
func (pp *Sys6PayloadProcessor) processK9Phase(token *CognitiveToken, phaseID int, phase K9Phase) (*K9TokenResult, error) {
	systemPrompt := fmt.Sprintf(`You are analyzing cognitive content from the %s perspective.
Temporal aspect: %s
Scope aspect: %s
Description: %s

Analyze the input through this specific lens:
- %s: Consider the temporal dimension (%s)
- %s: Consider the scope dimension (%s)

Provide insights that emerge from this particular combination of time and scope.
Be concise but insightful.`,
		phase.Name, phase.Temporal, phase.Scope, phase.Description,
		phase.Temporal, map[string]string{
			"past":    "What historical patterns or precedents are relevant?",
			"present": "What is the current state and immediate context?",
			"future":  "What potential outcomes or trajectories exist?",
		}[phase.Temporal],
		phase.Scope, map[string]string{
			"universal":   "What general principles or laws apply?",
			"particular":  "What specific details or instances are relevant?",
			"relational":  "What connections or relationships are important?",
		}[phase.Scope])

	prompt := fmt.Sprintf("Content type: %s\nContent: %s",
		token.ContentType, token.Content.Text)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    150,
		Temperature:  0.7,
	}

	output, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	return &K9TokenResult{
		PhaseID:  phaseID,
		GridPos:  [2]int{phase.Row, phase.Col},
		Temporal: phase.Temporal,
		Scope:    phase.Scope,
		Output:   output,
		Weight:   1.0 / 9.0, // Equal weighting for now
		Metadata: make(map[string]interface{}),
	}, nil
}

// processTokenPhi integrates C₈ and K₉ results through the delay fold
func (pp *Sys6PayloadProcessor) processTokenPhi(token *CognitiveToken) error {
	// Get current delay fold state
	status := pp.sys6.GetStatus()
	ds := status.DelayFoldState

	// Collect C₈ outputs (dyadic stream)
	var c8Summary string
	for i, result := range token.C8Results {
		if result != nil {
			c8Summary += fmt.Sprintf("[%s]: %s\n", C8Perspectives[i].Name, result.Output)
		}
	}

	// Collect K₉ outputs (triadic stream)
	var k9Summary string
	for i, result := range token.K9Results {
		if result != nil {
			k9Summary += fmt.Sprintf("[%s]: %s\n", K9Phases[i].Name, result.Output)
		}
	}

	// Build integration prompt based on delay fold state
	var integrationFocus string
	if ds.DyadHeld {
		integrationFocus = "Focus on the triadic (temporal-scope) insights, as the dyadic stream is held."
	} else if ds.TriadHeld {
		integrationFocus = "Focus on the dyadic (perspective) insights, as the triadic stream is held."
	} else {
		integrationFocus = "Integrate both dyadic and triadic streams equally."
	}

	systemPrompt := fmt.Sprintf(`You are performing the φ (phi) delay fold integration.
Current state: Step %d, Dyad %s, Triad %d
%s

Synthesize the parallel processing results into a coherent integrated understanding.
Identify key insights, contradictions, and emergent patterns.`,
		ds.Step, ds.Dyad, ds.Triad, integrationFocus)

	prompt := fmt.Sprintf(`DYADIC (C₈) RESULTS:
%s

TRIADIC (K₉) RESULTS:
%s

Original content: %s

Provide an integrated synthesis.`, c8Summary, k9Summary, token.Content.Text)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    300,
		Temperature:  0.7,
	}

	integration, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return err
	}

	token.PhiResult = &PhiTokenResult{
		Step:        ds.Step,
		StateNum:    ds.StateNum,
		Dyad:        ds.Dyad,
		Triad:       ds.Triad,
		DyadHeld:    ds.DyadHeld,
		TriadHeld:   ds.TriadHeld,
		Integration: integration,
		Metadata:    make(map[string]interface{}),
	}

	// Record transformation
	token.Transformations = append(token.Transformations, TokenTransformation{
		Timestamp:   time.Now(),
		Component:   "phi",
		Operation:   "delay_fold_integration",
		Description: fmt.Sprintf("Integrated at step %d (Dyad %s, Triad %d)", ds.Step, ds.Dyad, ds.Triad),
	})

	return nil
}

// processGraph processes a graph message
func (pp *Sys6PayloadProcessor) processGraph(envelope *PayloadEnvelope) error {
	graph := envelope.Graph
	if graph == nil {
		return fmt.Errorf("no graph in envelope")
	}

	// Phase 1: C₈ Graph Processing
	if err := pp.processGraphC8(graph); err != nil {
		return fmt.Errorf("C8 graph processing failed: %w", err)
	}
	graph.PipelineState.C8Processed = true
	envelope.Route = append(envelope.Route, "c8:complete")

	// Phase 2: K₉ Graph Processing
	if err := pp.processGraphK9(graph); err != nil {
		return fmt.Errorf("K9 graph processing failed: %w", err)
	}
	graph.PipelineState.K9Processed = true
	envelope.Route = append(envelope.Route, "k9:complete")

	// Phase 3: φ Graph Integration
	if err := pp.processGraphPhi(graph); err != nil {
		return fmt.Errorf("Phi graph processing failed: %w", err)
	}
	graph.PipelineState.PhiProcessed = true
	envelope.Route = append(envelope.Route, "phi:complete")

	graph.PipelineState.Completed = true
	return nil
}

// processGraphC8 processes a graph through C₈ states
func (pp *Sys6PayloadProcessor) processGraphC8(graph *GraphMessage) error {
	// Serialize graph for LLM processing
	graphDesc := pp.describeGraph(graph)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(stateID int) {
			defer wg.Done()

			perspective := C8Perspectives[stateID]
			result, err := pp.processGraphC8State(graph, graphDesc, stateID, perspective)

			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
				return
			}

			graph.C8GraphResults[stateID] = result
		}(i)
	}

	wg.Wait()
	return firstErr
}

// processGraphC8State processes a graph through a single C₈ state
func (pp *Sys6PayloadProcessor) processGraphC8State(graph *GraphMessage, graphDesc string, stateID int, perspective C8Perspective) (*C8GraphResult, error) {
	systemPrompt := fmt.Sprintf(`You are analyzing a cognitive graph from the %s perspective.
Binary code: %s
Description: %s

Analyze the graph structure and suggest modifications or insights from this perspective.
Consider how nodes and edges relate to your cognitive angle.`,
		perspective.Name, perspective.Binary, perspective.Description)

	prompt := fmt.Sprintf("Graph type: %s\n%s", graph.MessageType, graphDesc)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    200,
		Temperature:  0.7,
	}

	output, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	return &C8GraphResult{
		StateID:     stateID,
		Perspective: perspective.Name,
		Analysis:    output,
		Metadata:    make(map[string]interface{}),
	}, nil
}

// processGraphK9 processes a graph through K₉ phases
func (pp *Sys6PayloadProcessor) processGraphK9(graph *GraphMessage) error {
	graphDesc := pp.describeGraph(graph)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := 0; i < 9; i++ {
		wg.Add(1)
		go func(phaseID int) {
			defer wg.Done()

			phase := K9Phases[phaseID]
			result, err := pp.processGraphK9Phase(graph, graphDesc, phaseID, phase)

			mu.Lock()
			defer mu.Unlock()

			if err != nil && firstErr == nil {
				firstErr = err
				return
			}

			graph.K9GraphResults[phaseID] = result
		}(i)
	}

	wg.Wait()
	return firstErr
}

// processGraphK9Phase processes a graph through a single K₉ phase
func (pp *Sys6PayloadProcessor) processGraphK9Phase(graph *GraphMessage, graphDesc string, phaseID int, phase K9Phase) (*K9GraphResult, error) {
	systemPrompt := fmt.Sprintf(`You are analyzing a cognitive graph from the %s perspective.
Temporal: %s, Scope: %s
Description: %s

Extract or transform the graph according to this temporal-scope lens.`,
		phase.Name, phase.Temporal, phase.Scope, phase.Description)

	prompt := fmt.Sprintf("Graph type: %s\n%s", graph.MessageType, graphDesc)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    200,
		Temperature:  0.7,
	}

	output, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	return &K9GraphResult{
		PhaseID:  phaseID,
		Temporal: phase.Temporal,
		Scope:    phase.Scope,
		Analysis: output,
		Metadata: make(map[string]interface{}),
	}, nil
}

// processGraphPhi integrates graph processing results
func (pp *Sys6PayloadProcessor) processGraphPhi(graph *GraphMessage) error {
	status := pp.sys6.GetStatus()
	ds := status.DelayFoldState

	// Collect summaries
	var c8Summary, k9Summary string
	for i, result := range graph.C8GraphResults {
		if result != nil {
			c8Summary += fmt.Sprintf("[%s]: %s\n", C8Perspectives[i].Name, result.Analysis)
		}
	}
	for i, result := range graph.K9GraphResults {
		if result != nil {
			k9Summary += fmt.Sprintf("[%s]: %s\n", K9Phases[i].Name, result.Analysis)
		}
	}

	systemPrompt := fmt.Sprintf(`You are performing φ delay fold integration on a cognitive graph.
Current state: Step %d, Dyad %s, Triad %d

Synthesize the parallel graph analyses into recommendations for graph transformation.`,
		ds.Step, ds.Dyad, ds.Triad)

	prompt := fmt.Sprintf(`C₈ ANALYSES:\n%s\n\nK₉ ANALYSES:\n%s\n\nProvide integrated graph transformation recommendations.`,
		c8Summary, k9Summary)

	opts := llm.GenerateOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    300,
		Temperature:  0.7,
	}

	integration, err := pp.llmProvider.Generate(pp.ctx, prompt, opts)
	if err != nil {
		return err
	}

	graph.PhiGraphResult = &PhiGraphResult{
		Step:     ds.Step,
		Metadata: map[string]interface{}{"integration": integration},
	}

	return nil
}

// processTokenBatch processes multiple tokens
func (pp *Sys6PayloadProcessor) processTokenBatch(envelope *PayloadEnvelope) error {
	for _, token := range envelope.TokenBatch {
		tokenEnv := pp.factory.CreateEnvelope(token, envelope.Priority)
		if err := pp.processToken(tokenEnv); err != nil {
			return err
		}
	}
	return nil
}

// processGraphBatch processes multiple graphs
func (pp *Sys6PayloadProcessor) processGraphBatch(envelope *PayloadEnvelope) error {
	for _, graph := range envelope.GraphBatch {
		graphEnv := pp.factory.CreateEnvelope(graph, envelope.Priority)
		if err := pp.processGraph(graphEnv); err != nil {
			return err
		}
	}
	return nil
}

// describeGraph creates a text description of a graph
func (pp *Sys6PayloadProcessor) describeGraph(graph *GraphMessage) string {
	desc := fmt.Sprintf("Nodes (%d):\n", len(graph.Nodes))
	for _, node := range graph.Nodes {
		desc += fmt.Sprintf("  - %s (%s): %s\n", node.ID, node.NodeType, node.Label)
	}
	desc += fmt.Sprintf("\nEdges (%d):\n", len(graph.Edges))
	for _, edge := range graph.Edges {
		desc += fmt.Sprintf("  - %s -[%s]-> %s\n", edge.SourceID, edge.EdgeType, edge.TargetID)
	}
	return desc
}

// runC8Worker runs a C₈ worker goroutine
func (pp *Sys6PayloadProcessor) runC8Worker(stateID int) {
	for {
		select {
		case <-pp.ctx.Done():
			return
		case envelope := <-pp.c8Workers[stateID]:
			// Process envelope through this C₈ state
			if envelope.Token != nil {
				perspective := C8Perspectives[stateID]
				result, _ := pp.processC8State(envelope.Token, stateID, perspective)
				envelope.Token.C8Results[stateID] = result
			}
		}
	}
}

// runK9Worker runs a K₉ worker goroutine
func (pp *Sys6PayloadProcessor) runK9Worker(phaseID int) {
	for {
		select {
		case <-pp.ctx.Done():
			return
		case envelope := <-pp.k9Channels[phaseID]:
			// Process envelope through this K₉ phase
			if envelope.Token != nil {
				phase := K9Phases[phaseID]
				result, _ := pp.processK9Phase(envelope.Token, phaseID, phase)
				envelope.Token.K9Results[phaseID] = result
			}
		}
	}
}

// GetMetrics returns current processor metrics
func (pp *Sys6PayloadProcessor) GetMetrics() ProcessorMetrics {
	pp.metrics.mu.RLock()
	defer pp.metrics.mu.RUnlock()
	return *pp.metrics
}
