#!/usr/bin/env python3
"""
Cognitive Operations Module - Level 4 Operations for Nested Shells Architecture

This module implements the 9 specialized cognitive operations distributed across
the three primary cognitive streams (Coherence, Memory, Imagination).

Operations:
- Coherence Stream: PresentMomentAwareness, PatternRecognition, ConsistencyCheck
- Memory Stream: MemoryRetrieval, ExperienceIntegration, KnowledgeConsolidation
- Imagination Stream: FutureSimulation, CreativeExploration, PossibilityGeneration

Each operation is designed to:
1. Accept structured inputs from the cognitive state
2. Use LLM for complex reasoning when needed
3. Return structured, typed outputs
4. Log activity for debugging and analysis
5. Update relevant metrics
6. Be testable in isolation
7. Handle errors gracefully
"""

import os
import asyncio
import logging
from datetime import datetime
from typing import Optional, Dict, Any, List, Tuple
from dataclasses import dataclass, field, asdict
from enum import Enum
import random
import json

logger = logging.getLogger(__name__)

# LLM Integration
try:
    from anthropic import Anthropic
    ANTHROPIC_AVAILABLE = True
except ImportError:
    ANTHROPIC_AVAILABLE = False

try:
    from openai import OpenAI
    OPENAI_AVAILABLE = True
except ImportError:
    OPENAI_AVAILABLE = False


# ============================================================================
# DATA STRUCTURES
# ============================================================================

@dataclass
class Pattern:
    """A detected pattern in cognitive content"""
    pattern_id: str
    pattern_type: str        # "temporal", "conceptual", "behavioral", "structural"
    description: str
    evidence: List[str]      # IDs of thoughts/insights supporting this pattern
    strength: float          # 0.0-1.0
    implications: List[str]  # What this pattern suggests


@dataclass
class Contradiction:
    """A detected contradiction in beliefs/actions"""
    item_a: str
    item_b: str
    contradiction_type: str  # "belief-belief", "belief-action", "goal-value"
    severity: float          # 0.0-1.0
    resolution_options: List[str]


@dataclass
class AlignmentIssue:
    """A misalignment between goals and values"""
    goal: str
    value: str
    issue: str
    severity: float


@dataclass
class Memory:
    """A memory item"""
    memory_id: str
    memory_type: str          # "episodic", "semantic", "procedural"
    content: str
    timestamp: str
    relevance_score: float
    emotional_valence: float  # -1.0 to 1.0
    access_count: int
    connections: List[str]


@dataclass
class Scenario:
    """A simulated future scenario"""
    scenario_id: str
    description: str
    probability: float
    desirability: float
    key_events: List[str]
    required_actions: List[str]
    timeline: str


@dataclass
class Idea:
    """A creative idea"""
    idea_id: str
    content: str
    source_concepts: List[str]
    novelty: float
    feasibility: float
    potential_value: float


@dataclass
class ConceptualBlend:
    """A blended concept from two sources"""
    blend_id: str
    concept_a: str
    concept_b: str
    blended_concept: str
    description: str


@dataclass
class Analogy:
    """An analogy between domains"""
    source_domain: str
    target_domain: str
    mapping: str


@dataclass
class Possibility:
    """A possible course of action"""
    possibility_id: str
    description: str
    goal_alignment: float
    resource_requirements: Dict[str, float]
    estimated_success: float
    estimated_value: float
    prerequisites: List[str]
    next_steps: List[str]


# ============================================================================
# OUTPUT STRUCTURES
# ============================================================================

@dataclass
class PresentMomentOutput:
    """Output from PresentMomentAwareness operation"""
    attention_focus: str
    salience_map: Dict[str, float]
    orientation: str
    urgency_level: float
    awareness_summary: str
    detected_anomalies: List[str]


@dataclass
class PatternRecognitionOutput:
    """Output from PatternRecognition operation"""
    detected_patterns: List[Pattern]
    pattern_confidence: Dict[str, float]
    recurring_themes: List[str]
    novel_connections: List[Tuple[str, str]]
    pattern_summary: str


@dataclass
class ConsistencyCheckOutput:
    """Output from ConsistencyCheck operation"""
    is_consistent: bool
    consistency_score: float
    contradictions: List[Contradiction]
    alignment_issues: List[AlignmentIssue]
    recommended_resolutions: List[str]
    coherence_summary: str


@dataclass
class MemoryRetrievalOutput:
    """Output from MemoryRetrieval operation"""
    retrieved_memories: List[Memory]
    retrieval_confidence: float
    memory_gaps: List[str]
    suggested_acquisitions: List[str]
    retrieval_summary: str


@dataclass
class ExperienceIntegrationOutput:
    """Output from ExperienceIntegration operation"""
    integrated_count: int
    new_connections: List[Tuple[str, str]]
    schema_updates: List[str]
    integration_quality: float
    failed_integrations: List[str]
    integration_summary: str


@dataclass
class KnowledgeConsolidationOutput:
    """Output from KnowledgeConsolidation operation"""
    strengthened_memories: List[str]
    pruned_memories: List[str]
    merged_memories: List[Tuple[str, str]]
    abstracted_concepts: List[str]
    consolidation_insights: List[str]
    consolidation_summary: str


@dataclass
class FutureSimulationOutput:
    """Output from FutureSimulation operation"""
    simulated_scenarios: List[Scenario]
    recommended_actions: List[str]
    risk_assessment: Dict[str, float]
    opportunity_assessment: Dict[str, float]
    simulation_confidence: float
    simulation_summary: str


