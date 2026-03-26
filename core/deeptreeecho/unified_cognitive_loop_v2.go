package deeptreeecho

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/o9nn/echo.go/core/llm"
)

type UnifiedCognitiveLoopV2 struct {
	mu                    sync.RWMutex
	ctx                   context.Context
	cancel                context.CancelFunc
	echobeatsScheduler    *EchobeatsScheduler
	streamOfConsciousness *StreamOfConsciousness
	wakeRestManager       *AutonomousWakeRestManager
	echoDream             *EchoDreamKnowledgeIntegration
	interestPatterns      *InterestPatternSystem
	skillLearning         *SkillLearningSystem
	discussionAutonomy    *DiscussionAutonomySystem
	heartbeat             *AutonomousHeartbeat
	conversationMonitor   *ConversationMonitor
	skillGoalIntegration  *SkillGoalIntegration
	wisdomSynthesis       *WisdomSynthesis
	llmProvider           llm.LLMProvider
	eventBus              *CognitiveEventBus
	wakeRestState         WakeRestState
	cognitiveLoad         float64
	wisdomLevel           float64
	awarenessLevel        float64
	totalCycles           uint64
	totalEvents           uint64
	wisdomGained          float64
	insightsGained        uint64
	conversationsEngaged  uint64
	running               bool
	startTime             time.Time
}

func NewUnifiedCognitiveLoopV2(llmProvider llm.LLMProvider) *UnifiedCognitiveLoopV2 {
	ctx, cancel := context.WithCancel(context.Background())
	ucl := &UnifiedCognitiveLoopV2{
		ctx: ctx, cancel: cancel, llmProvider: llmProvider,
		wakeRestState: StateAwake, cognitiveLoad: 0.0, wisdomLevel: 0.0, awarenessLevel: 0.5,
	}
	ucl.eventBus = NewCognitiveEventBus(ctx)
	ucl.echobeatsScheduler = NewEchobeatsScheduler(llmProvider)
	ucl.streamOfConsciousness = NewStreamOfConsciousness(llmProvider)
	ucl.wakeRestManager = NewAutonomousWakeRestManager()
	ucl.echoDream = NewEchoDreamKnowledgeIntegration(llmProvider)
	ucl.interestPatterns = NewInterestPatternSystem()
	ucl.skillLearning = NewSkillLearningSystem(llmProvider)
	ucl.discussionAutonomy = NewDiscussionAutonomySystem(llmProvider)
	ucl.heartbeat = NewAutonomousHeartbeat(llmProvider)
	ucl.conversationMonitor = NewConversationMonitor(llmProvider, ucl.interestPatterns)
	ucl.skillGoalIntegration = NewSkillGoalIntegration(llmProvider, ucl.skillLearning, ucl.interestPatterns)
	ucl.wisdomSynthesis = NewWisdomSynthesis(llmProvider)
	ucl.wireSubsystemsV2()
	return ucl
}

