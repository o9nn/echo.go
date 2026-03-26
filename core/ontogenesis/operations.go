package ontogenesis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Operations struct {
	mu                    sync.RWMutex
	kernel                *Kernel
	evolution             *Evolution
	stageTransitions      []*StageTransition
	maturationOperations  []*MaturationOperation
	integrationOperations []*IntegrationOperation
	synthesisOperations   []*SynthesisOperation
	operationHistory      []*OperationRecord
}

type StageTransition struct {
	ID string; FromStage string; ToStage string; Timestamp time.Time
	Trigger string; Prerequisites []string; Completed bool; Effects map[string]interface{}
}

type MaturationOperation struct {
	ID string; CapabilityName string; FromLevel float64; ToLevel float64
	Method MaturationMethod; Duration time.Duration; StartTime time.Time; Completed bool
}

type MaturationMethod string
const (
	MaturationPractice    MaturationMethod = "practice"
	MaturationReflection  MaturationMethod = "reflection"
	MaturationIntegration MaturationMethod = "integration"
	MaturationEvolution   MaturationMethod = "evolution"
)

type IntegrationOperation struct {
	ID string; KnowledgeItems []string; IntegrationType IntegrationType
	Context string; Result interface{}; Timestamp time.Time; Completed bool
}

type IntegrationType string
const (
	IntegrationSynthetic    IntegrationType = "synthetic"
	IntegrationAnalytic     IntegrationType = "analytic"
	IntegrationAnalogical   IntegrationType = "analogical"
	IntegrationHierarchical IntegrationType = "hierarchical"
)

type SynthesisOperation struct {
	ID string; InputKnowledge []string; InputExperiences []string
	SynthesisMethod SynthesisMethod; WisdomOutput *WisdomArtifact; Timestamp time.Time; Completed bool
}

type SynthesisMethod string
const (
	SynthesisReflective    SynthesisMethod = "reflective"
	SynthesisExperiential  SynthesisMethod = "experiential"
	SynthesisContemplative SynthesisMethod = "contemplative"
	SynthesisDialogical    SynthesisMethod = "dialogical"
)

type WisdomArtifact struct {
	ID string; Content string; Type WisdomType; Depth float64
	Applicability float64; Confidence float64; Sources []string; CreatedAt time.Time
}

type WisdomType string
const (
	WisdomPractical     WisdomType = "practical"
	WisdomPhilosophical WisdomType = "philosophical"
	WisdomEthical       WisdomType = "ethical"
	WisdomMetaCognitive WisdomType = "metacognitive"
)

type OperationRecord struct {
	Timestamp time.Time; OperationType string; OperationID string
	Success bool; Duration time.Duration; Details map[string]interface{}
}

func NewOperations(kernel *Kernel, evolution *Evolution) *Operations {
	return &Operations{
		kernel: kernel, evolution: evolution,
		stageTransitions: make([]*StageTransition, 0), maturationOperations: make([]*MaturationOperation, 0),
		integrationOperations: make([]*IntegrationOperation, 0), synthesisOperations: make([]*SynthesisOperation, 0),
		operationHistory: make([]*OperationRecord, 0),
	}
}

func (o *Operations) ExecuteStageTransition(ctx context.Context, fromStage, toStage, trigger string) (*StageTransition, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	startTime := time.Now()
	if !o.isValidTransition(fromStage, toStage) {
		return nil, fmt.Errorf("invalid stage transition from %s to %s", fromStage, toStage)
	}
	transition := &StageTransition{ID: fmt.Sprintf("transition_%d", time.Now().Unix()), FromStage: fromStage, ToStage: toStage, Timestamp: startTime, Trigger: trigger, Prerequisites: o.getTransitionPrerequisites(toStage), Completed: false, Effects: make(map[string]interface{})}
	effects, err := o.performStageTransition(ctx, transition)
	if err != nil { return nil, fmt.Errorf("stage transition failed: %w", err) }
	transition.Effects = effects; transition.Completed = true
	o.stageTransitions = append(o.stageTransitions, transition)
	o.recordOperation("stage_transition", transition.ID, true, time.Since(startTime), map[string]interface{}{"from_stage": fromStage, "to_stage": toStage, "trigger": trigger})
	return transition, nil
}

func (o *Operations) isValidTransition(fromStage, toStage string) bool {
	validTransitions := map[string][]string{"emergent": {"developing"}, "developing": {"maturing"}, "maturing": {"mature"}, "mature": {}}
	validNext, exists := validTransitions[fromStage]
	if !exists { return false }
	for _, valid := range validNext { if valid == toStage { return true } }
	return false
}

func (o *Operations) getTransitionPrerequisites(toStage string) []string {
	prerequisites := map[string][]string{"developing": {"basic_perception", "basic_action"}, "maturing": {"multiple_capabilities", "moderate_proficiency"}, "mature": {"advanced_capabilities", "high_proficiency", "wisdom_emergence"}}
	return prerequisites[toStage]
}