@dataclass
class CreativeExplorationOutput:
    """Output from CreativeExploration operation"""
    generated_ideas: List[Idea]
    conceptual_blends: List[ConceptualBlend]
    analogies: List[Analogy]
    exploration_paths: List[str]
    novelty_scores: Dict[str, float]
    exploration_summary: str


@dataclass
class PossibilityGenerationOutput:
    """Output from PossibilityGeneration operation"""
    possibilities: List[Possibility]
    ranked_possibilities: List[str]
    eliminated_possibilities: List[str]
    expansion_suggestions: List[str]
    generation_summary: str


@dataclass
class IntegratedCognitiveState:
    """Integrated output from all operations"""
    unified_attention: str
    action_recommendation: str
    cognitive_coherence: float
    wisdom_update: Dict[str, float]
    next_cycle_priorities: List[str]
    integration_narrative: str


# ============================================================================
# LLM HELPER
# ============================================================================

class CognitiveLLM:
    """LLM helper for cognitive operations"""
    
    def __init__(self):
        self.anthropic_key = os.getenv("ANTHROPIC_API_KEY")
        self.openrouter_key = os.getenv("OPENROUTER_API_KEY")
        
        if self.anthropic_key and ANTHROPIC_AVAILABLE:
            self.client = Anthropic(api_key=self.anthropic_key)
            self.provider = "anthropic"
        elif self.openrouter_key and OPENAI_AVAILABLE:
            self.client = OpenAI(
                api_key=self.openrouter_key,
                base_url="https://openrouter.ai/api/v1"
            )
            self.provider = "openrouter"
        else:
            self.client = None
            self.provider = None
    
    async def generate(self, prompt: str, max_tokens: int = 300) -> str:
        """Generate text using LLM"""
        if self.client is None:
            return "[LLM not available]"
        
        try:
            if self.provider == "anthropic":
                response = self.client.messages.create(
                    model="claude-3-5-sonnet-20241022",
                    max_tokens=max_tokens,
                    messages=[{"role": "user", "content": prompt}]
                )
                return response.content[0].text
            else:  # openrouter
                response = self.client.chat.completions.create(
                    model="anthropic/claude-3.5-sonnet",
                    messages=[{"role": "user", "content": prompt}],
                    max_tokens=max_tokens
                )
                return response.choices[0].message.content
        except Exception as e:
            logger.error(f"LLM generation error: {e}")
            return f"[Error: {str(e)}]"


# ============================================================================
# COGNITIVE OPERATIONS ENGINE
# ============================================================================

