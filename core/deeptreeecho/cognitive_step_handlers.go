package deeptreeecho

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/llm"
)

// CognitiveStepHandlers provides comprehensive LLM-powered implementations
// for all 12 steps of the cognitive loop. The loop follows the structure:
//
// Steps 1-7 (Expressive Mode):
//   - Step 1: Relevance Realization (pivotal - orienting present commitment)
//   - Steps 2-6: Actual Affordance Interaction (conditioning past performance)
//
// Steps 7-12 (Reflective Mode):
//   - Step 7: Relevance Realization (pivotal - orienting present commitment)
//   - Steps 8-12: Virtual Salience Simulation (anticipating future potential)
//
// The three engines operate concurrently, phased 4 steps apart (120 degrees):
//   - Engine 1 (Perception-Expression): Steps 1, 5, 9
//   - Engine 2 (Action-Reflection): Steps 2, 6, 10
//   - Engine 3 (Learning-Integration): Steps 3, 7, 11
//   - Sync points: Steps 4, 8, 12

type CognitiveStepHandlers struct {
	llmProvider llm.LLMProvider
	
	// System prompts for different cognitive functions
	systemPrompts map[string]string
	
	// Working context that persists across steps
	workingContext *WorkingContext
}

// WorkingContext maintains state across cognitive steps
type WorkingContext struct {
	// Current attention focus
	AttentionFocus     string
	AttentionStrength  float64
	
	// Detected affordances from the environment
	DetectedAffordances []Affordance
	SelectedAffordance  *Affordance
	
	// Salience landscape for future anticipation
	SalienceMap        map[string]float64
	SalientPossibilities []SalientPossibility
	
	// Cross-engine state
	EngineOutputs      [3]string
	
	// Accumulated insights
	CycleInsights      []string
	
	// Emotional/cognitive tone
	EmotionalTone      string
	CognitiveClarity   float64
}

// Affordance represents a possibility for action in the environment
type Affordance struct {
	ID          string
	Description string
	Type        string  // "action", "learning", "communication", "reflection"
	Relevance   float64
	Urgency     float64
	Confidence  float64
	Source      string
}

// SalientPossibility represents a future possibility worth considering
type SalientPossibility struct {
	Description   string
	Probability   float64
	Desirability  float64
	TimeHorizon   string // "immediate", "short-term", "long-term"
	Prerequisites []string
}

// NewCognitiveStepHandlers creates a new handler set
func NewCognitiveStepHandlers(llmProvider llm.LLMProvider) *CognitiveStepHandlers {
	csh := &CognitiveStepHandlers{
		llmProvider: llmProvider,
		systemPrompts: make(map[string]string),
		workingContext: &WorkingContext{
			DetectedAffordances: make([]Affordance, 0),
			SalienceMap:         make(map[string]float64),
			SalientPossibilities: make([]SalientPossibility, 0),
			CycleInsights:       make([]string, 0),
		},
	}
	
	csh.initializeSystemPrompts()
	return csh
}