func (o *Operations) performStageTransition(ctx context.Context, transition *StageTransition) (map[string]interface{}, error) {
	effects := make(map[string]interface{})
	switch transition.ToStage {
	case "developing": effects["new_capabilities_unlocked"] = []string{"reasoning", "memory_formation"}; effects["growth_rate_multiplier"] = 1.5
	case "maturing": effects["new_capabilities_unlocked"] = []string{"metacognition", "pattern_recognition"}; effects["growth_rate_multiplier"] = 1.2
	case "mature": effects["new_capabilities_unlocked"] = []string{"wisdom_synthesis", "deep_reflection"}; effects["growth_rate_multiplier"] = 1.0
	}
	effects["transition_completed"] = true
	return effects, nil
}

func (o *Operations) ExecuteMaturationOperation(ctx context.Context, capabilityName string, method MaturationMethod, targetLevel float64) (*MaturationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	startTime := time.Now()
	maturation := &MaturationOperation{ID: fmt.Sprintf("maturation_%d", time.Now().Unix()), CapabilityName: capabilityName, FromLevel: 0.0, ToLevel: targetLevel, Method: method, StartTime: startTime, Completed: false}
	duration, err := o.performMaturation(ctx, maturation)
	if err != nil { return nil, fmt.Errorf("maturation failed: %w", err) }
	maturation.Duration = duration; maturation.Completed = true
	o.maturationOperations = append(o.maturationOperations, maturation)
	o.recordOperation("maturation", maturation.ID, true, duration, map[string]interface{}{"capability": capabilityName, "method": string(method), "target": targetLevel})
	return maturation, nil
}

func (o *Operations) performMaturation(ctx context.Context, maturation *MaturationOperation) (time.Duration, error) {
	startTime := time.Now()
	switch maturation.Method {
	case MaturationPractice: maturation.ToLevel = maturation.FromLevel + 0.1
	case MaturationReflection: maturation.ToLevel = maturation.FromLevel + 0.15
	case MaturationIntegration: maturation.ToLevel = maturation.FromLevel + 0.2
	case MaturationEvolution: maturation.ToLevel = maturation.FromLevel + 0.25
	}
	return time.Since(startTime), nil
}

func (o *Operations) ExecuteIntegrationOperation(ctx context.Context, knowledgeItems []string, integrationType IntegrationType, context string) (*IntegrationOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	startTime := time.Now()
	integration := &IntegrationOperation{ID: fmt.Sprintf("integration_%d", time.Now().Unix()), KnowledgeItems: knowledgeItems, IntegrationType: integrationType, Context: context, Timestamp: startTime, Completed: false}
	result, err := o.performIntegration(ctx, integration)
	if err != nil { return nil, fmt.Errorf("integration failed: %w", err) }
	integration.Result = result; integration.Completed = true
	o.integrationOperations = append(o.integrationOperations, integration)
	o.recordOperation("integration", integration.ID, true, time.Since(startTime), map[string]interface{}{"type": string(integrationType), "item_count": len(knowledgeItems), "context": context})
	return integration, nil
}

func (o *Operations) performIntegration(ctx context.Context, integration *IntegrationOperation) (interface{}, error) {
	return map[string]interface{}{"integrated_knowledge": fmt.Sprintf("Integrated %d items using %s method", len(integration.KnowledgeItems), integration.IntegrationType), "coherence_score": 0.8, "timestamp": time.Now()}, nil
}

func (o *Operations) ExecuteSynthesisOperation(ctx context.Context, knowledge, experiences []string, method SynthesisMethod) (*SynthesisOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	startTime := time.Now()
	synthesis := &SynthesisOperation{ID: fmt.Sprintf("synthesis_%d", time.Now().Unix()), InputKnowledge: knowledge, InputExperiences: experiences, SynthesisMethod: method, Timestamp: startTime, Completed: false}
	wisdom, err := o.performSynthesis(ctx, synthesis)
	if err != nil { return nil, fmt.Errorf("synthesis failed: %w", err) }
	synthesis.WisdomOutput = wisdom; synthesis.Completed = true
	o.synthesisOperations = append(o.synthesisOperations, synthesis)
	o.recordOperation("synthesis", synthesis.ID, true, time.Since(startTime), map[string]interface{}{"method": string(method), "knowledge_count": len(knowledge), "experience_count": len(experiences), "wisdom_depth": wisdom.Depth})
	return synthesis, nil
}