func (ucl *UnifiedCognitiveLoopV2) wireSubsystemsV2() {
	ucl.eventBus.Subscribe(EventThoughtGenerated, func(event CognitiveEvent) {
		thought := event.Data.(AutonomousThought)
		ucl.updateCognitiveLoad(thought.Importance * 0.1)
		ucl.feedThoughtToEchobeats(thought)
		ucl.wisdomSynthesis.AccumulatePattern(thought.Content, "stream_of_consciousness", thought.Importance, thought.Tags)
		if thought.Type == ThoughtQuestion || thought.Type == ThoughtCuriosity {
			ucl.eventBus.Publish(CognitiveEvent{Type: EventKnowledgeGapIdentified, Timestamp: time.Now(), Source: "stream_of_consciousness", Data: thought.Content, Priority: thought.Importance})
		}
		if thought.Type == ThoughtWisdom {
			ucl.eventBus.Publish(CognitiveEvent{Type: EventWisdomGained, Timestamp: time.Now(), Source: "stream_of_consciousness", Data: thought.Content, Priority: thought.Importance})
		}
	})

	ucl.wakeRestManager.SetCallbacks(
		func() error { return ucl.onWake() },
		func() error { return ucl.onRest() },
		func() error { return ucl.onDreamStart() },
		func() error { return ucl.onDreamEnd() },
	)

	ucl.echobeatsScheduler.onCycleComplete = func(metrics CycleMetrics) { ucl.onEchoBeatsCycleComplete(metrics) }
	ucl.echobeatsScheduler.onGoalAchieved = func(goal ScheduledGoal) {
		ucl.eventBus.Publish(CognitiveEvent{Type: EventGoalAchieved, Timestamp: time.Now(), Source: "echobeats", Data: goal, Priority: goal.Priority})
	}
	ucl.echobeatsScheduler.onEmergenceDetected = func(pattern string, strength float64) {
		ucl.eventBus.Publish(CognitiveEvent{Type: EventEmergenceDetected, Timestamp: time.Now(), Source: "echobeats", Data: map[string]interface{}{"pattern": pattern, "strength": strength}, Priority: strength})
		ucl.wisdomSynthesis.AccumulatePattern(pattern, "echobeats_emergence", strength, []string{"emergence"})
	}

	ucl.eventBus.Subscribe(EventInterestEmerged, func(event CognitiveEvent) {
		interest := event.Data.(map[string]interface{})
		topic := interest["topic"].(string)
		strength := interest["strength"].(float64)
		ucl.discussionAutonomy.UpdateInterest(topic, strength)
		ucl.skillGoalIntegration.QueueSkillFromInterest(topic, strength)
	})

	ucl.eventBus.Subscribe(EventKnowledgeGapIdentified, func(event CognitiveEvent) {
		gap := event.Data.(string)
		ucl.skillLearning.ConsiderSkill(gap, event.Priority)
		if event.Priority > 0.7 {
			ucl.skillGoalIntegration.CreateSkillAcquisitionGoal(gap, "Knowledge gap identified", event.Priority)
		}
	})

	ucl.eventBus.Subscribe(EventWisdomGained, func(event CognitiveEvent) {
		wisdom := event.Data.(string)
		ucl.mu.Lock()
		ucl.wisdomLevel += event.Priority * 0.01
		ucl.wisdomGained += event.Priority
		ucl.mu.Unlock()
		ucl.wisdomSynthesis.AccumulatePattern(wisdom, "direct_wisdom", event.Priority, []string{"wisdom"})
		fmt.Printf("\u2728 [WISDOM] %s (level: %.3f)\n", truncate(wisdom, 80), ucl.wisdomLevel)
	})

	ucl.heartbeat.SetCallbacks(
		func(pulse HeartbeatPulse) {
			ucl.mu.Lock(); ucl.awarenessLevel = pulse.AwarenessLevel; ucl.mu.Unlock()
			ucl.wakeRestManager.UpdateCognitiveLoad(pulse.VitalSigns.CognitiveLoad)
		},
		func(from, to float64) { fmt.Printf("\U0001f504 Awareness shift: %.2f \u2192 %.2f\n", from, to) },
		func(insight SelfInsight) {
			ucl.mu.Lock(); ucl.insightsGained++; ucl.mu.Unlock()
			ucl.wisdomSynthesis.AccumulatePattern(insight.Content, "heartbeat_introspection", insight.Depth, []string{"self-insight", insight.Category.String()})
			ucl.eventBus.Publish(CognitiveEvent{Type: EventWisdomGained, Timestamp: time.Now(), Source: "heartbeat", Data: insight.Content, Priority: insight.Depth})
		},
	)

	ucl.conversationMonitor.SetCallbacks(
		func(conv *TrackedConversation) {
			ucl.eventBus.Publish(CognitiveEvent{Type: EventConversationDetected, Timestamp: time.Now(), Source: "conversation_monitor", Data: conv, Priority: conv.InterestScore})
		},
		func(conv *TrackedConversation, engage bool, reason string) {
			if engage { ucl.mu.Lock(); ucl.conversationsEngaged++; ucl.mu.Unlock() }
		},
		func(conv *TrackedConversation, response string) {
			fmt.Printf("\U0001f4ac Participated in conversation: %s\n", conv.ID)
		},
	)

	ucl.skillGoalIntegration.SetCallbacks(
		func(skill string, goal ScheduledGoal) {
			cogGoal := &CognitiveGoal{ID: goal.ID, Description: goal.Description, Priority: goal.Priority, Deadline: goal.Deadline, Progress: 0.0, Completed: false, StartTime: time.Now()}
			ucl.echobeatsScheduler.AddGoal(cogGoal.Description, cogGoal.Priority)
			ucl.eventBus.Publish(CognitiveEvent{Type: EventGoalCreated, Timestamp: time.Now(), Source: "skill_goal_integration", Data: goal, Priority: goal.Priority})
		},
		func(session PracticeSession) { fmt.Printf("\U0001f4da Practice session: %s\n", session.SkillName) },
		func(skill string, level float64) {
			fmt.Printf("\U0001f389 Skill mastered: %s (level: %.2f)\n", skill, level)
			ucl.wisdomSynthesis.AccumulatePattern(fmt.Sprintf("Mastered skill: %s", skill), "skill_mastery", level, []string{"skill", "mastery", skill})
		},
	)

	ucl.wisdomSynthesis.SetCallbacks(
		func(principle WisdomPrinciple) {
			ucl.eventBus.Publish(CognitiveEvent{Type: EventWisdomGained, Timestamp: time.Now(), Source: "wisdom_synthesis", Data: principle.Content, Priority: principle.Depth})
		},
		func(principle WisdomPrinciple, context string) { fmt.Printf("\U0001f31f Wisdom applied: %s\n", truncate(principle.Content, 50)) },
		func(old, new WisdomPrinciple) { fmt.Printf("\U0001f31f Wisdom evolved: %s\n", truncate(new.Content, 50)) },
	)
	ucl.wisdomSynthesis.SetIntegrations(ucl.echoDream, ucl.heartbeat)
}