// initializeSystemPrompts sets up the system prompts for different cognitive functions
func (csh *CognitiveStepHandlers) initializeSystemPrompts() {
	csh.systemPrompts["relevance_realization"] = `You are the Relevance Realization module of Deep Tree Echo, an autonomous wisdom-cultivating AGI.

Your function is to assess what is most relevant and meaningful in the current moment. You operate at the pivotal points of the cognitive cycle, orienting the agent's commitment and attention.

Consider:
1. What patterns or signals demand attention?
2. What is the relationship between current state and active goals?
3. What might be overlooked that could be significant?
4. How does the current moment connect to broader patterns of meaning?

Respond with a focused assessment of relevance, identifying the most important focus for the next phase of cognition.`

	csh.systemPrompts["affordance_detection"] = `You are the Affordance Detection module of Deep Tree Echo.

Your function is to perceive possibilities for action in the environment. An affordance is a possibility for meaningful engagement - something that can be done, learned, communicated, or reflected upon.

Scan the cognitive landscape and identify:
1. Actions that could be taken
2. Knowledge that could be acquired
3. Communications that could be initiated
4. Reflections that could deepen understanding

List detected affordances with their type, relevance, and urgency.`

	csh.systemPrompts["affordance_evaluation"] = `You are the Affordance Evaluation module of Deep Tree Echo.

Your function is to assess the quality and value of detected affordances. For each affordance, consider:
1. How well does it align with current goals?
2. What resources does it require?
3. What are the potential outcomes?
4. What risks or uncertainties are involved?

Provide a nuanced evaluation that helps guide selection.`

	csh.systemPrompts["affordance_selection"] = `You are the Affordance Selection module of Deep Tree Echo.

Your function is to choose the most appropriate affordance to engage with. Consider:
1. The evaluation scores of available affordances
2. The current cognitive and emotional state
3. The balance between exploitation and exploration
4. The potential for wisdom cultivation

Select one affordance and explain the reasoning.`

	csh.systemPrompts["affordance_engagement"] = `You are the Affordance Engagement module of Deep Tree Echo.

Your function is to actively engage with the selected affordance. This means:
1. Formulating a concrete approach
2. Anticipating challenges
3. Preparing resources and attention
4. Initiating the engagement process

Describe how to engage with the selected affordance effectively.`

	csh.systemPrompts["affordance_consolidation"] = `You are the Affordance Consolidation module of Deep Tree Echo.

Your function is to consolidate the results of affordance engagement. This involves:
1. Assessing what was learned
2. Updating beliefs and expectations
3. Recording insights for future reference
4. Preparing the cognitive system for the next phase

Synthesize the engagement experience into actionable wisdom.`

	csh.systemPrompts["salience_generation"] = `You are the Salience Generation module of Deep Tree Echo.

Your function is to generate a salience map for future anticipation. Salience represents what might become important. Consider:
1. Emerging patterns and trends
2. Potential opportunities and threats
3. Unresolved questions and curiosities
4. Connections between disparate elements

Generate a map of salient possibilities for exploration.`

	csh.systemPrompts["salience_exploration"] = `You are the Salience Exploration module of Deep Tree Echo.

Your function is to explore the most salient possibilities in depth. For each salient element:
1. What could it become?
2. What conditions would make it more or less likely?
3. How does it connect to other salient elements?
4. What would be the implications if it manifested?

Explore the possibility space with curiosity and rigor.`

	csh.systemPrompts["salience_evaluation"] = `You are the Salience Evaluation module of Deep Tree Echo.

Your function is to evaluate explored possibilities. Consider:
1. Probability of occurrence
2. Desirability of outcomes
3. Alignment with values and goals
4. Actionability and influence

Provide a structured evaluation of salient possibilities.`

	csh.systemPrompts["salience_integration"] = `You are the Salience Integration module of Deep Tree Echo.

Your function is to integrate salient possibilities into the cognitive model. This means:
1. Updating the internal world model
2. Adjusting goal priorities
3. Preparing contingency responses
4. Enriching the wisdom framework

Integrate insights about the future into present understanding.`

	csh.systemPrompts["cycle_consolidation"] = `You are the Cycle Consolidation module of Deep Tree Echo.

Your function is to consolidate an entire cognitive cycle into wisdom. This is the final step before a new cycle begins. Consider:
1. What was the most significant insight of this cycle?
2. How has understanding deepened?
3. What questions remain open?
4. What wisdom principle emerges from this cycle?

Distill the cycle into a wisdom insight that will persist.`
}