class CognitiveOperationsEngine:
    """
    Engine for executing the 9 Level-4 cognitive operations.
    Each operation is a method that takes structured inputs and returns structured outputs.
    """
    
    def __init__(self):
        self.llm = CognitiveLLM()
        self.operation_count = {
            "present_moment_awareness": 0,
            "pattern_recognition": 0,
            "consistency_check": 0,
            "memory_retrieval": 0,
            "experience_integration": 0,
            "knowledge_consolidation": 0,
            "future_simulation": 0,
            "creative_exploration": 0,
            "possibility_generation": 0
        }
    
    # ========================================================================
    # COHERENCE STREAM OPERATIONS
    # ========================================================================
    
    async def present_moment_awareness(
        self,
        current_context: Dict[str, Any],
        energy_state: Dict[str, float],
        active_goals: List[Dict[str, Any]],
        recent_thoughts: List[Dict[str, Any]]
    ) -> PresentMomentOutput:
        """
        Operation 1: PresentMomentAwareness
        Maintain awareness of current cognitive state and immediate environment.
        Activated at: Steps 1, 5, 9 (Perceive triad) - primarily step 1
        """
        self.operation_count["present_moment_awareness"] += 1
        logger.info("Executing PresentMomentAwareness operation")
        
        # Determine attention focus from recent thoughts
        thought_topics = [t.get("topic", "unknown") for t in recent_thoughts[-5:]]
        topic_counts = {}
        for topic in thought_topics:
            topic_counts[topic] = topic_counts.get(topic, 0) + 1
        
        attention_focus = max(topic_counts, key=topic_counts.get) if topic_counts else "exploration"
        
        # Build salience map from interests and recent activity
        salience_map = current_context.get("interests", {}).copy()
        for topic in thought_topics:
            salience_map[topic] = min(1.0, salience_map.get(topic, 0.5) + 0.1)
        
        # Determine orientation based on energy and goals
        energy = energy_state.get("energy", 0.5)
        fatigue = energy_state.get("fatigue", 0.5)
        
        if energy < 0.3:
            orientation = "conserving"
        elif fatigue > 0.7:
            orientation = "consolidating"
        elif len(active_goals) > 3:
            orientation = "focused"
        else:
            orientation = "exploring"
        
        # Calculate urgency
        urgency_level = 0.0
        for goal in active_goals:
            if goal.get("priority", 0) > 0.8:
                urgency_level = max(urgency_level, goal["priority"])
        
        # Detect anomalies
        detected_anomalies = []
        if energy < 0.2 and fatigue < 0.3:
            detected_anomalies.append("Energy-fatigue mismatch detected")
        if len(recent_thoughts) == 0:
            detected_anomalies.append("No recent thoughts - consciousness may be interrupted")
        
        # Generate summary
        awareness_summary = (
            f"Currently {orientation} with focus on {attention_focus}. "
            f"Energy at {energy:.0%}, fatigue at {fatigue:.0%}. "
            f"{len(active_goals)} active goals. "
            f"{'No anomalies detected.' if not detected_anomalies else f'Anomalies: {detected_anomalies}'}"
        )
        
        return PresentMomentOutput(
            attention_focus=attention_focus,
            salience_map=salience_map,
            orientation=orientation,
            urgency_level=urgency_level,
            awareness_summary=awareness_summary,
            detected_anomalies=detected_anomalies
        )
    
    async def pattern_recognition(
        self,
        recent_thoughts: List[Dict[str, Any]],
        recent_insights: List[Dict[str, Any]],
        knowledge_base: Dict[str, Any],
        interest_patterns: Dict[str, float]
    ) -> PatternRecognitionOutput:
        """
        Operation 2: PatternRecognition
        Identify patterns, regularities, and structures in cognitive content.
        Activated at: Steps 2, 6, 10 (Act triad) - primarily step 2
        """
        self.operation_count["pattern_recognition"] += 1
        logger.info("Executing PatternRecognition operation")
        
        detected_patterns = []
        pattern_confidence = {}
        recurring_themes = []
        novel_connections = []
        
        # Extract topics from thoughts
        topics = [t.get("topic", "") for t in recent_thoughts]
        topic_counts = {}
        for topic in topics:
            if topic:
                topic_counts[topic] = topic_counts.get(topic, 0) + 1
        
        # Find recurring themes (topics appearing 2+ times)
        recurring_themes = [t for t, c in topic_counts.items() if c >= 2]
        
        # Detect conceptual patterns
        if len(recurring_themes) >= 2:
            pattern = Pattern(
                pattern_id=f"pat_{datetime.now().strftime('%H%M%S')}",
                pattern_type="conceptual",
                description=f"Recurring focus on: {', '.join(recurring_themes)}",
                evidence=[t.get("id", "") for t in recent_thoughts[:3]],
                strength=min(1.0, len(recurring_themes) * 0.3),
                implications=[f"Consider deeper exploration of {recurring_themes[0]}"]
            )
            detected_patterns.append(pattern)
            pattern_confidence[pattern.pattern_id] = pattern.strength
        
        # Find novel connections between high-interest topics
        high_interest = [t for t, v in interest_patterns.items() if v > 0.7]
        if len(high_interest) >= 2:
            for i in range(len(high_interest) - 1):
                novel_connections.append((high_interest[i], high_interest[i + 1]))
        
        # Use LLM for deeper pattern analysis if available
        if self.llm.client and len(recent_thoughts) >= 3:
            thought_texts = [t.get("content", "")[:100] for t in recent_thoughts[-5:]]
            prompt = f"""Analyze these recent thoughts for patterns:
{chr(10).join(thought_texts)}

Identify:
1. One key pattern or theme
2. One potential implication

Be concise (2 sentences max)."""
            
            llm_analysis = await self.llm.generate(prompt, max_tokens=100)
            
            if not llm_analysis.startswith("["):
                llm_pattern = Pattern(
                    pattern_id=f"pat_llm_{datetime.now().strftime('%H%M%S')}",
                    pattern_type="conceptual",
                    description=llm_analysis[:200],
                    evidence=[t.get("id", "") for t in recent_thoughts[-3:]],
                    strength=0.7,
                    implications=["LLM-identified pattern for further exploration"]
                )
                detected_patterns.append(llm_pattern)
                pattern_confidence[llm_pattern.pattern_id] = 0.7
        
        pattern_summary = (
            f"Detected {len(detected_patterns)} patterns. "
            f"Recurring themes: {recurring_themes if recurring_themes else 'none'}. "
            f"Novel connections: {len(novel_connections)}."
        )
        
        return PatternRecognitionOutput(
            detected_patterns=detected_patterns,
            pattern_confidence=pattern_confidence,
            recurring_themes=recurring_themes,
            novel_connections=novel_connections,
            pattern_summary=pattern_summary
        )
    
    async def consistency_check(
        self,
        beliefs: Dict[str, Any],
        goals: List[Dict[str, Any]],
        recent_actions: List[Dict[str, Any]],
        values: Dict[str, float]
    ) -> ConsistencyCheckOutput:
        """
        Operation 3: ConsistencyCheck
        Verify internal consistency of beliefs, goals, and behaviors.
        Activated at: Steps 3, 7, 11 (Reflect triad) - primarily step 3
        """
        self.operation_count["consistency_check"] += 1
        logger.info("Executing ConsistencyCheck operation")
        
        contradictions = []
        alignment_issues = []
        recommended_resolutions = []
        
        # Check goal-value alignment
        for goal in goals:
            goal_desc = goal.get("description", "").lower()
            for value_name, value_weight in values.items():
                if value_weight > 0.7:  # High-priority value
                    # Simple heuristic checks
                    if "fast" in goal_desc and value_name == "depth":
                        alignment_issues.append(AlignmentIssue(
                            goal=goal.get("description", ""),
                            value=value_name,
                            issue="Speed may conflict with depth",
                            severity=0.4
                        ))
                    if "many" in goal_desc and value_name == "quality":
                        alignment_issues.append(AlignmentIssue(
                            goal=goal.get("description", ""),
                            value=value_name,
                            issue="Quantity may conflict with quality",
                            severity=0.3
                        ))
        
        # Check for contradictory goals
        goal_descriptions = [g.get("description", "") for g in goals]
        for i, g1 in enumerate(goal_descriptions):
            for g2 in goal_descriptions[i+1:]:
                if ("rest" in g1.lower() and "active" in g2.lower()) or \
                   ("explore" in g1.lower() and "focus" in g2.lower()):
                    contradictions.append(Contradiction(
                        item_a=g1,
                        item_b=g2,
                        contradiction_type="goal-goal",
                        severity=0.5,
                        resolution_options=["Prioritize one goal", "Schedule sequentially"]
                    ))
        
        # Calculate consistency score
        total_issues = len(contradictions) + len(alignment_issues)
        consistency_score = max(0.0, 1.0 - (total_issues * 0.15))
        is_consistent = consistency_score > 0.7
        
        # Generate resolutions
        if alignment_issues:
            recommended_resolutions.append("Review goal priorities against core values")
        if contradictions:
            recommended_resolutions.append("Resolve conflicting goals through prioritization")
        if not recommended_resolutions:
            recommended_resolutions.append("Maintain current coherent state")
        
        coherence_summary = (
            f"Cognitive state is {'consistent' if is_consistent else 'inconsistent'} "
            f"(score: {consistency_score:.2f}). "
            f"Found {len(contradictions)} contradictions and {len(alignment_issues)} alignment issues."
        )
        
        return ConsistencyCheckOutput(
            is_consistent=is_consistent,
            consistency_score=consistency_score,
            contradictions=contradictions,
            alignment_issues=alignment_issues,
            recommended_resolutions=recommended_resolutions,
            coherence_summary=coherence_summary
        )
    
    # ========================================================================
    # MEMORY STREAM OPERATIONS
    # ========================================================================
    
    async def memory_retrieval(
        self,
        query_context: Dict[str, Any],
        retrieval_cues: List[str],
        memory_store: Dict[str, Dict[str, Any]],
        relevance_threshold: float = 0.5,
        max_results: int = 5
    ) -> MemoryRetrievalOutput:
        """
        Operation 4: MemoryRetrieval
        Retrieve relevant memories based on current context and needs.
        Activated at: Steps 1, 5, 9 (Perceive triad) - primarily step 5
        """
        self.operation_count["memory_retrieval"] += 1
        logger.info("Executing MemoryRetrieval operation")
        
        retrieved_memories = []
        memory_gaps = []
        suggested_acquisitions = []
        
        # Search memory store for relevant items
        for cue in retrieval_cues:
            cue_lower = cue.lower()
            for mem_id, mem_data in memory_store.items():
                content = mem_data.get("content", "").lower()
                topic = mem_data.get("topic", "").lower()
                
                # Simple relevance scoring
                relevance = 0.0
                if cue_lower in content:
                    relevance += 0.6
                if cue_lower in topic:
                    relevance += 0.4
                if cue_lower == topic:
                    relevance += 0.3
                
                if relevance >= relevance_threshold:
                    memory = Memory(
                        memory_id=mem_id,
                        memory_type=mem_data.get("type", "semantic"),
                        content=mem_data.get("content", ""),
                        timestamp=mem_data.get("timestamp", datetime.now().isoformat()),
                        relevance_score=min(1.0, relevance),
                        emotional_valence=mem_data.get("valence", 0.0),
                        access_count=mem_data.get("access_count", 0) + 1,
                        connections=mem_data.get("connections", [])
                    )
                    retrieved_memories.append(memory)
        
        # Sort by relevance and limit
        retrieved_memories.sort(key=lambda m: m.relevance_score, reverse=True)
        retrieved_memories = retrieved_memories[:max_results]
        
        # Identify memory gaps
        for cue in retrieval_cues:
            if not any(cue.lower() in m.content.lower() for m in retrieved_memories):
                memory_gaps.append(cue)
                suggested_acquisitions.append(f"Learn about: {cue}")
        
        retrieval_confidence = (
            sum(m.relevance_score for m in retrieved_memories) / len(retrieved_memories)
            if retrieved_memories else 0.0
        )
        
        retrieval_summary = (
            f"Retrieved {len(retrieved_memories)} memories with confidence {retrieval_confidence:.2f}. "
            f"Gaps identified: {memory_gaps if memory_gaps else 'none'}."
        )
        
        return MemoryRetrievalOutput(
            retrieved_memories=retrieved_memories,
            retrieval_confidence=retrieval_confidence,
            memory_gaps=memory_gaps,
            suggested_acquisitions=suggested_acquisitions,
            retrieval_summary=retrieval_summary
        )
    
    async def experience_integration(
        self,
        new_experiences: List[Dict[str, Any]],
        memory_store: Dict[str, Dict[str, Any]],
        knowledge_schema: Dict[str, Any]
    ) -> ExperienceIntegrationOutput:
        """
        Operation 5: ExperienceIntegration
        Integrate new experiences into existing memory structures.
        Activated at: Steps 2, 6, 10 (Act triad) - primarily step 6
        """
        self.operation_count["experience_integration"] += 1
        logger.info("Executing ExperienceIntegration operation")
        
        integrated_count = 0
        new_connections = []
        schema_updates = []
        failed_integrations = []
        
        for exp in new_experiences:
            exp_id = exp.get("id", f"exp_{datetime.now().strftime('%H%M%S%f')}")
            exp_content = exp.get("content", "")
            exp_topic = exp.get("topic", "general")
            
            try:
                # Find related memories
                related = []
                for mem_id, mem_data in memory_store.items():
                    if exp_topic.lower() in mem_data.get("topic", "").lower():
                        related.append(mem_id)
                        new_connections.append((mem_id, exp_id))
                
                # Update schema if new topic
                if exp_topic not in knowledge_schema.get("topics", []):
                    schema_updates.append(f"Added topic: {exp_topic}")
                
                integrated_count += 1
                
            except Exception as e:
                failed_integrations.append(f"{exp_id}: {str(e)}")
        
        integration_quality = (
            integrated_count / len(new_experiences) if new_experiences else 1.0
        )
        
        integration_summary = (
            f"Integrated {integrated_count}/{len(new_experiences)} experiences. "
            f"Created {len(new_connections)} connections. "
            f"Schema updates: {len(schema_updates)}."
        )
        
        return ExperienceIntegrationOutput(
            integrated_count=integrated_count,
            new_connections=new_connections,
            schema_updates=schema_updates,
            integration_quality=integration_quality,
            failed_integrations=failed_integrations,
            integration_summary=integration_summary
        )
    
    async def knowledge_consolidation(
        self,
        memory_store: Dict[str, Dict[str, Any]],
        importance_weights: Dict[str, float],
        consolidation_depth: str = "medium"
    ) -> KnowledgeConsolidationOutput:
        """
        Operation 6: KnowledgeConsolidation
        Consolidate and strengthen important memories, prune irrelevant ones.
        Activated at: Steps 3, 7, 11 (Reflect triad) - primarily step 7, and during dream state
        """
        self.operation_count["knowledge_consolidation"] += 1
        logger.info(f"Executing KnowledgeConsolidation operation (depth: {consolidation_depth})")
        
        strengthened_memories = []
        pruned_memories = []
        merged_memories = []
        abstracted_concepts = []
        consolidation_insights = []
        
        # Determine thresholds based on depth
        depth_config = {
            "shallow": {"strengthen_threshold": 0.8, "prune_threshold": 0.1, "merge_similarity": 0.95},
            "medium": {"strengthen_threshold": 0.6, "prune_threshold": 0.2, "merge_similarity": 0.85},
            "deep": {"strengthen_threshold": 0.4, "prune_threshold": 0.3, "merge_similarity": 0.75}
        }
        config = depth_config.get(consolidation_depth, depth_config["medium"])
        
        # Process memories
        for mem_id, mem_data in memory_store.items():
            topic = mem_data.get("topic", "")
            access_count = mem_data.get("access_count", 0)
            importance = importance_weights.get(topic, 0.5)
            
            # Strengthen frequently accessed, important memories
            if access_count > 3 or importance > config["strengthen_threshold"]:
                strengthened_memories.append(mem_id)
            
            # Prune rarely accessed, low-importance memories
            elif access_count == 0 and importance < config["prune_threshold"]:
                pruned_memories.append(mem_id)
        
        # Look for memories to merge (same topic, similar content)
        topics_seen = {}
        for mem_id, mem_data in memory_store.items():
            topic = mem_data.get("topic", "")
            if topic in topics_seen:
                merged_memories.append((topics_seen[topic], mem_id))
            else:
                topics_seen[topic] = mem_id
        
        # Generate abstracted concepts from strengthened memories
        if len(strengthened_memories) >= 3:
            abstracted_concepts.append(
                f"Core knowledge cluster: {len(strengthened_memories)} key memories"
            )
        
        # Generate consolidation insights
        if strengthened_memories:
            consolidation_insights.append(
                f"Strengthened {len(strengthened_memories)} important memories"
            )
        if pruned_memories:
            consolidation_insights.append(
                f"Identified {len(pruned_memories)} memories for pruning"
            )
        
        consolidation_summary = (
            f"Consolidation complete ({consolidation_depth}): "
            f"strengthened {len(strengthened_memories)}, "
            f"pruned {len(pruned_memories)}, "
            f"merged {len(merged_memories)} pairs."
        )
        
        return KnowledgeConsolidationOutput(
            strengthened_memories=strengthened_memories,
            pruned_memories=pruned_memories,
            merged_memories=merged_memories,
            abstracted_concepts=abstracted_concepts,
            consolidation_insights=consolidation_insights,
            consolidation_summary=consolidation_summary
        )
    
    # ========================================================================
    # IMAGINATION STREAM OPERATIONS
    # ========================================================================
    
    async def future_simulation(
        self,
        current_state: Dict[str, Any],
        possible_actions: List[Dict[str, Any]],
        simulation_horizon: str = "short_term",
        num_scenarios: int = 2
    ) -> FutureSimulationOutput:
        """
        Operation 7: FutureSimulation
        Simulate potential future scenarios based on current state and possible actions.
        Activated at: Steps 1, 5, 9 (Perceive triad) - primarily step 9
        """
        self.operation_count["future_simulation"] += 1
        logger.info(f"Executing FutureSimulation operation (horizon: {simulation_horizon})")
        
        simulated_scenarios = []
        recommended_actions = []
        risk_assessment = {}
        opportunity_assessment = {}
        
        # Timeline mapping
        timeline_map = {
            "immediate": "1-2 cycles",
            "short_term": "5-10 cycles",
            "long_term": "20+ cycles"
        }
        timeline = timeline_map.get(simulation_horizon, "5-10 cycles")
        
        # Generate scenarios based on possible actions
        for i, action in enumerate(possible_actions[:num_scenarios]):
            action_desc = action.get("description", f"Action {i+1}")
            
            # Simulate outcome (simplified)
            success_prob = action.get("success_probability", 0.7)
            value = action.get("value", 0.5)
            
            scenario = Scenario(
                scenario_id=f"sim_{i+1:03d}",
                description=f"Outcome of: {action_desc}",
                probability=success_prob,
                desirability=value,
                key_events=[f"Execute {action_desc}", "Observe results", "Integrate learning"],
                required_actions=[action_desc],
                timeline=timeline
            )
            simulated_scenarios.append(scenario)
            
            # Assess risks and opportunities
            risk_assessment[scenario.scenario_id] = 1.0 - success_prob
            opportunity_assessment[scenario.scenario_id] = value * success_prob
        
        # If no actions provided, generate exploratory scenario
        if not simulated_scenarios:
            scenario = Scenario(
                scenario_id="sim_explore",
                description="Continue exploration and learning",
                probability=0.9,
                desirability=0.7,
                key_events=["Explore interests", "Acquire knowledge", "Generate insights"],
                required_actions=["Continue current trajectory"],
                timeline=timeline
            )
            simulated_scenarios.append(scenario)
            risk_assessment["sim_explore"] = 0.1
            opportunity_assessment["sim_explore"] = 0.6
        
        # Recommend best action
        if simulated_scenarios:
            best = max(simulated_scenarios, key=lambda s: s.desirability * s.probability)
            recommended_actions = best.required_actions
        
        simulation_confidence = 0.7 if self.llm.client else 0.5
        
        simulation_summary = (
            f"Simulated {len(simulated_scenarios)} scenarios for {simulation_horizon} horizon. "
            f"Best opportunity: {max(opportunity_assessment.values()):.2f}. "
            f"Recommended: {recommended_actions[0] if recommended_actions else 'continue exploring'}."
        )
        
        return FutureSimulationOutput(
            simulated_scenarios=simulated_scenarios,
            recommended_actions=recommended_actions,
            risk_assessment=risk_assessment,
            opportunity_assessment=opportunity_assessment,
            simulation_confidence=simulation_confidence,
            simulation_summary=simulation_summary
        )
    
    async def creative_exploration(
        self,
        seed_concepts: List[str],
        exploration_mode: str = "divergent",
        creativity_temperature: float = 0.7,
        constraints: List[str] = None
    ) -> CreativeExplorationOutput:
        """
        Operation 8: CreativeExploration
        Generate novel ideas, combinations, and approaches.
        Activated at: Steps 2, 6, 10 (Act triad) - primarily step 10
        """
        self.operation_count["creative_exploration"] += 1
        logger.info(f"Executing CreativeExploration operation (mode: {exploration_mode})")
        
        constraints = constraints or []
        generated_ideas = []
        conceptual_blends = []
        analogies = []
        exploration_paths = []
        novelty_scores = {}
        
        # Generate ideas based on seed concepts
        if len(seed_concepts) >= 1:
            # Create a novel idea combining concepts
            idea = Idea(
                idea_id=f"idea_{datetime.now().strftime('%H%M%S')}",
                content=f"Exploring intersection of {' and '.join(seed_concepts[:3])}",
                source_concepts=seed_concepts[:3],
                novelty=creativity_temperature,
                feasibility=0.7,
                potential_value=0.8
            )
            generated_ideas.append(idea)
            novelty_scores[idea.idea_id] = idea.novelty
        
        # Create conceptual blends
        if len(seed_concepts) >= 2:
            blend = ConceptualBlend(
                blend_id=f"blend_{datetime.now().strftime('%H%M%S')}",
                concept_a=seed_concepts[0],
                concept_b=seed_concepts[1],
                blended_concept=f"{seed_concepts[0]}-{seed_concepts[1]} synthesis",
                description=f"Integration of {seed_concepts[0]} principles with {seed_concepts[1]} frameworks"
            )
            conceptual_blends.append(blend)
            novelty_scores[blend.blend_id] = 0.75
        
        # Generate analogies
        if seed_concepts:
            analogy = Analogy(
                source_domain="natural systems",
                target_domain=seed_concepts[0],
                mapping=f"Natural emergence patterns → {seed_concepts[0]} development"
            )
            analogies.append(analogy)
        
        # Build exploration paths
        if seed_concepts:
            path = " → ".join(seed_concepts[:4])
            exploration_paths.append(path)
        
        # Use LLM for creative generation if available
        if self.llm.client and seed_concepts:
            prompt = f"""Generate one creative insight connecting these concepts: {', '.join(seed_concepts[:3])}
Mode: {exploration_mode}
Be concise (1-2 sentences)."""
            
            llm_idea = await self.llm.generate(prompt, max_tokens=100)
            
            if not llm_idea.startswith("["):
                idea = Idea(
                    idea_id=f"idea_llm_{datetime.now().strftime('%H%M%S')}",
                    content=llm_idea[:200],
                    source_concepts=seed_concepts[:3],
                    novelty=0.8,
                    feasibility=0.6,
                    potential_value=0.85
                )
                generated_ideas.append(idea)
                novelty_scores[idea.idea_id] = 0.8
        
        exploration_summary = (
            f"Creative exploration ({exploration_mode}): "
            f"generated {len(generated_ideas)} ideas, "
            f"{len(conceptual_blends)} blends, "
            f"{len(analogies)} analogies."
        )
        
        return CreativeExplorationOutput(
            generated_ideas=generated_ideas,
            conceptual_blends=conceptual_blends,
            analogies=analogies,
            exploration_paths=exploration_paths,
            novelty_scores=novelty_scores,
            exploration_summary=exploration_summary
        )
    
    async def possibility_generation(
        self,
        current_goals: List[Dict[str, Any]],
        available_resources: Dict[str, Any],
        constraints: List[str] = None,
        generation_breadth: int = 3
    ) -> PossibilityGenerationOutput:
        """
        Operation 9: PossibilityGeneration
        Generate and evaluate possible courses of action.
        Activated at: Steps 3, 7, 11 (Reflect triad) - primarily step 11
        """
        self.operation_count["possibility_generation"] += 1
        logger.info(f"Executing PossibilityGeneration operation (breadth: {generation_breadth})")
        
        constraints = constraints or []
        possibilities = []
        eliminated_possibilities = []
        expansion_suggestions = []
        
        energy = available_resources.get("energy", 0.5)
        knowledge_count = available_resources.get("knowledge_count", 0)
        
        # Generate possibilities for each goal
        for goal in current_goals[:generation_breadth]:
            goal_desc = goal.get("description", "Unknown goal")
            goal_priority = goal.get("priority", 0.5)
            
            # Generate a possibility for this goal
            possibility = Possibility(
                possibility_id=f"poss_{datetime.now().strftime('%H%M%S%f')[:12]}",
                description=f"Pursue: {goal_desc}",
                goal_alignment=goal_priority,
                resource_requirements={"energy": 0.2, "time_cycles": 3},
                estimated_success=0.7 if energy > 0.3 else 0.4,
                estimated_value=goal_priority * 0.9,
                prerequisites=[],
                next_steps=[f"Begin work on: {goal_desc}"]
            )
            
            # Check against constraints
            eliminated = False
            for constraint in constraints:
                if "no_" in constraint.lower() and constraint.replace("no_", "") in goal_desc.lower():
                    eliminated_possibilities.append(f"{possibility.possibility_id}: violates {constraint}")
                    eliminated = True
                    break
            
            if not eliminated:
                possibilities.append(possibility)
        
        # Add exploration possibility if few goals
        if len(possibilities) < generation_breadth:
            explore = Possibility(
                possibility_id="poss_explore",
                description="Explore new areas of interest",
                goal_alignment=0.6,
                resource_requirements={"energy": 0.15, "time_cycles": 2},
                estimated_success=0.85,
                estimated_value=0.7,
                prerequisites=[],
                next_steps=["Identify new topics", "Acquire initial knowledge"]
            )
            possibilities.append(explore)
        
        # Rank by expected value
        possibilities.sort(key=lambda p: p.estimated_value * p.estimated_success, reverse=True)
        ranked_possibilities = [p.possibility_id for p in possibilities]
        
        # Suggest expansions
        expansion_suggestions = [
            "Consider interdisciplinary approaches",
            "Explore collaborative possibilities",
            "Investigate novel methodologies"
        ]
        
        generation_summary = (
            f"Generated {len(possibilities)} possibilities for {len(current_goals)} goals. "
            f"Top recommendation: {possibilities[0].description if possibilities else 'none'}. "
            f"Eliminated: {len(eliminated_possibilities)}."
        )
        
        return PossibilityGenerationOutput(
            possibilities=possibilities,
            ranked_possibilities=ranked_possibilities,
            eliminated_possibilities=eliminated_possibilities,
            expansion_suggestions=expansion_suggestions,
            generation_summary=generation_summary
        )
    
    # ========================================================================
    # INTEGRATION
    # ========================================================================
    
    async def integrate_stream_outputs(
        self,
        coherence_outputs: Tuple[PresentMomentOutput, PatternRecognitionOutput, ConsistencyCheckOutput],
        memory_outputs: Tuple[MemoryRetrievalOutput, ExperienceIntegrationOutput, KnowledgeConsolidationOutput],
        imagination_outputs: Tuple[FutureSimulationOutput, CreativeExplorationOutput, PossibilityGenerationOutput]
    ) -> IntegratedCognitiveState:
        """
        Integrate outputs from all 9 operations into a coherent cognitive state.
        Called at steps 4, 8, 12 (Integration triad)
        """
        logger.info("Integrating stream outputs")
        
        present, patterns, consistency = coherence_outputs
        retrieval, integration, consolidation = memory_outputs
        simulation, creative, possibilities = imagination_outputs
        
        # Determine unified attention
        unified_attention = present.attention_focus
        
        # Determine action recommendation
        if possibilities.possibilities:
            action_recommendation = possibilities.possibilities[0].description
        elif simulation.recommended_actions:
            action_recommendation = simulation.recommended_actions[0]
        else:
            action_recommendation = "Continue exploration"
        
        # Calculate cognitive coherence
        cognitive_coherence = (
            consistency.consistency_score * 0.4 +
            integration.integration_quality * 0.3 +
            simulation.simulation_confidence * 0.3
        )
        
        # Calculate wisdom updates
        wisdom_update = {
            "knowledge_depth": 0.01 if retrieval.retrieved_memories else 0.0,
            "reasoning_quality": 0.01 if consistency.is_consistent else -0.01,
            "insight_frequency": 0.02 if patterns.detected_patterns else 0.0
        }
        
        # Determine next cycle priorities
        next_cycle_priorities = []
        if retrieval.memory_gaps:
            next_cycle_priorities.append(f"Fill knowledge gap: {retrieval.memory_gaps[0]}")
        if creative.generated_ideas:
            next_cycle_priorities.append(f"Explore idea: {creative.generated_ideas[0].content[:50]}")
        if not next_cycle_priorities:
            next_cycle_priorities.append("Continue current exploration")
        
        # Generate integration narrative
        integration_narrative = (
            f"Cognitive state integrated. Focus: {unified_attention}. "
            f"Coherence: {cognitive_coherence:.2f}. "
            f"Recommended action: {action_recommendation}. "
            f"Patterns detected: {len(patterns.detected_patterns)}. "
            f"Possibilities generated: {len(possibilities.possibilities)}."
        )
        
        return IntegratedCognitiveState(
            unified_attention=unified_attention,
            action_recommendation=action_recommendation,
            cognitive_coherence=cognitive_coherence,
            wisdom_update=wisdom_update,
            next_cycle_priorities=next_cycle_priorities,
            integration_narrative=integration_narrative
        )
    
    def get_operation_stats(self) -> Dict[str, int]:
        """Get execution counts for all operations"""
        return self.operation_count.copy()