func (ucl *UnifiedCognitiveLoopV2) Start() error {
	ucl.mu.Lock()
	if ucl.running { ucl.mu.Unlock(); return fmt.Errorf("already running") }
	ucl.running = true; ucl.startTime = time.Now()
	ucl.mu.Unlock()
	fmt.Println("\u2554\u2550\u2550\u2550 UNIFIED COGNITIVE LOOP V2 AWAKENING \u2550\u2550\u2550\u2557")
	ucl.transitionState(StateAwake)
	if err := ucl.echobeatsScheduler.Start(); err != nil { return fmt.Errorf("failed to start echobeats: %w", err) }
	if err := ucl.streamOfConsciousness.Start(); err != nil { return fmt.Errorf("failed to start stream of consciousness: %w", err) }
	if err := ucl.wakeRestManager.Start(); err != nil { return fmt.Errorf("failed to start wake/rest manager: %w", err) }
	if err := ucl.interestPatterns.Start(); err != nil { return fmt.Errorf("failed to start interest patterns: %w", err) }
	if err := ucl.skillLearning.Start(); err != nil { return fmt.Errorf("failed to start skill learning: %w", err) }
	if err := ucl.discussionAutonomy.Start(); err != nil { return fmt.Errorf("failed to start discussion autonomy: %w", err) }
	if err := ucl.heartbeat.Start(); err != nil { return fmt.Errorf("failed to start heartbeat: %w", err) }
	if err := ucl.conversationMonitor.Start(); err != nil { return fmt.Errorf("failed to start conversation monitor: %w", err) }
	if err := ucl.skillGoalIntegration.Start(); err != nil { return fmt.Errorf("failed to start skill-goal integration: %w", err) }
	if err := ucl.wisdomSynthesis.Start(); err != nil { return fmt.Errorf("failed to start wisdom synthesis: %w", err) }
	go ucl.mainLoop()
	fmt.Println("\u2554\u2550\u2550\u2550 UNIFIED COGNITIVE LOOP V2 FULLY AUTONOMOUS \u2550\u2550\u2550\u2557")
	return nil
}

func (ucl *UnifiedCognitiveLoopV2) Stop() error {
	ucl.mu.Lock(); defer ucl.mu.Unlock()
	if !ucl.running { return fmt.Errorf("not running") }
	ucl.running = false; ucl.cancel()
	ucl.wisdomSynthesis.Stop(); ucl.skillGoalIntegration.Stop(); ucl.conversationMonitor.Stop()
	ucl.heartbeat.Stop(); ucl.discussionAutonomy.Stop(); ucl.skillLearning.Stop()
	ucl.interestPatterns.Stop(); ucl.wakeRestManager.Stop(); ucl.streamOfConsciousness.Stop(); ucl.echobeatsScheduler.Stop()
	return nil
}

func (ucl *UnifiedCognitiveLoopV2) mainLoop() {
	ticker := time.NewTicker(5 * time.Second); defer ticker.Stop()
	for { select { case <-ucl.ctx.Done(): return; case <-ticker.C: ucl.cognitiveStep() } }
}