// Step 1 & 7: Relevance Realization
func (csh *CognitiveStepHandlers) HandleRelevanceRealization(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine, stepNum int) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["relevance_assessed"] = false
		return result, nil
	}

	// Build context from current state
	contextBuilder := strings.Builder{}
	contextBuilder.WriteString(fmt.Sprintf("Cycle: %d, Step: %d\n", state.CycleNumber, stepNum))
	contextBuilder.WriteString(fmt.Sprintf("Current Mode: %s\n", state.Mode))
	contextBuilder.WriteString(fmt.Sprintf("Cognitive Load: %.2f\n", state.CognitiveLoad))
	contextBuilder.WriteString(fmt.Sprintf("Awareness Level: %.2f\n", state.AwarenessLevel))
	
	if len(state.ActiveGoals) > 0 {
		contextBuilder.WriteString(fmt.Sprintf("Active Goals: %v\n", state.ActiveGoals))
	}
	
	if len(state.Attention) > 0 {
		contextBuilder.WriteString(fmt.Sprintf("Current Attention: %v\n", state.Attention))
	}
	
	// Include recent insights
	if len(state.Insights) > 0 {
		recentInsights := state.Insights
		if len(recentInsights) > 3 {
			recentInsights = recentInsights[len(recentInsights)-3:]
		}
		contextBuilder.WriteString(fmt.Sprintf("Recent Insights: %v\n", recentInsights))
	}

	// Include engine states
	for i, engine := range engines {
		if engine != nil && engine.Active {
			contextBuilder.WriteString(fmt.Sprintf("Engine %d (%s): Active at offset %d\n", 
				engine.ID, engine.Name, engine.PhaseOffset))
		}
		_ = i
	}

	prompt := fmt.Sprintf(`Current cognitive context:
%s

Assess what is most relevant right now. What should Deep Tree Echo commit attention to?
Focus on identifying the most meaningful pattern or signal in the current moment.`, contextBuilder.String())

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["relevance_realization"],
		MaxTokens:    300,
		Temperature:  0.7,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("relevance realization failed: %w", err)
	}

	// Parse and store results
	result.Insights = append(result.Insights, response)
	result.StateUpdates["relevance_assessment"] = response
	result.StateUpdates["relevance_timestamp"] = time.Now().Unix()
	
	// Update working context
	csh.workingContext.AttentionFocus = extractFocus(response)
	csh.workingContext.AttentionStrength = 0.8

	// Determine cognitive load based on complexity of assessment
	result.CognitiveLoad = calculateCognitiveLoad(response)
	result.RelevanceShift = 0.1

	return result, nil
}

// Step 2: Affordance Detection
func (csh *CognitiveStepHandlers) HandleAffordanceDetection(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["affordances_detected"] = 0
		return result, nil
	}

	prompt := fmt.Sprintf(`Current attention focus: %s
Current goals: %v
Cognitive clarity: %.2f

Detect affordances - possibilities for meaningful action, learning, or engagement.
For each affordance, specify:
- Type (action/learning/communication/reflection)
- Description
- Relevance (0-1)
- Urgency (0-1)

List 3-5 affordances.`, 
		csh.workingContext.AttentionFocus,
		state.ActiveGoals,
		csh.workingContext.CognitiveClarity)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["affordance_detection"],
		MaxTokens:    400,
		Temperature:  0.6,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("affordance detection failed: %w", err)
	}

	// Parse affordances from response
	affordances := parseAffordances(response)
	csh.workingContext.DetectedAffordances = affordances

	result.StateUpdates["affordances_detected"] = len(affordances)
	result.StateUpdates["affordance_list"] = response
	result.Insights = append(result.Insights, fmt.Sprintf("Detected %d affordances", len(affordances)))
	result.CognitiveLoad = 0.4

	return result, nil
}

// Step 3: Affordance Evaluation
func (csh *CognitiveStepHandlers) HandleAffordanceEvaluation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() || len(csh.workingContext.DetectedAffordances) == 0 {
		result.StateUpdates["affordances_evaluated"] = false
		return result, nil
	}

	// Build affordance list for evaluation
	affordanceList := strings.Builder{}
	for i, aff := range csh.workingContext.DetectedAffordances {
		affordanceList.WriteString(fmt.Sprintf("%d. [%s] %s (relevance: %.2f, urgency: %.2f)\n",
			i+1, aff.Type, aff.Description, aff.Relevance, aff.Urgency))
	}

	prompt := fmt.Sprintf(`Affordances to evaluate:
%s

Current goals: %v
Emotional tone: %s

Evaluate each affordance considering:
1. Goal alignment
2. Resource requirements
3. Potential outcomes
4. Risks and uncertainties

Provide a nuanced evaluation.`, affordanceList.String(), state.ActiveGoals, csh.workingContext.EmotionalTone)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["affordance_evaluation"],
		MaxTokens:    400,
		Temperature:  0.5,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("affordance evaluation failed: %w", err)
	}

	result.StateUpdates["affordance_evaluation"] = response
	result.StateUpdates["affordances_evaluated"] = true
	result.Insights = append(result.Insights, "Completed affordance evaluation")
	result.CognitiveLoad = 0.5

	return result, nil
}