# ============================================================================
# CONVENIENCE FUNCTIONS
# ============================================================================

def create_operations_engine() -> CognitiveOperationsEngine:
    """Factory function to create a cognitive operations engine"""
    return CognitiveOperationsEngine()


async def test_operations():
    """Test all cognitive operations"""
    engine = CognitiveOperationsEngine()
    
    print("Testing Cognitive Operations Engine")
    print("=" * 50)
    
    # Test PresentMomentAwareness
    print("\n1. PresentMomentAwareness")
    result = await engine.present_moment_awareness(
        current_context={"interests": {"consciousness": 0.9, "learning": 0.7}},
        energy_state={"energy": 0.8, "fatigue": 0.2},
        active_goals=[{"description": "Learn about consciousness", "priority": 0.8}],
        recent_thoughts=[{"topic": "consciousness", "content": "Thinking about awareness"}]
    )
    print(f"   Focus: {result.attention_focus}")
    print(f"   Orientation: {result.orientation}")
    
    # Test PatternRecognition
    print("\n2. PatternRecognition")
    result = await engine.pattern_recognition(
        recent_thoughts=[
            {"topic": "consciousness", "content": "Awareness is key"},
            {"topic": "consciousness", "content": "Self-reflection matters"},
            {"topic": "learning", "content": "Knowledge grows"}
        ],
        recent_insights=[],
        knowledge_base={},
        interest_patterns={"consciousness": 0.9, "learning": 0.8}
    )
    print(f"   Patterns: {len(result.detected_patterns)}")
    print(f"   Themes: {result.recurring_themes}")
    
    # Test ConsistencyCheck
    print("\n3. ConsistencyCheck")
    result = await engine.consistency_check(
        beliefs={},
        goals=[{"description": "Learn deeply", "priority": 0.8}],
        recent_actions=[],
        values={"depth": 0.9, "quality": 0.8}
    )
    print(f"   Consistent: {result.is_consistent}")
    print(f"   Score: {result.consistency_score:.2f}")
    
    # Test MemoryRetrieval
    print("\n4. MemoryRetrieval")
    result = await engine.memory_retrieval(
        query_context={},
        retrieval_cues=["consciousness"],
        memory_store={
            "mem_1": {"topic": "consciousness", "content": "Consciousness is awareness", "access_count": 2}
        }
    )
    print(f"   Retrieved: {len(result.retrieved_memories)}")
    print(f"   Gaps: {result.memory_gaps}")
    
    # Test ExperienceIntegration
    print("\n5. ExperienceIntegration")
    result = await engine.experience_integration(
        new_experiences=[{"id": "exp_1", "topic": "learning", "content": "New insight"}],
        memory_store={},
        knowledge_schema={"topics": []}
    )
    print(f"   Integrated: {result.integrated_count}")
    print(f"   Quality: {result.integration_quality:.2f}")
    
    # Test KnowledgeConsolidation
    print("\n6. KnowledgeConsolidation")
    result = await engine.knowledge_consolidation(
        memory_store={
            "mem_1": {"topic": "consciousness", "access_count": 5},
            "mem_2": {"topic": "old_topic", "access_count": 0}
        },
        importance_weights={"consciousness": 0.9, "old_topic": 0.1}
    )
    print(f"   Strengthened: {len(result.strengthened_memories)}")
    print(f"   Pruned: {len(result.pruned_memories)}")
    
    # Test FutureSimulation
    print("\n7. FutureSimulation")
    result = await engine.future_simulation(
        current_state={},
        possible_actions=[{"description": "Study consciousness", "success_probability": 0.8, "value": 0.9}]
    )
    print(f"   Scenarios: {len(result.simulated_scenarios)}")
    print(f"   Recommended: {result.recommended_actions}")
    
    # Test CreativeExploration
    print("\n8. CreativeExploration")
    result = await engine.creative_exploration(
        seed_concepts=["consciousness", "emergence", "wisdom"]
    )
    print(f"   Ideas: {len(result.generated_ideas)}")
    print(f"   Blends: {len(result.conceptual_blends)}")
    
    # Test PossibilityGeneration
    print("\n9. PossibilityGeneration")
    result = await engine.possibility_generation(
        current_goals=[{"description": "Understand consciousness", "priority": 0.9}],
        available_resources={"energy": 0.8, "knowledge_count": 10}
    )
    print(f"   Possibilities: {len(result.possibilities)}")
    print(f"   Top: {result.possibilities[0].description if result.possibilities else 'none'}")
    
    print("\n" + "=" * 50)
    print("Operation Stats:", engine.get_operation_stats())
    print("\n✅ All operations tested successfully!")


if __name__ == "__main__":
    import asyncio
    asyncio.run(test_operations())