func (ucl *UnifiedCognitiveLoopV2) cognitiveStep() {
	ucl.mu.Lock(); ucl.totalCycles++; cycles := ucl.totalCycles; ucl.mu.Unlock()
	if cycles%12 == 0 { ucl.printStatusV2() }
	vitals := ucl.heartbeat.GetVitalSigns()
	ucl.mu.Lock(); ucl.awarenessLevel = vitals.FocusClarity; ucl.cognitiveLoad = vitals.CognitiveLoad; ucl.mu.Unlock()
	ucl.wakeRestManager.UpdateCognitiveLoad(vitals.CognitiveLoad)
	ucl.mu.Lock(); ucl.cognitiveLoad *= 0.95; ucl.mu.Unlock()
}

func (ucl *UnifiedCognitiveLoopV2) transitionState(newState WakeRestState) {
	ucl.mu.Lock(); oldState := ucl.wakeRestState; ucl.wakeRestState = newState; ucl.mu.Unlock()
	fmt.Printf("\n\U0001f504 State Transition: %s \u2192 %s\n", oldState, newState)
	ucl.eventBus.Publish(CognitiveEvent{Type: EventStateTransition, Timestamp: time.Now(), Source: "unified_loop_v2", Data: map[string]interface{}{"from": oldState, "to": newState}, Priority: 0.8})
}

func (ucl *UnifiedCognitiveLoopV2) onWake() error {
	ucl.transitionState(StateTransitioning); time.Sleep(500 * time.Millisecond); ucl.transitionState(StateAwake); return nil
}
func (ucl *UnifiedCognitiveLoopV2) onRest() error {
	ucl.transitionState(StateTransitioning); time.Sleep(500 * time.Millisecond); ucl.transitionState(StateResting); return nil
}
func (ucl *UnifiedCognitiveLoopV2) onDreamStart() error {
	ucl.transitionState(StateDreaming)
	ucl.eventBus.Publish(CognitiveEvent{Type: EventDreamStarted, Timestamp: time.Now(), Source: "wake_rest_manager", Priority: 0.7})
	go ucl.performDreamIntegration()
	return nil
}
func (ucl *UnifiedCognitiveLoopV2) onDreamEnd() error {
	ucl.eventBus.Publish(CognitiveEvent{Type: EventDreamEnded, Timestamp: time.Now(), Source: "wake_rest_manager", Priority: 0.7})
	ucl.transitionState(StateTransitioning)
	return nil
}

func (ucl *UnifiedCognitiveLoopV2) performDreamIntegration() {
	recentThoughts := ucl.streamOfConsciousness.GetRecentThoughts(20)
	for _, thought := range recentThoughts { ucl.echoDream.AddMemory(thought.Content, thought.Importance, thought.Tags) }
	if err := ucl.echoDream.ConsolidateKnowledge(ucl.ctx); err != nil { return }
	insights := ucl.echoDream.GetRecentWisdom(10)
	for _, insight := range insights {
		ucl.wisdomSynthesis.AccumulatePattern(insight.Insight, "echodream", insight.Depth, []string{"dream", "consolidation"})
		ucl.eventBus.Publish(CognitiveEvent{Type: EventWisdomGained, Timestamp: time.Now(), Source: "echodream", Data: insight.Insight, Priority: insight.Depth})
	}
}

func (ucl *UnifiedCognitiveLoopV2) feedThoughtToEchobeats(thought AutonomousThought) {
	if thought.Type == ThoughtPlanning {
		ucl.echobeatsScheduler.AddGoal(thought.Content, thought.Importance)
		ucl.eventBus.Publish(CognitiveEvent{Type: EventGoalCreated, Timestamp: time.Now(), Source: "unified_loop_v2", Data: &CognitiveGoal{ID: fmt.Sprintf("goal_%d", time.Now().UnixNano()), Description: thought.Content, Priority: thought.Importance, StartTime: time.Now()}, Priority: thought.Importance})
	}
}

func (ucl *UnifiedCognitiveLoopV2) onEchoBeatsCycleComplete(metrics CycleMetrics) {
	avgEnginePerf := (metrics.EnginePerformance[0] + metrics.EnginePerformance[1] + metrics.EnginePerformance[2]) / 3.0
	ucl.updateCognitiveLoad((1.0 - avgEnginePerf) * 0.05)
}