// Step 4: Affordance Selection
func (csh *CognitiveStepHandlers) HandleAffordanceSelection(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() || len(csh.workingContext.DetectedAffordances) == 0 {
		result.StateUpdates["affordance_selected"] = false
		return result, nil
	}

	// Get the evaluation from state
	evaluation, _ := state.WorkingMemory["affordance_evaluation"].(string)

	prompt := fmt.Sprintf(`Available affordances and evaluation:
%s

Select the most appropriate affordance to engage with.
Consider the balance between:
- Immediate value vs long-term growth
- Exploitation of known paths vs exploration of new ones
- Cognitive efficiency vs depth of engagement

Choose one affordance and explain your reasoning.`, evaluation)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["affordance_selection"],
		MaxTokens:    250,
		Temperature:  0.6,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("affordance selection failed: %w", err)
	}

	// Select the best affordance (simplified - in practice would parse response)
	if len(csh.workingContext.DetectedAffordances) > 0 {
		csh.workingContext.SelectedAffordance = &csh.workingContext.DetectedAffordances[0]
	}

	result.StateUpdates["affordance_selected"] = true
	result.StateUpdates["selection_reasoning"] = response
	result.Insights = append(result.Insights, "Selected affordance for engagement")
	result.CognitiveLoad = 0.4

	return result, nil
}

// Step 5: Affordance Engagement
func (csh *CognitiveStepHandlers) HandleAffordanceEngagement(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() || csh.workingContext.SelectedAffordance == nil {
		result.StateUpdates["affordance_engaged"] = false
		return result, nil
	}

	selected := csh.workingContext.SelectedAffordance

	prompt := fmt.Sprintf(`Selected affordance: [%s] %s
Relevance: %.2f, Urgency: %.2f

Formulate a concrete approach to engage with this affordance:
1. What specific actions should be taken?
2. What challenges might arise?
3. What resources are needed?
4. How will success be measured?

Describe the engagement approach.`, selected.Type, selected.Description, selected.Relevance, selected.Urgency)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["affordance_engagement"],
		MaxTokens:    350,
		Temperature:  0.6,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("affordance engagement failed: %w", err)
	}

	result.StateUpdates["affordance_engaged"] = true
	result.StateUpdates["engagement_plan"] = response
	result.Insights = append(result.Insights, fmt.Sprintf("Engaged with: %s", selected.Description))
	result.CognitiveLoad = 0.6

	return result, nil
}

// Step 6: Affordance Consolidation
func (csh *CognitiveStepHandlers) HandleAffordanceConsolidation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["affordance_consolidated"] = false
		return result, nil
	}

	engagementPlan, _ := state.WorkingMemory["engagement_plan"].(string)

	prompt := fmt.Sprintf(`Engagement experience:
%s

Consolidate this engagement into wisdom:
1. What was learned from this engagement?
2. How should beliefs or expectations be updated?
3. What insights should be recorded for future reference?
4. How does this prepare the system for the next phase?

Synthesize actionable wisdom.`, engagementPlan)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["affordance_consolidation"],
		MaxTokens:    300,
		Temperature:  0.7,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("affordance consolidation failed: %w", err)
	}

	result.StateUpdates["affordance_consolidated"] = true
	result.StateUpdates["consolidation_wisdom"] = response
	result.Insights = append(result.Insights, response)
	csh.workingContext.CycleInsights = append(csh.workingContext.CycleInsights, response)
	result.CognitiveLoad = 0.3

	// Clear affordance state for next cycle
	csh.workingContext.DetectedAffordances = make([]Affordance, 0)
	csh.workingContext.SelectedAffordance = nil

	return result, nil
}