func (o *Operations) performSynthesis(ctx context.Context, synthesis *SynthesisOperation) (*WisdomArtifact, error) {
	depth := 0.5
	switch synthesis.SynthesisMethod {
	case SynthesisReflective: depth = 0.7
	case SynthesisExperiential: depth = 0.6
	case SynthesisContemplative: depth = 0.8
	case SynthesisDialogical: depth = 0.65
	}
	return &WisdomArtifact{ID: fmt.Sprintf("wisdom_%d", time.Now().Unix()), Content: fmt.Sprintf("Wisdom synthesized from %d knowledge items and %d experiences", len(synthesis.InputKnowledge), len(synthesis.InputExperiences)), Type: WisdomPractical, Depth: depth, Applicability: 0.75, Confidence: 0.7, Sources: append(synthesis.InputKnowledge, synthesis.InputExperiences...), CreatedAt: time.Now()}, nil
}

func (o *Operations) recordOperation(opType, opID string, success bool, duration time.Duration, details map[string]interface{}) {
	o.operationHistory = append(o.operationHistory, &OperationRecord{Timestamp: time.Now(), OperationType: opType, OperationID: opID, Success: success, Duration: duration, Details: details})
}

func (o *Operations) GetOperationMetrics() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()
	successCount := 0; totalDuration := time.Duration(0)
	for _, record := range o.operationHistory { if record.Success { successCount++ }; totalDuration += record.Duration }
	avgDuration := time.Duration(0); if len(o.operationHistory) > 0 { avgDuration = totalDuration / time.Duration(len(o.operationHistory)) }
	successRate := 0.0; if len(o.operationHistory) > 0 { successRate = float64(successCount) / float64(len(o.operationHistory)) }
	return map[string]interface{}{"total_operations": len(o.operationHistory), "successful_operations": successCount, "stage_transitions": len(o.stageTransitions), "maturation_operations": len(o.maturationOperations), "integration_operations": len(o.integrationOperations), "synthesis_operations": len(o.synthesisOperations), "average_operation_time": avgDuration.String(), "success_rate": successRate}
}

func (o *Operations) GetWisdomArtifacts() []*WisdomArtifact {
	o.mu.RLock()
	defer o.mu.RUnlock()
	artifacts := make([]*WisdomArtifact, 0)
	for _, synthesis := range o.synthesisOperations {
		if synthesis.Completed && synthesis.WisdomOutput != nil { artifacts = append(artifacts, synthesis.WisdomOutput) }
	}
	return artifacts
}

// =============================================================================
// CROSS-DOMAIN SYNTHESIS OPERATIONS
// =============================================================================

type CrossDomainSynthesisOperation struct {
	ID string; DomainA string; DomainB string
	InputKnowledgeA []string; InputKnowledgeB []string
	BridgingMethod BridgingMethod; WisdomOutput *WisdomArtifact
	Timestamp time.Time; Completed bool
}

type BridgingMethod string
const (
	BridgingAnalogical  BridgingMethod = "analogical"
	BridgingDialectical BridgingMethod = "dialectical"
	BridgingFractal     BridgingMethod = "fractal"
	BridgingEmergent    BridgingMethod = "emergent"
)

func (o *Operations) ExecuteCrossDomainSynthesis(ctx context.Context, domainA, domainB string, knowledgeA, knowledgeB []string, method BridgingMethod) (*CrossDomainSynthesisOperation, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	startTime := time.Now()
	crossOp := &CrossDomainSynthesisOperation{ID: fmt.Sprintf("xdomain_%d", time.Now().Unix()), DomainA: domainA, DomainB: domainB, InputKnowledgeA: knowledgeA, InputKnowledgeB: knowledgeB, BridgingMethod: method, Timestamp: startTime, Completed: false}
	depth := 0.6
	switch method {
	case BridgingAnalogical: depth = 0.7
	case BridgingDialectical: depth = 0.75
	case BridgingFractal: depth = 0.8
	case BridgingEmergent: depth = 0.85
	}
	allSources := make([]string, 0, len(knowledgeA)+len(knowledgeB))
	allSources = append(allSources, knowledgeA...)
	allSources = append(allSources, knowledgeB...)
	wisdom := &WisdomArtifact{ID: fmt.Sprintf("xwisdom_%d", time.Now().Unix()), Content: fmt.Sprintf("Cross-domain synthesis [%s x %s via %s]: bridging %d concepts", domainA, domainB, method, len(allSources)), Type: WisdomPhilosophical, Depth: depth, Applicability: 0.85, Confidence: 0.65, Sources: allSources, CreatedAt: time.Now()}
	crossOp.WisdomOutput = wisdom; crossOp.Completed = true
	o.recordOperation("cross_domain_synthesis", crossOp.ID, true, time.Since(startTime), map[string]interface{}{"domain_a": domainA, "domain_b": domainB, "method": string(method), "depth": depth})
	return crossOp, nil
}

func (o *Operations) GetSuccessRate() float64 {
	o.mu.RLock()
	defer o.mu.RUnlock()
	if len(o.operationHistory) == 0 { return 0.0 }
	successCount := 0
	for _, record := range o.operationHistory { if record.Success { successCount++ } }
	return float64(successCount) / float64(len(o.operationHistory))
}

func (o *Operations) GetOperationCount() int {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.operationHistory)
}