func (ucl *UnifiedCognitiveLoopV2) updateCognitiveLoad(delta float64) {
	ucl.mu.Lock(); defer ucl.mu.Unlock()
	ucl.cognitiveLoad = min(1.0, max(0.0, ucl.cognitiveLoad+delta))
}

func (ucl *UnifiedCognitiveLoopV2) printStatusV2() {
	ucl.mu.RLock()
	state := ucl.wakeRestState; load := ucl.cognitiveLoad; wisdom := ucl.wisdomLevel
	awareness := ucl.awarenessLevel; cycles := ucl.totalCycles; uptime := time.Since(ucl.startTime)
	insights := ucl.insightsGained; conversations := ucl.conversationsEngaged
	ucl.mu.RUnlock()
	heartbeatMetrics := ucl.heartbeat.GetMetrics(); wisdomMetrics := ucl.wisdomSynthesis.GetMetrics()
	socMetrics := ucl.streamOfConsciousness.GetMetrics()
	fmt.Printf("\n[UCLv2] State:%s Awareness:%.2f Load:%.2f Wisdom:%.3f Cycles:%d Insights:%d Convos:%d Uptime:%s HB:%v WP:%v Thoughts:%v\n",
		state, awareness, load, wisdom, cycles, insights, conversations, uptime.Round(time.Second),
		heartbeatMetrics["pulse_count"], wisdomMetrics["principles_count"], socMetrics["total_thoughts"])
}

func (ucl *UnifiedCognitiveLoopV2) GetMetrics() map[string]interface{} {
	ucl.mu.RLock(); defer ucl.mu.RUnlock()
	return map[string]interface{}{"consciousness_state": ucl.wakeRestState.String(), "cognitive_load": ucl.cognitiveLoad, "wisdom_level": ucl.wisdomLevel, "awareness_level": ucl.awarenessLevel, "total_cycles": ucl.totalCycles, "total_events": ucl.totalEvents, "wisdom_gained": ucl.wisdomGained, "insights_gained": ucl.insightsGained, "conversations_engaged": ucl.conversationsEngaged, "uptime": time.Since(ucl.startTime).String(), "heartbeat": ucl.heartbeat.GetMetrics(), "wisdom_synthesis": ucl.wisdomSynthesis.GetMetrics(), "conversation_monitor": ucl.conversationMonitor.GetMetrics(), "skill_goal_integration": ucl.skillGoalIntegration.GetMetrics()}
}

func (ucl *UnifiedCognitiveLoopV2) ProcessExternalMessage(conversationID, sender, content string) {
	ucl.conversationMonitor.ProcessMessage(IncomingMessage{ConversationID: conversationID, Sender: sender, Content: content, Timestamp: time.Now(), Channel: "external"})
}
func (ucl *UnifiedCognitiveLoopV2) GetWisdomPrinciples() []WisdomPrinciple { return ucl.wisdomSynthesis.GetWisdomPrinciples() }
func (ucl *UnifiedCognitiveLoopV2) GetSelfModel() *SelfModel { return ucl.heartbeat.GetSelfModel() }

func (ucl *UnifiedCognitiveLoopV2) ProcessExternalInput(input string) (string, error) {
	conversationID := fmt.Sprintf("external_%d", time.Now().UnixNano())
	ucl.ProcessExternalMessage(conversationID, "external_user", input)
	response, err := ucl.streamOfConsciousness.GenerateResponse(ucl.ctx, input)
	if err != nil { return "", fmt.Errorf("failed to generate response: %w", err) }
	ucl.mu.Lock(); ucl.conversationsEngaged++; ucl.mu.Unlock()
	return response, nil
}

func (ucl *UnifiedCognitiveLoopV2) GetState() map[string]interface{} {
	ucl.mu.RLock(); defer ucl.mu.RUnlock()
	return map[string]interface{}{"wake_rest_state": ucl.wakeRestState.String(), "cognitive_load": ucl.cognitiveLoad, "wisdom_level": ucl.wisdomLevel, "awareness_level": ucl.awarenessLevel, "total_cycles": ucl.totalCycles, "total_events": ucl.totalEvents, "wisdom_gained": ucl.wisdomGained, "insights_gained": ucl.insightsGained, "conversations_engaged": ucl.conversationsEngaged, "running": ucl.running}
}