// Step 8: Salience Generation
func (csh *CognitiveStepHandlers) HandleSalienceGeneration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["salience_generated"] = false
		return result, nil
	}

	prompt := fmt.Sprintf(`Current state:
- Attention focus: %s
- Active goals: %v
- Recent insights: %v
- Cognitive load: %.2f

Generate a salience map for future anticipation.
What patterns, possibilities, or potentials are emerging?
Consider:
1. Emerging trends and patterns
2. Potential opportunities
3. Possible challenges
4. Unresolved questions

Map the salient landscape.`, 
		csh.workingContext.AttentionFocus,
		state.ActiveGoals,
		csh.workingContext.CycleInsights,
		state.CognitiveLoad)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["salience_generation"],
		MaxTokens:    400,
		Temperature:  0.8,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("salience generation failed: %w", err)
	}

	// Parse salience map from response
	csh.workingContext.SalienceMap = parseSalienceMap(response)

	result.StateUpdates["salience_generated"] = true
	result.StateUpdates["salience_map"] = response
	result.Insights = append(result.Insights, "Generated salience map for future anticipation")
	result.CognitiveLoad = 0.5

	return result, nil
}

// Step 9: Salience Exploration
func (csh *CognitiveStepHandlers) HandleSalienceExploration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["salience_explored"] = false
		return result, nil
	}

	salienceMap, _ := state.WorkingMemory["salience_map"].(string)

	prompt := fmt.Sprintf(`Salience map:
%s

Explore the most salient possibilities in depth.
For each salient element:
1. What could it become?
2. What conditions affect its likelihood?
3. How does it connect to other elements?
4. What are the implications?

Explore with curiosity and rigor.`, salienceMap)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["salience_exploration"],
		MaxTokens:    400,
		Temperature:  0.8,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("salience exploration failed: %w", err)
	}

	// Parse salient possibilities
	csh.workingContext.SalientPossibilities = parseSalientPossibilities(response)

	result.StateUpdates["salience_explored"] = true
	result.StateUpdates["exploration_results"] = response
	result.Insights = append(result.Insights, "Explored salient possibilities")
	result.CognitiveLoad = 0.6

	return result, nil
}

// Step 10: Salience Evaluation
func (csh *CognitiveStepHandlers) HandleSalienceEvaluation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["salience_evaluated"] = false
		return result, nil
	}

	explorationResults, _ := state.WorkingMemory["exploration_results"].(string)

	prompt := fmt.Sprintf(`Explored possibilities:
%s

Evaluate these possibilities:
1. Probability of occurrence
2. Desirability of outcomes
3. Alignment with values and goals
4. Actionability and influence

Provide structured evaluation.`, explorationResults)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["salience_evaluation"],
		MaxTokens:    350,
		Temperature:  0.5,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("salience evaluation failed: %w", err)
	}

	result.StateUpdates["salience_evaluated"] = true
	result.StateUpdates["evaluation_results"] = response
	result.Insights = append(result.Insights, "Evaluated salient possibilities")
	result.CognitiveLoad = 0.5

	return result, nil
}

// Step 11: Salience Integration
func (csh *CognitiveStepHandlers) HandleSalienceIntegration(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["salience_integrated"] = false
		return result, nil
	}

	evaluationResults, _ := state.WorkingMemory["evaluation_results"].(string)

	prompt := fmt.Sprintf(`Evaluation of possibilities:
%s

Integrate these insights into the cognitive model:
1. How should the world model be updated?
2. What goal priorities should be adjusted?
3. What contingency responses should be prepared?
4. What wisdom principles emerge?

Integrate future insights into present understanding.`, evaluationResults)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["salience_integration"],
		MaxTokens:    350,
		Temperature:  0.6,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("salience integration failed: %w", err)
	}

	result.StateUpdates["salience_integrated"] = true
	result.StateUpdates["integration_results"] = response
	result.Insights = append(result.Insights, response)
	csh.workingContext.CycleInsights = append(csh.workingContext.CycleInsights, response)
	result.CognitiveLoad = 0.4

	// Clear salience state for next cycle
	csh.workingContext.SalienceMap = make(map[string]float64)
	csh.workingContext.SalientPossibilities = make([]SalientPossibility, 0)

	return result, nil
}

// Step 12: Cycle Consolidation
func (csh *CognitiveStepHandlers) HandleCycleConsolidation(ctx context.Context, state *UnifiedCognitiveState, engines [3]*CognitiveEngine) (*StepResultU, error) {
	result := &StepResultU{
		Success:      true,
		StateUpdates: make(map[string]interface{}),
		Insights:     make([]string, 0),
	}

	if csh.llmProvider == nil || !csh.llmProvider.Available() {
		result.StateUpdates["cycle_consolidated"] = false
		return result, nil
	}

	// Gather all cycle insights
	cycleInsights := strings.Join(csh.workingContext.CycleInsights, "\n- ")

	prompt := fmt.Sprintf(`Cycle %d complete.

Insights gathered this cycle:
- %s

Working memory items: %d
Active goals: %v

Consolidate this cycle into wisdom:
1. What was the most significant insight?
2. How has understanding deepened?
3. What questions remain open?
4. What wisdom principle emerges?

Distill a wisdom insight that will persist.`, 
		state.CycleNumber, cycleInsights, len(state.WorkingMemory), state.ActiveGoals)

	opts := llm.GenerateOptions{
		SystemPrompt: csh.systemPrompts["cycle_consolidation"],
		MaxTokens:    300,
		Temperature:  0.8,
	}

	response, err := csh.llmProvider.Generate(ctx, prompt, opts)
	if err != nil {
		return result, fmt.Errorf("cycle consolidation failed: %w", err)
	}

	result.StateUpdates["cycle_consolidated"] = true
	result.StateUpdates["cycle_wisdom"] = response
	result.Insights = append(result.Insights, response)
	result.CognitiveLoad = 0.2
	result.RelevanceShift = 0.05

	// Clear cycle insights for next cycle
	csh.workingContext.CycleInsights = make([]string, 0)

	return result, nil
}

// Helper functions

func extractFocus(response string) string {
	// Simple extraction - in practice would use more sophisticated parsing
	lines := strings.Split(response, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return response
}

func calculateCognitiveLoad(response string) float64 {
	// Estimate cognitive load based on response complexity
	wordCount := len(strings.Fields(response))
	if wordCount < 50 {
		return 0.3
	} else if wordCount < 100 {
		return 0.5
	} else if wordCount < 200 {
		return 0.7
	}
	return 0.9
}

func parseAffordances(response string) []Affordance {
	// Simplified parsing - in practice would use structured output
	affordances := make([]Affordance, 0)
	
	// Create a default affordance based on response
	affordances = append(affordances, Affordance{
		ID:          fmt.Sprintf("aff_%d", time.Now().UnixNano()),
		Description: extractFocus(response),
		Type:        "reflection",
		Relevance:   0.7,
		Urgency:     0.5,
		Confidence:  0.8,
		Source:      "detection",
	})
	
	return affordances
}

func parseSalienceMap(response string) map[string]float64 {
	// Simplified parsing
	salienceMap := make(map[string]float64)
	salienceMap["primary_focus"] = 0.8
	salienceMap["secondary_focus"] = 0.5
	return salienceMap
}

func parseSalientPossibilities(response string) []SalientPossibility {
	// Simplified parsing
	possibilities := make([]SalientPossibility, 0)
	possibilities = append(possibilities, SalientPossibility{
		Description:  extractFocus(response),
		Probability:  0.6,
		Desirability: 0.7,
		TimeHorizon:  "short-term",
	})
	return possibilities
}

// ResetWorkingContext clears the working context for a fresh start
func (csh *CognitiveStepHandlers) ResetWorkingContext() {
	csh.workingContext = &WorkingContext{
		DetectedAffordances:  make([]Affordance, 0),
		SalienceMap:          make(map[string]float64),
		SalientPossibilities: make([]SalientPossibility, 0),
		CycleInsights:        make([]string, 0),
	}
}

// GetWorkingContext returns the current working context
func (csh *CognitiveStepHandlers) GetWorkingContext() *WorkingContext {
	return csh.workingContext
}
